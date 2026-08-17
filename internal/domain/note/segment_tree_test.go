// segment_tree_test.go
package note

import "testing"

func TestCreateSegmentsTree(t *testing.T) {
	blocks := []ContentBlock{
		{Metadata: BlockMetadata{Type: TypeParagraph, Depth: 0, Path: "root"}}, // root own block
		{Metadata: BlockMetadata{Type: TypeHeading, Depth: 0, Path: "H1", Sequence: 1}},
		{Metadata: BlockMetadata{Type: TypeParagraph, Depth: 0, Path: "H1"}},
		{Metadata: BlockMetadata{Type: TypeHeading, Depth: 1, Path: "H1 > H2", Sequence: 1}},
		{Metadata: BlockMetadata{Type: TypeCode, Depth: 1, Path: "H1 > H2"}},
		{Metadata: BlockMetadata{Type: TypeHeading, Depth: 0, Path: "H3", Sequence: 2}},
	}

	tree := CreateSegmentsTree(blocks)
	if tree.Root == nil {
		t.Fatal("root should not be nil")
	}
	// root own blocks: one paragraph
	if len(tree.Root.OwnBlocks) != 1 {
		t.Errorf("root own blocks = %d, want 1", len(tree.Root.OwnBlocks))
	}
	// root children: H1 and H3
	if len(tree.Root.Children) != 2 {
		t.Fatalf("root children = %d, want 2", len(tree.Root.Children))
	}
	h1 := tree.Root.Children[0]
	if h1.Header == nil || h1.Header.Metadata.Path != "H1" {
		t.Error("H1 not correctly set")
	}
	if len(h1.OwnBlocks) != 1 { // paragraph under H1
		t.Errorf("H1 own blocks = %d, want 1", len(h1.OwnBlocks))
	}
	if len(h1.Children) != 1 {
		t.Fatalf("H1 children = %d, want 1", len(h1.Children))
	}
	h2 := h1.Children[0]
	if h2.Header == nil || h2.Header.Metadata.Path != "H1 > H2" {
		t.Error("H2 not correctly set")
	}
	if len(h2.OwnBlocks) != 1 { // code block
		t.Errorf("H2 own blocks = %d, want 1", len(h2.OwnBlocks))
	}
}

func TestFlatten(t *testing.T) {
	// Manually build tree
	root := &Segment{ID: "root"}
	child1 := &Segment{ID: "child1", Parent: root}
	child2 := &Segment{ID: "child2", Parent: root}
	grandchild := &Segment{ID: "grandchild", Parent: child1}
	root.Children = []*Segment{child1, child2}
	child1.Children = []*Segment{grandchild}

	tree := SegmentTree{Root: root}
	segs := tree.Flatten()
	expectedIDs := []string{"root", "child1", "grandchild", "child2"}
	if len(segs) != len(expectedIDs) {
		t.Fatalf("Flatten returned %d segments, want %d", len(segs), len(expectedIDs))
	}
	for i, seg := range segs {
		if seg.ID != expectedIDs[i] {
			t.Errorf("Flatten[%d] = %s, want %s", i, seg.ID, expectedIDs[i])
		}
	}
}