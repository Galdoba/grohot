package note

// BlockParser defines the contract for parsing the note body into blocks.
type BlockParser interface {
	Parse(lines []string) ([]ContentBlock, error)
}

// HierarchyBuilder defines the contract for building hierarchical metadata.
type HierarchyBuilder interface {
	Build(blocks []ContentBlock, filepath string) ([]ContentBlock, error)
}

// Chunker определяет контракт для преобразования сегмента в один или несколько чанков.
type Chunker interface {
	Chunk(segment *Segment) ([]Chunk, error)
}

// Embedder определяет контракт для получения векторного представления текста.
type Embedder interface {
	Embed(text string) ([]float64, error)
}
