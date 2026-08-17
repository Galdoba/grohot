package note

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/Galdoba/grohot/internal/domain/note/frontmatter"
)

// runesPerToken is the divisor used to approximate token count from character length.
// This heuristic assumes roughly 4 characters per token for English text.
const runesPerToken = 4

// segmentIDHashBytes is the number of bytes from SHA-256 used in segment ID.
const segmentIDHashBytes = 16

// PopulateMetadata fills in all computed fields for every segment in the tree.
// This method mutates the segments.
func (st *SegmentTree) PopulateMetadata(filepath string, fm *frontmatter.Frontmatter) {
	st.walkAndPopulate(st.Root, filepath, fm)
}

// walkAndPopulate recursively traverses the segment tree and populates metadata
// for each node.
func (st *SegmentTree) walkAndPopulate(seg *Segment, filepath string, fm *frontmatter.Frontmatter) {
	if seg == nil {
		return
	}
	seg.NoteFrontmatter = fm
	seg.FilePath = filepath
	seg.Ancestors = seg.getAncestors()
	seg.ID = seg.computeID(filepath)
	seg.TokenCount = seg.computeTokenCounts()
	seg.CharCount = seg.computeCharCount()
	seg.DomType = seg.determineDomType()

	for _, child := range seg.Children {
		st.walkAndPopulate(child, filepath, fm)
	}
}

// ToChunkText returns the segment's text prefixed by breadcrumbs and ancestor headings
// to improve contextual search during indexing.
// For the root segment, the note title (if available) is used as context.
func (seg *Segment) ToChunkText() string {
	var b strings.Builder

	if seg.Parent == nil && seg.Header == nil {
		// Root segment: prepend note title from frontmatter if present.
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

// Text returns the concatenated text of the segment's immediate blocks (OwnBlocks),
// without considering child segments.
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
	segmentMarker := "root"
	if seg.Header != nil {
		// Use the clean heading text (without # markers)
		segmentMarker = stripHeadingMarkers(seg.Header.RawText)
	}
	// Build the full path: filepath + ancestors + own heading
	parts := append(seg.Ancestors, segmentMarker)
	h := sha256.Sum256([]byte(filepath + " > " + strings.Join(parts, ">")))
	return fmt.Sprintf("%x", h[:segmentIDHashBytes])
}

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

// computeTokenCounts returns approximate token counts using the heuristic len(text)/runesPerToken.
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
// Note: this considers only immediate blocks, not descendants.
// This is intentional to keep the type relevant for the segment's own content.
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
