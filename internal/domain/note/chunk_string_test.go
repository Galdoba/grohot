// chunk_string_test.go
package note

import (
	"strings"
	"testing"
)

func TestChunkString(t *testing.T) {
	chunk := Chunk{
		ID:   "seg1:abc123",
		Text: "Hello",
		Metadata: ChunkMetadata{
			SegmentID:   "seg1",
			ChunkIndex:  0,
			TotalChunks: 1,
			FilePath:    "/vault/note.md",
			HeadingText: "Section",
		},
	}
	s := chunk.String()
	checks := []string{
		"═══ Chunk ═══",
		"ID:        seg1:abc123",
		"Embedding: <empty>",
		"Text:",
		"Hello",
		"Metadata:",
		"SegmentID:    seg1",
		"ChunkIndex:   0",
		"TotalChunks:  1",
		"FilePath:     /vault/note.md",
		"HeadingText:  \"Section\"",
		"Links:        []",
	}
	for _, substr := range checks {
		if !strings.Contains(s, substr) {
			t.Errorf("Chunk.String() missing substring %q in:\n%s", substr, s)
		}
	}
}