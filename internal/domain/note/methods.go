package note

// BlockByID returns a block by its stable ID.
func (n *Note) BlockByID(id string) *ContentBlock {
	for i := range n.Blocks {
		if n.Blocks[i].ID() == id {
			return &n.Blocks[i]
		}
	}
	return nil
}

// BlocksByType returns all blocks of the given type.
func (n *Note) BlocksByType(typ BlockType) []ContentBlock {
	var result []ContentBlock
	for _, b := range n.Blocks {
		if b.Metadata.Type == typ {
			result = append(result, b)
		}
	}
	return result
}

// Headings returns all heading blocks in order.
func (n *Note) Headings() []ContentBlock {
	return n.BlocksByType(TypeHeading)
}

// BlocksByPath returns all blocks belonging to the specified path.
func (n *Note) BlocksByPath(path string) []ContentBlock {
	var result []ContentBlock
	for _, b := range n.Blocks {
		if b.Metadata.Path == path {
			result = append(result, b)
		}
	}
	return result
}

// BlocksByDepth returns all blocks with the given nesting depth.
func (n *Note) BlocksByDepth(depth int) []ContentBlock {
	var result []ContentBlock
	for _, b := range n.Blocks {
		if b.Metadata.Depth == depth {
			result = append(result, b)
		}
	}
	return result
}
