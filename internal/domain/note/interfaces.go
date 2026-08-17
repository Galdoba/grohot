package note

// BlockParser defines the contract for parsing the note body into blocks.
type BlockParser interface {
	Parse(lines []string) ([]ContentBlock, error)
}

// HierarchyBuilder defines the contract for building hierarchical metadata.
type HierarchyBuilder interface {
	Build(blocks []ContentBlock, filepath string) ([]ContentBlock, error)
}

// Chunker defines the contract for converting a segment into one or more chunks.
type Chunker interface {
	Chunk(segment *Segment) ([]Chunk, error)
}

// Embedder defines the contract for obtaining a vector representation of text.
type Embedder interface {
	Embed(text string) ([]float64, error)
}
