package note

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Galdoba/grohot/internal/domain/link"
	"github.com/Galdoba/grohot/internal/domain/note/frontmatter"
)

// Constants for chunk ID and text embedding formatting.
const (
	chunkIDHashBytes    = 16 // number of bytes from SHA-256 used in chunk ID
	embeddingSeparator  = "\n\n\n"
	breadcrumbSeparator = " > "
	contentTypePrefix   = "Content Type: "
)

// markdownReplacer removes common Markdown formatting symbols.
// Immutable: safe for concurrent read access.
var markdownReplacer = strings.NewReplacer(
	"**", "",
	"*", "",
	"__", "",
	"_", "",
)

// Chunk represents a unit of content ready for embedding and indexing.
type Chunk struct {
	ID        string        // stable identifier: segmentID + ":" + hash(Text)
	Text      string        // original segment text (without breadcrumbs)
	Embedding []float64     // vector representation, filled after embedding
	Metadata  ChunkMetadata // structured metadata
}

// ChunkInput contains parameters for creating a chunk from a segment.
type ChunkInput struct {
	Text        string // text of this specific chunk
	Index       int    // ordinal number of the chunk inside the segment (0‑based)
	TotalChunks int    // total number of chunks the segment was split into
}

// ChunkMetadata holds all information needed to describe, filter, and
// reconstruct the context of a chunk.
type ChunkMetadata struct {
	// Identification and ordering
	SegmentID   string `json:"segment_id"`
	ChunkIndex  int    `json:"chunk_index"`  // chunk number within the segment (0-based)
	TotalChunks int    `json:"total_chunks"` // total number of chunks in this segment

	// Source and hierarchy
	FilePath     string   `json:"file_path"`
	Ancestors    []string `json:"ancestors"`     // e.g. ["Chapter 1"]
	HeadingText  string   `json:"heading_text"`  // heading text of the segment (without #)
	HeadingLevel int      `json:"heading_level"` // 1..6, 0 for root (number of #)

	// Content characteristics
	DomType    string `json:"dom_type"`    // "text", "table", "code", "mixed"
	TokenCount int    `json:"token_count"` // approximate number of tokens
	CharCount  int    `json:"char_count"`

	// Frontmatter properties of the note
	Tags    []string `json:"tags,omitempty"`
	Aliases []string `json:"aliases,omitempty"`
	Created string   `json:"created,omitempty"`
	Updated string   `json:"updated,omitempty"`

	// Horizontal links (will be added later)
	Links []link.Link `json:"links,omitempty"`
}

// NewChunk creates a single Chunk from a segment and input data.
// Returns an error if input parameters are invalid.
func NewChunk(seg *Segment, input ChunkInput) (*Chunk, error) {
	if err := validateChunkInput(input); err != nil {
		return nil, err
	}
	if seg == nil {
		return nil, fmt.Errorf("segment cannot be nil")
	}

	chunk := &Chunk{
		ID:   computeChunkID(seg.ID, input.Text),
		Text: input.Text,
	}
	chunk.Metadata = buildChunkMetadata(seg, input)
	return chunk, nil
}

// buildChunkMetadata constructs ChunkMetadata from segment and input.
// This is a separate function to keep NewChunk at a single level of abstraction.
func buildChunkMetadata(seg *Segment, input ChunkInput) ChunkMetadata {
	meta := ChunkMetadata{
		SegmentID:   seg.ID,
		ChunkIndex:  input.Index,
		TotalChunks: input.TotalChunks,
		FilePath:    seg.FilePath,
		Ancestors:   seg.Ancestors,
		DomType:     seg.DomType,
		CharCount:   len([]rune(input.Text)),
	}
	meta.TokenCount = determineChunkTokenCount(seg, input.Text)
	setChunkHeading(&meta, seg)
	setChunkFrontmatter(&meta, seg)
	return meta
}

// determineChunkTokenCount selects the best available token count heuristic.
// Currently falls back to len(text)/runesPerToken if no heuristic is present.
func determineChunkTokenCount(seg *Segment, chunkText string) int {
	if count, ok := seg.TokenCount["char_div_4"]; ok {
		return count
	}
	// Fallback: compute from chunk text directly
	return len([]rune(chunkText)) / runesPerToken
}

// setChunkHeading fills heading-related metadata from the segment's header.
func setChunkHeading(meta *ChunkMetadata, seg *Segment) {
	if seg.Header == nil {
		return
	}
	meta.HeadingText = stripHeadingMarkers(seg.Header.RawText)
	meta.HeadingLevel = headingLevel(seg.Header.RawText)
}

// setChunkFrontmatter extracts relevant frontmatter properties.
// It assumes seg.NoteFrontmatter is already populated by the segment tree.
func setChunkFrontmatter(meta *ChunkMetadata, seg *Segment) {
	if seg.NoteFrontmatter == nil {
		return
	}
	meta.Tags = extractStringList(seg.NoteFrontmatter, "tags")
	meta.Aliases = extractStringList(seg.NoteFrontmatter, "aliases")
	meta.Created = extractScalar(seg.NoteFrontmatter, "created")
	meta.Updated = extractScalar(seg.NoteFrontmatter, "updated")
}

// validateChunkInput validates input parameters for chunk creation.
func validateChunkInput(input ChunkInput) error {
	if strings.TrimSpace(input.Text) == "" {
		return fmt.Errorf("chunk text cannot be empty")
	}
	if input.TotalChunks <= 0 {
		return fmt.Errorf("total chunks must be positive")
	}
	if input.Index < 0 || input.Index >= input.TotalChunks {
		return fmt.Errorf("chunk index %d out of range [0, %d)", input.Index, input.TotalChunks)
	}
	return nil
}

// computeChunkID returns a stable identifier for the chunk.
// Format: <segmentID>:<hex hash of first chunkIDHashBytes bytes of SHA-256 of chunk text>.
func computeChunkID(segmentID, chunkText string) string {
	h := sha256.Sum256([]byte(chunkText))
	return fmt.Sprintf("%s:%x", segmentID, h[:chunkIDHashBytes])
}

// TextToEmbed returns the chunk text prefixed with contextual information
// (file path, ancestors, current heading) to improve embedding quality.
func (c *Chunk) TextToEmbed() string {
	var lines []string

	if c.Metadata.FilePath != "" {
		lines = append(lines, c.Metadata.FilePath)
	}
	if len(c.Metadata.Ancestors) > 0 || c.Metadata.HeadingText != "" {
		parts := append(c.Metadata.Ancestors, c.Metadata.HeadingText)
		lines = append(lines, strings.Join(parts, breadcrumbSeparator))
	}
	if c.Metadata.DomType != "" && c.Metadata.DomType != domTypeText {
		lines = append(lines, contentTypePrefix+c.Metadata.DomType)
	}

	if len(lines) == 0 {
		return c.Text
	}
	return strings.Join(lines, "\n") + embeddingSeparator + c.Text
}

// TextToLLM returns the chunk text with Markdown formatting removed.
// This is suitable for direct inclusion in an LLM prompt.
func (c *Chunk) TextToLLM() string {
	return markdownReplacer.Replace(c.Text)
}

// IsContentChanged compares the current chunk ID with a previously saved ID.
// Returns true if the chunk text has changed (ID includes hash of text).
func (c *Chunk) IsContentChanged(oldChunkID string) bool {
	return c.ID != oldChunkID
}

// TokenEstimate returns the approximate token count from metadata.
func (c *Chunk) TokenEstimate() int {
	return c.Metadata.TokenCount
}

// SetLinks sets horizontal links in the chunk metadata.
func (c *Chunk) SetLinks(links []link.Link) {
	c.Metadata.Links = links
}

// MetadataJSON returns the JSON representation of the chunk metadata.
func (c *Chunk) MetadataJSON() ([]byte, error) {
	return json.Marshal(c.Metadata)
}

// --- helper functions for frontmatter property extraction ---

// extractStringList extracts a list of strings from the property with the given name.
// Returns nil if the property is missing or is not a list.
func extractStringList(fm *frontmatter.Frontmatter, name string) []string {
	if fm == nil {
		return nil
	}
	p := fm.Get(name)
	if p == nil || p.Value.List == nil {
		return nil
	}
	return p.Value.List
}

// extractScalar extracts a scalar value from the property with the given name.
// Returns an empty string if the property is missing or is not a scalar.
func extractScalar(fm *frontmatter.Frontmatter, name string) string {
	if fm == nil {
		return ""
	}
	p := fm.Get(name)
	if p == nil || p.Value.List != nil {
		return ""
	}
	return p.Value.Scalar
}
