// segment_metadata_test.go
package note

import (
	"reflect"
	"strings"
	"testing"
)

func TestPopulateMetadata(t *testing.T) {
	// Build a simple tree
	root := &Segment{}
	child := &Segment{
		Header: &ContentBlock{RawText: "## Child", Metadata: BlockMetadata{Type: TypeHeading}},
		Parent: root,
	}
	root.Children = []*Segment{child}

	// Add some own blocks to child
	child.OwnBlocks = []ContentBlock{
		{RawText: "para", Metadata: BlockMetadata{Type: TypeParagraph}},
		{RawText: "```code```", Metadata: BlockMetadata{Type: TypeCode}},
	}

	st := &SegmentTree{Root: root}
	fm := newTestFrontmatter()
	st.PopulateMetadata("/vault/note.md", fm)

	// Check root
	if root.FilePath != "/vault/note.md" {
		t.Errorf("root FilePath = %q", root.FilePath)
	}
	if root.NoteFrontmatter != fm {
		t.Error("root NoteFrontmatter not set")
	}
	if root.ID == "" {
		t.Error("root ID empty")
	}
	if root.Ancestors != nil {
		t.Errorf("root Ancestors should be nil, got %v", root.Ancestors)
	}
	if root.CharCount == 0 {
		t.Error("root CharCount should be > 0")
	}
	if root.DomType != domTypeText {
		t.Errorf("root DomType = %q, want text", root.DomType)
	}

	// Check child
	if child.ID == "" {
		t.Error("child ID empty")
	}
	if len(child.Ancestors) != 0 { // root has no header, so ancestors should be empty
		t.Errorf("child Ancestors = %v, want empty", child.Ancestors)
	}
	if child.DomType != domTypeMixed { // has code and paragraph -> text? actually only table+code gives mixed, code only -> code
		// child has paragraph and code, no table, so should be code? determineDomType returns code if hasCode
		if child.DomType != domTypeCode {
			t.Errorf("child DomType = %q, want code", child.DomType)
		}
	}
}

func TestToChunkText(t *testing.T) {
	t.Run("root with title", func(t *testing.T) {
		seg := &Segment{
			OwnBlocks:       []ContentBlock{{RawText: "Body"}},
			NoteFrontmatter: newTestFrontmatter(),
		}
		got := seg.ToChunkText()
		want := "Note: Test Note\n\nBody"
		if got != want {
			t.Errorf("ToChunkText() = %q, want %q", got, want)
		}
	})
	t.Run("non-root without title", func(t *testing.T) {
		seg := &Segment{
			Header:    &ContentBlock{RawText: "## Section"},
			OwnBlocks: []ContentBlock{{RawText: "Text"}},
		}
		got := seg.ToChunkText()
		want := "Text"
		if got != want {
			t.Errorf("ToChunkText() = %q, want %q", got, want)
		}
	})
}

func TestText(t *testing.T) {
	seg := &Segment{OwnBlocks: []ContentBlock{{RawText: "A"}, {RawText: "B"}}}
	if got := seg.Text(); got != "A\n\nB" {
		t.Errorf("Text() = %q, want %q", got, "A\n\nB")
	}
}

func TestComputeID(t *testing.T) {
	seg := &Segment{
		Header:    &ContentBlock{RawText: "## Section"},
		Ancestors: []string{"Chapter"},
	}
	id1 := seg.computeID("/vault/note.md")
	id2 := seg.computeID("/vault/note.md")
	if id1 != id2 {
		t.Error("computeID should be deterministic")
	}
	if id1 == "" {
		t.Error("computeID returned empty")
	}
	// change ancestor and expect different
	seg.Ancestors = []string{"Other"}
	if id1 == seg.computeID("/vault/note.md") {
		t.Error("computeID should change with different ancestors")
	}
}

func TestGetAncestors(t *testing.T) {
	root := &Segment{}
	child := &Segment{Parent: root}
	grandchild := &Segment{Parent: child}

	if got := root.getAncestors(); got != nil {
		t.Errorf("root getAncestors = %v, want nil", got)
	}
	if got := child.getAncestors(); len(got) != 0 {
		t.Errorf("child getAncestors = %v, want empty slice", got)
	}

	// now add header to root
	root.Header = &ContentBlock{RawText: "# Root"}
	if got := child.getAncestors(); !reflect.DeepEqual(got, []string{"Root"}) {
		t.Errorf("child getAncestors = %v, want [Root]", got)
	}
	if got := grandchild.getAncestors(); !reflect.DeepEqual(got, []string{"Root"}) {
		t.Errorf("grandchild getAncestors = %v, want [Root]", got)
	}
}

func TestComputeCharCount(t *testing.T) {
	seg := &Segment{OwnBlocks: []ContentBlock{{RawText: "abc"}, {RawText: "def"}}}
	// collectFullText adds "\n" after each block -> "abc\ndef\n" = 8 runes
	if got := seg.computeCharCount(); got != 8 {
		t.Errorf("computeCharCount = %d, want 8", got)
	}
}

func TestComputeTokenCounts(t *testing.T) {
	seg := &Segment{OwnBlocks: []ContentBlock{{RawText: "12345678"}}} // 8 runes
	counts := seg.computeTokenCounts()
	if counts["char_div_4"] != 2 {
		t.Errorf("char_div_4 = %d, want 2", counts["char_div_4"])
	}
}

func TestDetermineDomType(t *testing.T) {
	tests := []struct {
		name   string
		blocks []ContentBlock
		want   string
	}{
		{"empty", nil, domTypeText},
		{"paragraph only", []ContentBlock{{Metadata: BlockMetadata{Type: TypeParagraph}}}, domTypeText},
		{"code only", []ContentBlock{{Metadata: BlockMetadata{Type: TypeCode}}}, domTypeCode},
		{"table only", []ContentBlock{{Metadata: BlockMetadata{Type: TypeTable}}}, domTypeTable},
		{"table and code", []ContentBlock{{Metadata: BlockMetadata{Type: TypeTable}}, {Metadata: BlockMetadata{Type: TypeCode}}}, domTypeMixed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seg := &Segment{OwnBlocks: tt.blocks}
			if got := seg.determineDomType(); got != tt.want {
				t.Errorf("determineDomType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCollectFullText(t *testing.T) {
	child := &Segment{OwnBlocks: []ContentBlock{{RawText: "Child"}}}
	seg := &Segment{
		OwnBlocks: []ContentBlock{{RawText: "Parent"}},
		Children:  []*Segment{child},
	}
	got := seg.collectFullText()
	if !strings.Contains(got, "Parent") || !strings.Contains(got, "Child") {
		t.Errorf("collectFullText missing content: %q", got)
	}
}

func TestBreadCrumbs(t *testing.T) {
	seg := &Segment{
		FilePath:  "/vault/note.md",
		Ancestors: []string{"Chapter 1"},
		Header:    &ContentBlock{RawText: "## Section"},
	}
	got := seg.BreadCrumbs()
	want := "/vault/note.md > Chapter 1 > Section"
	if got != want {
		t.Errorf("BreadCrumbs() = %q, want %q", got, want)
	}
}

func TestHeadingText(t *testing.T) {
	seg := &Segment{Header: &ContentBlock{RawText: "### Deep"}}
	if got := seg.HeadingText(); got != "Deep" {
		t.Errorf("HeadingText() = %q, want Deep", got)
	}
	if seg2 := (&Segment{}); seg2.HeadingText() != "" {
		t.Errorf("HeadingText() with nil header should be empty")
	}
}
