// segment_visual_test.go
package note

import (
	"strings"
	"testing"
)

func TestVisual(t *testing.T) {
	// Build a small tree
	root := &Segment{
		OwnBlocks: []ContentBlock{{RawText: "root text", Metadata: BlockMetadata{Type: TypeParagraph}}},
	}
	child := &Segment{
		Header: &ContentBlock{RawText: "## Child", Metadata: BlockMetadata{Type: TypeHeading}},
		Parent: root,
	}
	root.Children = []*Segment{child}
	tree := SegmentTree{Root: root}

	vis := tree.Visual()
	if !strings.Contains(vis, "root") {
		t.Error("Visual() missing root")
	}
	if !strings.Contains(vis, "## Child") {
		t.Error("Visual() missing child heading")
	}
	if !strings.Contains(vis, "[paragraph]") {
		t.Error("Visual() missing paragraph block")
	}
}

func TestStripHeadingMarkers(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"# Title", "Title"},
		{"## Sub", "Sub"},
		{"###   Deep", "Deep"},
		{"NoHeading", "NoHeading"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := stripHeadingMarkers(tt.in); got != tt.want {
			t.Errorf("stripHeadingMarkers(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestHeadingLevel(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"# Title", 1},
		{"## Sub", 2},
		{"### Deep", 3},
		{"NoHeading", 0},
		{"", 0},
		{"####### Seven", 7}, // seven hashes
	}
	for _, tt := range tests {
		if got := headingLevel(tt.in); got != tt.want {
			t.Errorf("headingLevel(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestTruncatePad(t *testing.T) {
	if got := truncatePad("hello", 3); got != "hel" {
		t.Errorf("truncatePad(hello,3) = %q, want hel", got)
	}
	if got := truncatePad("hi", 4); got != "hi  " {
		t.Errorf("truncatePad(hi,4) = %q, want 'hi  '", got)
	}
	if got := truncatePad("", 2); got != "  " {
		t.Errorf("truncatePad('',2) = %q, want two spaces", got)
	}
}

func TestFormatHeadingText(t *testing.T) {
	got := formatHeadingText("## Section")
	if got != "## Section          " { // length = headingDisplayWidth (20)
		t.Errorf("formatHeadingText = %q", got)
	}
	if len([]rune(got)) != headingDisplayWidth {
		t.Errorf("formatHeadingText length = %d, want %d", len([]rune(got)), headingDisplayWidth)
	}
}

func TestFormatOwnBlockText(t *testing.T) {
	got := formatOwnBlockText("first line\nsecond line")
	expectedPrefix := "first line"
	expected := expectedPrefix + strings.Repeat(" ", ownBlockDisplayWidth-len(expectedPrefix))
	if got != expected {
		t.Errorf("formatOwnBlockText() = %q, want %q", got, expected)
	}
	if len([]rune(got)) != ownBlockDisplayWidth {
		t.Errorf("formatOwnBlockText length = %d, want %d", len([]rune(got)), ownBlockDisplayWidth)
	}
}

func TestChooseConnector(t *testing.T) {
	if chooseConnector(0, 2) != "├── " {
		t.Error("chooseConnector should return branch for non-last")
	}
	if chooseConnector(1, 2) != "└── " {
		t.Error("chooseConnector should return last for last")
	}
}
