package note

import "fmt"

// BlockType represents the type of a content block.
type BlockType string

// Constants for all supported block types.
const (
	TypeHeading   BlockType = "heading"
	TypeParagraph BlockType = "paragraph"
	TypeList      BlockType = "list"
	TypeCode      BlockType = "code"
	TypeTable     BlockType = "table"
	TypeQuote     BlockType = "quote"
	TypeCallout   BlockType = "callout"
	TypeHr        BlockType = "hr"
)

// ContentBlock is a logical part of a note (heading, paragraph, list, etc.)
type ContentBlock struct {
	RawText  string
	Metadata BlockMetadata
}

// BlockMetadata describes the position and type of a content block.
type BlockMetadata struct {
	Filepath string                 // source file path (required for cross‑note operations)
	Path     string                 // hierarchical path, e.g. "Chapter 1 > Section 2"
	Sequence int                    // ordinal number within this Path (starting at 1)
	Index    int                    // global index in the original note (0‑based)
	Depth    int                    // nesting depth (number of ancestor headings)
	Type     BlockType              // one of the Type* constants
	Extra    map[string]interface{} `json:"extra,omitempty"` // for future extensions
}

// GenerateBlockID creates a stable ID from a block's metadata.
func GenerateBlockID(meta BlockMetadata) string {
	return fmt.Sprintf("%s|%d", meta.Path, meta.Sequence)
}

// ID returns the block's stable identifier.
func (b ContentBlock) ID() string {
	return GenerateBlockID(b.Metadata)
}

// String returns a human-readable representation of the block.
func (b ContentBlock) String() string {
	return fmt.Sprintf("[%s] %q", b.Metadata.Type, b.RawText)
}
