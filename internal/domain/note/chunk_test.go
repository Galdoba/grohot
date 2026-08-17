// chunk_test.go
package note

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/Galdoba/grohot/internal/domain/link"
	"github.com/Galdoba/grohot/internal/domain/note/frontmatter"
	"github.com/Galdoba/grohot/internal/domain/note/frontmatter/property"
)

func TestValidateChunkInput(t *testing.T) {
	tests := []struct {
		name    string
		input   ChunkInput
		wantErr bool
	}{
		{"empty text", ChunkInput{Text: "   "}, true},
		{"zero total chunks", ChunkInput{Text: "abc", TotalChunks: 0}, true},
		{"negative total", ChunkInput{Text: "abc", TotalChunks: -1}, true},
		{"index negative", ChunkInput{Text: "abc", TotalChunks: 3, Index: -1}, true},
		{"index equal total", ChunkInput{Text: "abc", TotalChunks: 3, Index: 3}, true},
		{"valid", ChunkInput{Text: "abc", TotalChunks: 3, Index: 1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateChunkInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateChunkInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComputeChunkID(t *testing.T) {
	id1 := computeChunkID("seg1", "hello")
	id2 := computeChunkID("seg1", "hello")
	id3 := computeChunkID("seg1", "world")
	id4 := computeChunkID("seg2", "hello")

	if id1 != id2 {
		t.Errorf("same input should produce same ID, got %q and %q", id1, id2)
	}
	if id1 == id3 {
		t.Errorf("different text should produce different ID")
	}
	if id1 == id4 {
		t.Errorf("different segment ID should produce different ID")
	}
	if !strings.HasPrefix(id1, "seg1:") {
		t.Errorf("ID should start with segment ID prefix, got %q", id1)
	}
}

// helper to build a segment for chunk tests
func newTestSegmentForChunk() *Segment {
	return &Segment{
		ID:        "seg123",
		FilePath:  "/vault/note.md",
		Ancestors: []string{"Chapter 1"},
		DomType:   domTypeText,
		TokenCount: map[string]int{"char_div_4": 10},
		Header: &ContentBlock{
			RawText:  "## Section A",
			Metadata: BlockMetadata{Type: TypeHeading},
		},
		NoteFrontmatter: newTestFrontmatter(),
	}
}

func newTestFrontmatter() *frontmatter.Frontmatter {
	return &frontmatter.Frontmatter{
		Properties: []*property.Property{
			{Name: "title", Value: property.Value{Scalar: "Test Note"}},
			{Name: "tags", Value: property.Value{List: []string{"a", "b"}}},
			{Name: "aliases", Value: property.Value{List: []string{"alias1"}}},
			{Name: "created", Value: property.Value{Scalar: "2025-01-01"}},
			{Name: "updated", Value: property.Value{Scalar: "2025-01-02"}},
		},
	}
}

func TestNewChunk(t *testing.T) {
	seg := newTestSegmentForChunk()
	input := ChunkInput{Text: "Some text", Index: 0, TotalChunks: 1}

	chunk, err := NewChunk(seg, input)
	if err != nil {
		t.Fatalf("NewChunk() unexpected error: %v", err)
	}

	if chunk.ID == "" {
		t.Error("chunk ID should not be empty")
	}
	if chunk.Text != input.Text {
		t.Errorf("chunk text = %q, want %q", chunk.Text, input.Text)
	}
	if chunk.Metadata.SegmentID != seg.ID {
		t.Errorf("SegmentID = %q, want %q", chunk.Metadata.SegmentID, seg.ID)
	}
	if chunk.Metadata.ChunkIndex != input.Index {
		t.Errorf("ChunkIndex = %d, want %d", chunk.Metadata.ChunkIndex, input.Index)
	}
	if chunk.Metadata.TotalChunks != input.TotalChunks {
		t.Errorf("TotalChunks = %d, want %d", chunk.Metadata.TotalChunks, input.TotalChunks)
	}
	if chunk.Metadata.FilePath != seg.FilePath {
		t.Errorf("FilePath = %q, want %q", chunk.Metadata.FilePath, seg.FilePath)
	}
	if !reflect.DeepEqual(chunk.Metadata.Ancestors, seg.Ancestors) {
		t.Errorf("Ancestors = %v, want %v", chunk.Metadata.Ancestors, seg.Ancestors)
	}
	if chunk.Metadata.HeadingText != "Section A" {
		t.Errorf("HeadingText = %q, want %q", chunk.Metadata.HeadingText, "Section A")
	}
	if chunk.Metadata.HeadingLevel != 2 {
		t.Errorf("HeadingLevel = %d, want %d", chunk.Metadata.HeadingLevel, 2)
	}
	if chunk.Metadata.DomType != domTypeText {
		t.Errorf("DomType = %q, want %q", chunk.Metadata.DomType, domTypeText)
	}
	if chunk.Metadata.CharCount != len([]rune(input.Text)) {
		t.Errorf("CharCount = %d, want %d", chunk.Metadata.CharCount, len([]rune(input.Text)))
	}
	if chunk.Metadata.TokenCount != seg.TokenCount["char_div_4"] {
		t.Errorf("TokenCount = %d, want %d", chunk.Metadata.TokenCount, seg.TokenCount["char_div_4"])
	}
	if len(chunk.Metadata.Tags) != 2 || chunk.Metadata.Tags[0] != "a" {
		t.Errorf("Tags = %v, want [a b]", chunk.Metadata.Tags)
	}
	if len(chunk.Metadata.Aliases) != 1 || chunk.Metadata.Aliases[0] != "alias1" {
		t.Errorf("Aliases = %v, want [alias1]", chunk.Metadata.Aliases)
	}
	if chunk.Metadata.Created != "2025-01-01" {
		t.Errorf("Created = %q, want 2025-01-01", chunk.Metadata.Created)
	}
	if chunk.Metadata.Updated != "2025-01-02" {
		t.Errorf("Updated = %q, want 2025-01-02", chunk.Metadata.Updated)
	}
}

func TestNewChunkValidation(t *testing.T) {
	seg := newTestSegmentForChunk()
	_, err := NewChunk(seg, ChunkInput{Text: ""})
	if err == nil {
		t.Error("expected error for empty text")
	}
	_, err = NewChunk(nil, ChunkInput{Text: "x", TotalChunks: 1})
	if err == nil {
		t.Error("expected error for nil segment")
	}
}

func TestTextToEmbed(t *testing.T) {
	seg := newTestSegmentForChunk()
	chunk, _ := NewChunk(seg, ChunkInput{Text: "Body text", TotalChunks: 1, Index: 0})

	t.Run("with context", func(t *testing.T) {
		expected := "/vault/note.md\nChapter 1 > Section A\n\n\nBody text"
		if got := chunk.TextToEmbed(); got != expected {
			t.Errorf("TextToEmbed() = %q, want %q", got, expected)
		}
	})

	// test without context
	seg2 := &Segment{ID: "seg", DomType: domTypeText}
	chunk2, _ := NewChunk(seg2, ChunkInput{Text: "Only text", TotalChunks: 1, Index: 0})
	if got := chunk2.TextToEmbed(); got != "Only text" {
		t.Errorf("TextToEmbed() without context = %q, want %q", got, "Only text")
	}

	// test with non-text dom type
	seg3 := &Segment{ID: "seg", DomType: domTypeTable}
	chunk3, _ := NewChunk(seg3, ChunkInput{Text: "Table", TotalChunks: 1, Index: 0})
	expected := "Content Type: table\n\n\nTable"
	if got := chunk3.TextToEmbed(); got != expected {
		t.Errorf("TextToEmbed() with table = %q, want %q", got, expected)
	}
}

func TestTextToLLM(t *testing.T) {
	chunk := &Chunk{Text: "Hello **bold** and *italic* and __under__ and _under2_"}
	expected := "Hello bold and italic and under and under2"
	if got := chunk.TextToLLM(); got != expected {
		t.Errorf("TextToLLM() = %q, want %q", got, expected)
	}
}

func TestIsContentChanged(t *testing.T) {
	chunk := &Chunk{ID: "id1"}
	if chunk.IsContentChanged("id1") {
		t.Error("same ID should return false")
	}
	if !chunk.IsContentChanged("id2") {
		t.Error("different ID should return true")
	}
}

func TestTokenEstimate(t *testing.T) {
	chunk := &Chunk{Metadata: ChunkMetadata{TokenCount: 42}}
	if got := chunk.TokenEstimate(); got != 42 {
		t.Errorf("TokenEstimate() = %d, want 42", got)
	}
}

func TestSetLinks(t *testing.T) {
	chunk := &Chunk{}
	links := []link.Link{{}, {}}
	chunk.SetLinks(links)
	if len(chunk.Metadata.Links) != 2 {
		t.Errorf("SetLinks did not set links, got %d", len(chunk.Metadata.Links))
	}
}

func TestMetadataJSON(t *testing.T) {
	chunk := &Chunk{Metadata: ChunkMetadata{SegmentID: "seg1", ChunkIndex: 0, TotalChunks: 1}}
	data, err := chunk.MetadataJSON()
	if err != nil {
		t.Fatalf("MetadataJSON() error: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["segment_id"] != "seg1" {
		t.Errorf("segment_id = %v, want seg1", m["segment_id"])
	}
}

func TestExtractStringList(t *testing.T) {
	fm := newTestFrontmatter()
	got := extractStringList(fm, "tags")
	want := []string{"a", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractStringList(tags) = %v, want %v", got, want)
	}
	if got := extractStringList(fm, "nonexistent"); got != nil {
		t.Errorf("extractStringList(nonexistent) = %v, want nil", got)
	}
	if got := extractStringList(fm, "title"); got != nil {
		t.Errorf("extractStringList(title) = %v, want nil (scalar)", got)
	}
	if got := extractStringList(nil, "tags"); got != nil {
		t.Errorf("extractStringList(nil) = %v, want nil", got)
	}
}

func TestExtractScalar(t *testing.T) {
	fm := newTestFrontmatter()
	if got := extractScalar(fm, "title"); got != "Test Note" {
		t.Errorf("extractScalar(title) = %q, want Test Note", got)
	}
	if got := extractScalar(fm, "tags"); got != "" {
		t.Errorf("extractScalar(tags) = %q, want empty (list)", got)
	}
	if got := extractScalar(fm, "nonexistent"); got != "" {
		t.Errorf("extractScalar(nonexistent) = %q, want empty", got)
	}
	if got := extractScalar(nil, "title"); got != "" {
		t.Errorf("extractScalar(nil) = %q, want empty", got)
	}
}