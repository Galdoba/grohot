// segment_string_test.go
package note

import (
	"strings"
	"testing"

	"github.com/Galdoba/grohot/internal/domain/note/frontmatter/property"
)

func TestSegmentString(t *testing.T) {
	seg := &Segment{
		ID:        "seg1",
		FilePath:  "/vault/note.md",
		Ancestors: []string{"Parent"},
		DomType:   domTypeText,
		TokenCount: map[string]int{"char_div_4": 5},
		CharCount: 20,
		Header:    &ContentBlock{RawText: "## Section"},
		OwnBlocks: []ContentBlock{
			{RawText: "block1", Metadata: BlockMetadata{Type: TypeParagraph}},
		},
		NoteFrontmatter: newTestFrontmatter(),
		Parent:          &Segment{ID: "parentID"},
		Children:        []*Segment{{ID: "childID"}},
	}
	s := seg.String()
	checks := []string{
		"═══ Segment ═══",
		"ID:          seg1",
		"FilePath:    /vault/note.md",
		"Parent ID:  parentID",
		"Children:   [childID]",
		"Header:     \"## Section\"",
		"OwnBlocks:",
		"Note props:",
		"Projected Frontmatter: none",
	}
	for _, substr := range checks {
		if !strings.Contains(s, substr) {
			t.Errorf("Segment.String() missing %q in:\n%s", substr, s)
		}
	}
}

func TestFormatPropertyValue(t *testing.T) {
	tests := []struct {
		name string
		p    *property.Property
		want string
	}{
		{
			name: "short scalar",
			p:    &property.Property{Value: property.Value{Scalar: "hello"}},
			want: `"hello"`,
		},
		{
			name: "long scalar",
			p:    &property.Property{Value: property.Value{Scalar: strings.Repeat("a", 50)}},
			want: `"` + strings.Repeat("a", 40) + `..."`,
		},
		{
			name: "short list",
			p:    &property.Property{Value: property.Value{List: []string{"a", "b"}}},
			want: `[a, b]`,
		},
		{
			name: "long list",
			p:    &property.Property{Value: property.Value{List: []string{"a", "b", "c", "d"}}},
			want: `[a, b, c, ...] (4 total)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatPropertyValue(tt.p); got != tt.want {
				t.Errorf("formatPropertyValue() = %q, want %q", got, tt.want)
			}
		})
	}
}