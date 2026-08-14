package note

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/Galdoba/grohot/internal/domain/note/frontmatter"
)

// bytesPerToken is the divisor used to approximate token count from character length.
const runesPerToken = 4

// PopulateMetadata fills in all computed fields (ID, Breadcrumb, Ancestors, etc.)
// for every segment in the tree. This method mutates the segments.
func (st *SegmentTree) PopulateMetadata(filepath string, fm *frontmatter.Frontmatter) {
	var walk func(seg *Segment)
	walk = func(seg *Segment) {
		seg.NoteFrontmatter = fm
		seg.FilePath = filepath
		seg.Ancestors = seg.getAncestors() // уже есть
		seg.ID = seg.computeID(filepath)
		seg.TokenCount = seg.computeTokenCounts()
		seg.CharCount = seg.computeCharCount()
		seg.DomType = seg.determineDomType()
		//WARN:
		//CharCount и TokenCount считаются рекурсивно по всему поддереву, а DomType – только по собственным блокам.
		// Если в родительском сегменте одни параграфы, а в дочернем – большая таблица, родитель получит тип "text", хотя при чанковании всего сегмента в нём будет таблица.
		// Это может запутать стратегии чанкования.
		//
		// Рекомендация: либо рассчитывать DomType тоже рекурсивно (с учётом детей), либо добавить отдельное поле LocalDomType.
		// Пока можно оставить как есть, но пометить комментарием, что тип отражает только непосредственное содержимое.
		for _, child := range seg.Children {
			walk(child)
		}
	}
	walk(st.Root)
}

// ToChunkText returns the segment's text prefixed by breadcrumbs and ancestor headings
// to improve contextual search during indexing.
func (seg *Segment) ToChunkText() string {
	var b strings.Builder
	if len(seg.Ancestors) > 0 || seg.Header != nil {
		// b.WriteString("Context: ")
		// b.WriteString(seg.BreadCrumbs())
		// b.WriteString("\n\n")
	} else if seg.Header == nil && seg.Parent == nil {
		// корень
		if seg.NoteFrontmatter != nil {
			if title := seg.NoteFrontmatter.Get("title"); title != nil {
				b.WriteString("Note: ")
				b.WriteString(title.Value.Scalar)
				b.WriteString("\n\n")
			}
		}
	}
	for _, block := range seg.OwnBlocks {
		b.WriteString(block.RawText)
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}

// Text возвращает склеенный текст непосредственных блоков сегмента (OwnBlocks),
// без учёта дочерних сегментов.
func (seg *Segment) Text() string {
	var b strings.Builder
	for _, block := range seg.OwnBlocks {
		b.WriteString(block.RawText)
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}

// computeID returns a hex hash of filepath + breadcrumb as a stable segment identifier.
func (seg *Segment) computeID(filepath string) string {
	var segmentMarker string
	if seg.Header != nil {
		// используем чистый текст заголовка (без #)
		segmentMarker = stripHeadingMarkers(seg.Header.RawText)
	} else {
		segmentMarker = "root"
	}
	// Собираем полный путь: filepath + предки + собственный заголовок
	parts := append(seg.Ancestors, segmentMarker)
	h := sha256.Sum256([]byte(filepath + " > " + strings.Join(parts, ">")))
	return fmt.Sprintf("%x", h[:16])
}

// // buildBreadcrumb assembles the full heading path from the root down to (and including)
// // this segment. The root segment returns an empty string.
// func (seg *Segment) buildBreadcrumb() string {
// 	if seg.Parent == nil {
// 		return ""
// 	}
// 	parts := make([]string, 0)
// 	node := seg
// 	for node.Parent != nil {
// 		if node.Header != nil {
// 			text := stripHeadingMarkers(node.Header.RawText)
// 			parts = append([]string{text}, parts...)
// 		}
// 		node = node.Parent
// 	}
// 	return strings.Join(parts, " > ")
// }

// getAncestors returns the titles of all ancestor headings (excluding the current segment).
func (seg *Segment) getAncestors() []string {
	if seg.Parent == nil {
		return nil
	}
	ancestors := []string{}
	node := seg.Parent
	for node != nil {
		if node.Header != nil {
			ancestors = append([]string{stripHeadingMarkers(node.Header.RawText)}, ancestors...)
		}
		node = node.Parent
	}
	return ancestors
}

// computeTokenCounts returns approximate token counts using the heuristic len(text)/bytesPerToken.
func (seg *Segment) computeTokenCounts() map[string]int {
	text := seg.collectFullText()
	return map[string]int{
		"char_div_4": len([]rune(text)) / runesPerToken,
	}
}

// computeCharCount returns the total character length of the segment's text.
func (seg *Segment) computeCharCount() int {
	return len([]rune(seg.collectFullText()))
}

// determineDomType analyses OwnBlocks and returns the dominant content type.
func (seg *Segment) determineDomType() string {
	hasTable, hasCode := false, false
	for _, b := range seg.OwnBlocks {
		switch b.Metadata.Type {
		case TypeTable:
			hasTable = true
		case TypeCode:
			hasCode = true
		}
	}
	if hasTable && hasCode {
		return domTypeMixed
	}
	if hasTable {
		return domTypeTable
	}
	if hasCode {
		return domTypeCode
	}
	return domTypeText
}

// collectFullText gathers the full text content of the segment (OwnBlocks and all descendants).
func (seg *Segment) collectFullText() string {
	var b strings.Builder
	for _, block := range seg.OwnBlocks {
		b.WriteString(block.RawText)
		b.WriteString("\n")
	}
	for _, child := range seg.Children {
		b.WriteString(child.collectFullText())
	}
	return b.String()
}

// Breadcrumbs returns the full hierarchical path for display and debugging,
// e.g. "/vault/Note.md > Chapter 1 > Section 2".
func (seg *Segment) BreadCrumbs() string {
	parts := []string{seg.FilePath}
	parts = append(parts, seg.Ancestors...)
	if seg.Header != nil {
		parts = append(parts, stripHeadingMarkers(seg.Header.RawText))
	}
	return strings.Join(parts, " > ")
}

// HeadingText returns the heading text without # markers.
func (seg *Segment) HeadingText() string {
	if seg.Header == nil {
		return ""
	}
	return stripHeadingMarkers(seg.Header.RawText)
}
