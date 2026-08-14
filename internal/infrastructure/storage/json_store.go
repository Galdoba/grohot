package storage

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"

	"github.com/Galdoba/grohot/internal/domain/note"
)

// JSONStore хранит чанки в памяти и умеет сохранять/загружать их в JSON‑файл.
// Поддерживает простой поиск по косинусной близости.
type JSONStore struct {
	path   string
	chunks []note.Chunk
}

func NewJSONStore(path string) *JSONStore {
	return &JSONStore{path: path}
}

// Insert добавляет чанк в память.
func (s *JSONStore) Insert(c note.Chunk) {
	s.chunks = append(s.chunks, c)
}

// Save записывает все чанки в JSON‑файл.
func (s *JSONStore) Save() error {
	data, err := json.MarshalIndent(s.chunks, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal chunks: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0666); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

// Load читает чанки из JSON‑файла.
func (s *JSONStore) Load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	if err := json.Unmarshal(data, &s.chunks); err != nil {
		return fmt.Errorf("unmarshal chunks: %w", err)
	}
	return nil
}

// SearchResult содержит результат поиска.
type SearchResult struct {
	Chunk note.Chunk
	Score float64
}

// Search возвращает topK чанков, отсортированных по убыванию косинусной близости.
func (s *JSONStore) Search(queryEmbedding []float64, topK int) ([]SearchResult, error) {
	if topK <= 0 {
		return nil, fmt.Errorf("topK must be positive")
	}
	results := make([]SearchResult, 0, len(s.chunks))
	for _, c := range s.chunks {
		if len(c.Embedding) == 0 {
			continue
		}
		score, err := cosineSimilarity(queryEmbedding, c.Embedding)
		if err != nil {
			return nil, err
		}
		results = append(results, SearchResult{Chunk: c, Score: score})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if len(results) > topK {
		results = results[:topK]
	}
	return results, nil
}

// cosineSimilarity вычисляет косинусную близость двух векторов одинаковой длины.
func cosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vector length mismatch: %d vs %d", len(a), len(b))
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0, nil
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB)), nil
}
