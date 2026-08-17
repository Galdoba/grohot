package note

import (
	"fmt"
	"strings"

	"github.com/Galdoba/grohot/internal/domain/note/frontmatter/property"
)

// Display limits for property value formatting.
const (
	maxPropertyListItems    = 3  // maximum number of list items shown in String()
	maxPropertyScalarLength = 40 // maximum length of scalar value shown in String()
)

// String returns detailed information about the segment and its contents.
// This method is intended for debugging and diagnostics.
func (seg *Segment) String() string {
	b := &strings.Builder{}
	b.WriteString("═══ Segment ═══\n")
	seg.writeIDAndHierarchy(b)
	seg.writeParentAndChildren(b)
	seg.writeHeader(b)
	seg.writeOwnBlocks(b)
	seg.writeNoteFrontmatter(b)
	seg.writeProjectedFrontmatter(b)
	return b.String()
}

// writeIDAndHierarchy appends identifier, breadcrumb, ancestors, dom type,
// token counts, and character count to the builder.
func (seg *Segment) writeIDAndHierarchy(b *strings.Builder) {
	fmt.Fprintf(b, "ID:          %s\n", seg.ID)
	fmt.Fprintf(b, "FilePath:    %s\n", seg.FilePath)
	fmt.Fprintf(b, "Ancestors:   %v\n", seg.Ancestors)
	fmt.Fprintf(b, "DomType:     %s\n", seg.DomType)
	fmt.Fprintf(b, "Tokens:\n")
	for method, count := range seg.TokenCount {
		fmt.Fprintf(b, "  %s: %d\n", method, count)
	}
	fmt.Fprintf(b, "Chars:       %d\n", seg.CharCount)
	fmt.Fprintf(b, "BreadCrumbs: %s\n", seg.BreadCrumbs())
}

// writeParentAndChildren appends parent ID and children IDs.
func (seg *Segment) writeParentAndChildren(b *strings.Builder) {
	if seg.Parent != nil {
		fmt.Fprintf(b, "Parent ID:  %s\n", seg.Parent.ID)
	} else {
		fmt.Fprintf(b, "Parent ID:  <root>\n")
	}
	if len(seg.Children) > 0 {
		childrenIDs := make([]string, len(seg.Children))
		for i, c := range seg.Children {
			childrenIDs[i] = c.ID
		}
		fmt.Fprintf(b, "Children:   [%s]\n", strings.Join(childrenIDs, ", "))
	} else {
		fmt.Fprintf(b, "Children:   none\n")
	}
}

// writeHeader appends the heading line if present.
func (seg *Segment) writeHeader(b *strings.Builder) {
	if seg.Header != nil {
		fmt.Fprintf(b, "Header:     %q\n", seg.Header.RawText)
	} else {
		fmt.Fprintf(b, "Header:     <none> (root segment)\n")
	}
}

// writeOwnBlocks appends a preview of each OwnBlock.
func (seg *Segment) writeOwnBlocks(b *strings.Builder) {
	if len(seg.OwnBlocks) > 0 {
		b.WriteString("OwnBlocks:\n")
		for i, block := range seg.OwnBlocks {
			textPreview := block.RawText
			if len(textPreview) > ownBlockDisplayWidth {
				textPreview = textPreview[:ownBlockDisplayWidth] + "..."
			}
			fmt.Fprintf(b, "  %d) [%s] %q\n", i, block.Metadata.Type, textPreview)
		}
	} else {
		b.WriteString("OwnBlocks:  none\n")
	}
}

// writeNoteFrontmatter appends a summary of the note's frontmatter properties.
func (seg *Segment) writeNoteFrontmatter(b *strings.Builder) {
	if seg.NoteFrontmatter != nil {
		b.WriteString("Note props:\n")
		for _, p := range seg.NoteFrontmatter.Properties {
			valStr := formatPropertyValue(p)
			fmt.Fprintf(b, "  %s: %s\n", p.Name, valStr)
		}
	} else {
		b.WriteString("Note props: none\n")
	}
}

// writeProjectedFrontmatter appends whether a projected frontmatter is present.
func (seg *Segment) writeProjectedFrontmatter(b *strings.Builder) {
	if seg.ProjectedFrontmatter != nil {
		b.WriteString("Projected Frontmatter: present\n")
	} else {
		b.WriteString("Projected Frontmatter: none\n")
	}
}

// formatPropertyValue returns a shortened string representation of a property value.
func formatPropertyValue(p *property.Property) string {
	if p.Value.List != nil {
		items := p.Value.List
		if len(items) > maxPropertyListItems {
			items = items[:maxPropertyListItems]
			return fmt.Sprintf("[%s, ...] (%d total)", strings.Join(items, ", "), len(p.Value.List))
		}
		return fmt.Sprintf("[%s]", strings.Join(items, ", "))
	}
	val := p.Value.Scalar
	if len(val) > maxPropertyScalarLength {
		val = val[:maxPropertyScalarLength] + "..."
	}
	return fmt.Sprintf("%q", val)
}
