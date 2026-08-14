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
	OwnBlocks            []ContentBlock           // non‑heading blocks directly inside this section
	Parent               *Segment                 // parent section (nil for root)
	Children             []*Segment               // immediately nested subsections
	ProjectedFrontmatter *frontmatter.Frontmatter // reserved for future cross‑note extraction

	//fullpath to note
	FilePath string
	// Stable identifier computed from filepath and breadcrumb.
	ID string
	// Titles of all ancestor headings (without # markers), from root to parent.
	Ancestors []string
	// Approximate token counts using different heuristics.
	TokenCount map[string]int
	// Total character length of the segment's text.
	CharCount int
	// Dominant content type: "text", "table", "code", or "mixed".
	DomType string

	// Reference to the note's frontmatter for access to tags, dates, etc.
	NoteFrontmatter *frontmatter.Frontmatter

	IncomingLinks []link.Link
	OutgoingLinks []link.Link
}

// SegmentTree is the top‑level container holding the entire hierarchical structure of a note.
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
