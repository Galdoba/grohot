package note

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Galdoba/grohot/internal/domain/link"
	"github.com/Galdoba/grohot/internal/domain/note/frontmatter"
)

// Chunk представляет собой единицу контента, готовую для эмбеддинга и индексации.
type Chunk struct {
	ID        string        // стабильный идентификатор: segmentID + ":" + hash(Text)
	Text      string        // исходный текст сегмента (без крошек)
	Embedding []float64     // векторное представление, заполняется после эмбеддинга
	Metadata  ChunkMetadata // структурированные метаданные
}

// ChunkInput содержит параметры для создания чанка из сегмента.
type ChunkInput struct {
	Text        string // текст этого конкретного чанка
	Index       int    // порядковый номер чанка внутри сегмента (0‑based)
	TotalChunks int    // общее число чанков, на которые разбит сегмент
}

// ChunkMetadata holds all information needed to describe, filter, and
// reconstruct the context of a chunk.
type ChunkMetadata struct {
	// Идентификация и порядок
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
	Links []link.Link `json:"links,omitempty"`
}

// NewChunk создаёт один Chunk из сегмента и входных данных.
// Возвращает ошибку, если входные параметры некорректны.
func NewChunk(seg *Segment, input ChunkInput) (*Chunk, error) {
	if err := validateChunkInput(input); err != nil {
		return nil, err
	}
	if seg == nil {
		return nil, fmt.Errorf("segment cannot be nil")
	}

	chunkID := computeChunkID(seg.ID, input.Text)

	tokenCount := len([]rune(input.Text)) / runesPerToken

	meta := ChunkMetadata{
		SegmentID:   seg.ID,
		ChunkIndex:  input.Index,
		TotalChunks: input.TotalChunks,
		FilePath:    seg.FilePath,
		Ancestors:   seg.Ancestors,
		DomType:     seg.DomType,
		CharCount:   len([]rune(input.Text)),
		Links:       nil, // будет заполнено позже
	}
	meta.TokenCount = tokenCount

	// Заголовок сегмента
	if seg.Header != nil {
		meta.HeadingText = stripHeadingMarkers(seg.Header.RawText)
		meta.HeadingLevel = headingLevel(seg.Header.RawText)
	}

	// Количество токенов из карты (первое доступное)
	if count, ok := seg.TokenCount["char_div_4"]; ok {
		meta.TokenCount = count
	} else {
		// fallback
		for _, count := range seg.TokenCount {
			meta.TokenCount = count
			break
		}
	}

	// Свойства из frontmatter заметки
	if seg.NoteFrontmatter != nil {
		meta.Tags = extractStringList(seg.NoteFrontmatter, "tags")
		meta.Aliases = extractStringList(seg.NoteFrontmatter, "aliases")
		meta.Created = extractScalar(seg.NoteFrontmatter, "created")
		meta.Updated = extractScalar(seg.NoteFrontmatter, "updated")
	}

	chunk := &Chunk{
		ID:       chunkID,
		Text:     input.Text,
		Metadata: meta,
	}
	return chunk, nil
}

// validateChunkInput проверяет входные параметры чанка.
func validateChunkInput(input ChunkInput) error {
	if strings.TrimSpace(input.Text) == "" {
		return fmt.Errorf("chunk text cannot be empty")
	}
	if input.TotalChunks <= 0 {
		return fmt.Errorf("total chunks must be positive")
	}
	if input.Index < 0 || input.Index >= input.TotalChunks {
		return fmt.Errorf("chunk index %d out of range [0, %d)", input.Index, input.TotalChunks)
	}
	return nil
}

// computeChunkID возвращает стабильный идентификатор чанка.
// Формат: <segmentID>:<hex hash первых 16 байт SHA-256 от текста чанка>.
func computeChunkID(segmentID, chunkText string) string {
	h := sha256.Sum256([]byte(chunkText))
	return fmt.Sprintf("%s:%x", segmentID, h[:16])
}

// TextToEmbed возвращает текст чанка, предварённый контекстной информацией
// (путь к файлу, предки, заголовок текущего сегмента) для улучшения эмбеддинга.
func (c *Chunk) TextToEmbed() string {
	var lines []string

	// 1. Полный путь к файлу
	if c.Metadata.FilePath != "" {
		lines = append(lines, c.Metadata.FilePath)
	}

	// 2. Хлебные крошки (предки + текущий заголовок)
	if len(c.Metadata.Ancestors) > 0 || c.Metadata.HeadingText != "" {
		parts := append(c.Metadata.Ancestors, c.Metadata.HeadingText)
		lines = append(lines, strings.Join(parts, " > "))
	}

	// 3. Content Type для нетекстовых чанков
	if c.Metadata.DomType != "" && c.Metadata.DomType != "text" {
		lines = append(lines, "Content Type: "+c.Metadata.DomType)
	}

	if len(lines) == 0 {
		return c.Text
	}

	// 4. Разделитель: две пустые строки
	return strings.Join(lines, "\n") + "\n\n\n" + c.Text
}

// TextToLLM возвращает очищенный от Markdown-форматирования текст чанка,
// пригодный для использования в промпте.
func (c *Chunk) TextToLLM() string {
	replacer := strings.NewReplacer(
		"**", "",
		"*", "",
		"__", "",
		"_", "",
	)
	return replacer.Replace(c.Text)
}

// IsContentChanged сравнивает текущий ID чанка с ранее сохранённым.
// Возвращает true, если текст чанка изменился (изменился ID, т.к. ID включает хеш текста).
func (c *Chunk) IsContentChanged(oldChunkID string) bool {
	return c.ID != oldChunkID
}

// TokenEstimate возвращает приблизительное количество токенов из метаданных.
func (c *Chunk) TokenEstimate() int {
	return c.Metadata.TokenCount
}

// SetLinks устанавливает горизонтальные связи в метаданных чанка.
func (c *Chunk) SetLinks(links []link.Link) {
	c.Metadata.Links = links
}

// MetadataJSON возвращает JSON-представление метаданных чанка.
func (c *Chunk) MetadataJSON() ([]byte, error) {
	return json.Marshal(c.Metadata)
}

// --- вспомогательные функции для извлечения свойств frontmatter ---

// extractStringList извлекает список строк из свойства с указанным именем.
// Если свойство отсутствует или не является списком, возвращает nil.
func extractStringList(fm *frontmatter.Frontmatter, name string) []string {
	if fm == nil {
		return nil
	}
	p := fm.Get(name)
	if p == nil || p.Value.List == nil {
		return nil
	}
	return p.Value.List
}

// extractScalar извлекает скалярное значение свойства с указанным именем.
// Возвращает пустую строку, если свойство отсутствует или не является скаляром.
func extractScalar(fm *frontmatter.Frontmatter, name string) string {
	if fm == nil {
		return ""
	}
	p := fm.Get(name)
	if p == nil || p.Value.List != nil {
		return ""
	}
	return p.Value.Scalar
}

// PropertyAliases import can be removed if not used; we only need the strings.
