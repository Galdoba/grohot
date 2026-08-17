package note

import (
	"github.com/Galdoba/grohot/internal/domain/link"
	"github.com/Galdoba/grohot/internal/domain/note/frontmatter"
)

// Segment represents a section of a note, delimited by a heading.
// The root segment has no Header and contains all blocks that appear before any heading,
// as well as top-level sections.
type Segment struct {
	Header               *ContentBlock            // heading block that defines this section (nil for root)
	OwnBlocks            []ContentBlock           // non-heading blocks directly inside this section
	Parent               *Segment                 // parent section (nil for root)
	Children             []*Segment               // immediately nested subsections
	ProjectedFrontmatter *frontmatter.Frontmatter // reserved for future cross-note extraction

	FilePath   string         // full path to the source note file
	ID         string         // stable identifier computed from filepath and breadcrumb
	Ancestors  []string       // titles of all ancestor headings (without # markers), from root to parent
	TokenCount map[string]int // approximate token counts using different heuristics
	CharCount  int            // total character length of the segment's text
	DomType    string         // dominant content type: "text", "table", "code", or "mixed"

	NoteFrontmatter *frontmatter.Frontmatter // reference to the note's frontmatter for tags, dates, etc.

	IncomingLinks []link.Link // links pointing to this segment
	OutgoingLinks []link.Link // links originating from this segment
}

// SegmentTree is the top-level container holding the entire hierarchical structure of a note.
type SegmentTree struct {
	Root *Segment
}

// Constants for dominant content types returned by determineDomType.
const (
	domTypeText  = "text"
	domTypeTable = "table"
	domTypeCode  = "code"
	domTypeMixed = "mixed"
)
