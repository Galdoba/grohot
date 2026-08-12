package note

// Chunk represents a piece of content ready for embedding and indexing.
type Chunk struct {

	// стабильный идентификатор (хеш) - строим как fileID + hash(chunkText)
	ID string

	// Text содержит текст чанка, как есть из Сегмента.
	Text string

	Embedding []float64 //удаляется перед внесением в базу, используется как ключ в базе

	Metadata ChunkMetadata // структурированные метаданные для поиска заполняется от Segments
}

// ChunkMetadata holds all information needed to describe, filter, and
// reconstruct the context of a chunk.
type ChunkMetadata struct {
	// Идентификация и порядок
	ChunkID     string `json:"chunk_id"`
	SegmentID   string `json:"segment_id"`
	ChunkIndex  int    `json:"chunk_index"`  // номер чанка внутри сегмента (0-based)
	TotalChunks int    `json:"total_chunks"` // общее число чанков в этом сегменте

	// Источник и иерархия
	FilePath     string   `json:"file_path"`
	Ancestors    []string `json:"ancestors"`     // ["Chapter 1"]
	HeadingText  string   `json:"heading_text"`  //текст заголовка сегмента (без #).
	HeadingLevel int      `json:"heading_level"` // 1..6, 0 для корня. (количество #)

	// Характеристики контента
	DomType    string `json:"dom_type"`    // "text", "table", "code", "mixed"
	TokenCount int    `json:"token_count"` // приблизительное число токенов
	CharCount  int    `json:"char_count"`

	// Свойства из frontmatter заметки
	Tags    []string `json:"tags,omitempty"`
	Aliases []string `json:"aliases,omitempty"`
	Created string   `json:"created,omitempty"`
	Updated string   `json:"updated,omitempty"`

	// Горизонтальные связи (будут добавлены позже)
	Links []Link `json:"links,omitempty"`
}

// Link represents a directed connection from this chunk to another note or segment.
type Link struct {
	TargetFile      string `json:"target_file"`
	TargetHeading   string `json:"target_heading,omitempty"`
	TargetSegmentID string `json:"target_segment_id,omitempty"`
	LinkType        string `json:"link_type"` // "wikilink", "backlink", "transclusion"
}

func (ch *Chunk) TextToEbbed() string {
	//берёт Breadcrumb и Ancestors из Metadata и добавляет их перед Text.
	return ""
}

func (ch *Chunk) TextToLLM() string {
	//удаляет из Text форматирование (звёздочки, подчёркивания и т.п.), может оставить только plain text.
	return ""
}
