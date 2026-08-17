// methods_test.go
package note

import "testing"

func sampleNote() *Note {
	return &Note{
		Name:     "test",
		Filepath: "/vault/test.md",
		Blocks: []ContentBlock{
			{Metadata: BlockMetadata{Path: "P1", Sequence: 1, Type: TypeHeading, Depth: 0}},
			{Metadata: BlockMetadata{Path: "P1", Sequence: 2, Type: TypeParagraph, Depth: 0}},
			{Metadata: BlockMetadata{Path: "P2", Sequence: 1, Type: TypeHeading, Depth: 1}},
			{Metadata: BlockMetadata{Path: "P2", Sequence: 2, Type: TypeCode, Depth: 1}},
			{Metadata: BlockMetadata{Path: "P2", Sequence: 3, Type: TypeTable, Depth: 1}},
		},
		Tree: SegmentTree{Root: &Segment{}},
	}
}

func TestBlockByID(t *testing.T) {
	n := sampleNote()
	id := GenerateBlockID(BlockMetadata{Path: "P1", Sequence: 2})
	block := n.BlockByID(id)
	if block == nil {
		t.Fatal("BlockByID returned nil")
	}
	if block.Metadata.Type != TypeParagraph {
		t.Errorf("expected paragraph, got %v", block.Metadata.Type)
	}
	if n.BlockByID("nonexistent") != nil {
		t.Error("BlockByID should return nil for unknown id")
	}
}

func TestBlocksByType(t *testing.T) {
	n := sampleNote()
	headings := n.BlocksByType(TypeHeading)
	if len(headings) != 2 {
		t.Errorf("expected 2 headings, got %d", len(headings))
	}
	for _, b := range headings {
		if b.Metadata.Type != TypeHeading {
			t.Errorf("got non-heading block in result")
		}
	}
}

func TestHeadings(t *testing.T) {
	n := sampleNote()
	h := n.Headings()
	if len(h) != 2 {
		t.Errorf("Headings() returned %d, want 2", len(h))
	}
}

func TestBlocksByPath(t *testing.T) {
	n := sampleNote()
	blocks := n.BlocksByPath("P1")
	if len(blocks) != 2 {
		t.Errorf("BlocksByPath(P1) returned %d, want 2", len(blocks))
	}
	blocks = n.BlocksByPath("P2")
	if len(blocks) != 3 {
		t.Errorf("BlocksByPath(P2) returned %d, want 3", len(blocks))
	}
	if len(n.BlocksByPath("nope")) != 0 {
		t.Error("BlocksByPath(nope) should be empty")
	}
}

func TestBlocksByDepth(t *testing.T) {
	n := sampleNote()
	d0 := n.BlocksByDepth(0)
	if len(d0) != 2 {
		t.Errorf("BlocksByDepth(0) returned %d, want 2", len(d0))
	}
	d1 := n.BlocksByDepth(1)
	if len(d1) != 3 {
		t.Errorf("BlocksByDepth(1) returned %d, want 3", len(d1))
	}
}