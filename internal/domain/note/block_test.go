// block_test.go
package note

import (
	"testing"
)

func TestGenerateBlockID(t *testing.T) {
	tests := []struct {
		name     string
		meta     BlockMetadata
		expected string
	}{
		{
			name: "path and sequence",
			meta: BlockMetadata{Path: "Chapter 1", Sequence: 3},
			expected: "Chapter 1|3",
		},
		{
			name: "empty path",
			meta: BlockMetadata{Path: "", Sequence: 1},
			expected: "|1",
		},
		{
			name: "sequence zero",
			meta: BlockMetadata{Path: "P", Sequence: 0},
			expected: "P|0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateBlockID(tt.meta)
			if got != tt.expected {
				t.Errorf("GenerateBlockID() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestContentBlockID(t *testing.T) {
	b := ContentBlock{Metadata: BlockMetadata{Path: "A > B", Sequence: 5}}
	if got := b.ID(); got != "A > B|5" {
		t.Errorf("ID() = %q, want %q", got, "A > B|5")
	}
}

func TestContentBlockString(t *testing.T) {
	b := ContentBlock{Metadata: BlockMetadata{Type: TypeParagraph}, RawText: "Hello"}
	expected := `[paragraph] "Hello"`
	if got := b.String(); got != expected {
		t.Errorf("String() = %q, want %q", got, expected)
	}
}