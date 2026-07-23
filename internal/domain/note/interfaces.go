package note

// BlockParser defines the contract for parsing the note body into blocks.
type BlockParser interface {
	Parse(lines []string) ([]ContentBlock, error)
}

// HierarchyBuilder defines the contract for building hierarchical metadata.
type HierarchyBuilder interface {
	Build(blocks []ContentBlock) ([]ContentBlock, error)
}
