package note

// CreateSegmentsTree builds a hierarchical tree of segments from a flat, depth-annotated
// slice of ContentBlocks. It assumes that the blocks have already been processed by
// HierarchyBuilder (i.e. Metadata.Depth and Metadata.Path are correctly set).
// Non-heading blocks are placed into the OwnBlocks of the currently active segment.
// Heading blocks become new segments and are linked to their parent according to Depth.
func CreateSegmentsTree(blocks []ContentBlock) SegmentTree {
	root := &Segment{}
	var stack []*Segment
	cur := root

	for i := range blocks {
		b := &blocks[i]
		if b.Metadata.Type != TypeHeading {
			cur.OwnBlocks = append(cur.OwnBlocks, *b)
			continue
		}
		cur = processHeadingBlock(b, root, &stack)
	}

	return SegmentTree{Root: root}
}

// processHeadingBlock handles a heading block: updates the stack, creates a new segment,
// links it to its parent, and returns the new current segment.
func processHeadingBlock(block *ContentBlock, root *Segment, stack *[]*Segment) *Segment {
	depth := block.Metadata.Depth
	if depth < len(*stack) {
		*stack = (*stack)[:depth]
	}

	var parent *Segment
	if depth > 0 {
		parent = (*stack)[depth-1]
	} else {
		parent = root
	}

	seg := &Segment{
		Header: block,
		Parent: parent,
	}
	parent.Children = append(parent.Children, seg)

	*stack = append(*stack, seg)
	return seg
}

// Flatten returns a flat slice of all segments in the tree in depth-first order
// (parent before children).
func (st *SegmentTree) Flatten() []*Segment {
	var list []*Segment
	var walk func(seg *Segment)
	walk = func(seg *Segment) {
		list = append(list, seg)
		for _, child := range seg.Children {
			walk(child)
		}
	}
	walk(st.Root)
	return list
}
