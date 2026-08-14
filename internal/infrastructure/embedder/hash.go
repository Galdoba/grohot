package embedder

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"unicode"
)

type HashEmbedder struct {
	Dim       int
	StopWords map[string]bool
	UseNgrams bool
	NgramSize int // 2 для биграмм, 3 для триграмм, 0 – только униграммы
}

func NewHashEmbedder(dim int) *HashEmbedder {
	return &HashEmbedder{
		Dim:       dim,
		StopWords: defaultStopWords,
		UseNgrams: true,
		NgramSize: 3, // биграммы и триграммы
	}
}

func (e *HashEmbedder) Embed(text string) ([]float64, error) {
	if e.Dim <= 0 {
		return nil, fmt.Errorf("dimension must be positive")
	}
	tokens := tokenize(text, e.StopWords)
	if e.UseNgrams {
		tokens = append(tokens, generateNgrams(tokens, e.NgramSize)...)
	}

	vec := make([]float64, e.Dim)
	for _, tok := range tokens {
		h := hashToken(tok)
		idx := int(h % uint64(e.Dim))
		vec[idx] += 1.0
	}
	// L2 нормализация
	norm := 0.0
	for _, v := range vec {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range vec {
			vec[i] /= norm
		}
	}
	return vec, nil
}

func tokenize(s string, stopWords map[string]bool) []string {
	s = strings.ToLower(s)
	words := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	if stopWords == nil {
		return words
	}
	filtered := make([]string, 0, len(words))
	for _, w := range words {
		if !stopWords[w] {
			filtered = append(filtered, w)
		}
	}
	return filtered
}

// generateNgrams возвращает все n-граммы от 2 до maxN.
func generateNgrams(tokens []string, maxN int) []string {
	if maxN < 2 {
		return nil
	}
	var ngrams []string
	for n := 2; n <= maxN && n <= len(tokens); n++ {
		for i := 0; i+n <= len(tokens); i++ {
			ngrams = append(ngrams, strings.Join(tokens[i:i+n], "_"))
		}
	}
	return ngrams
}

func hashToken(tok string) uint64 {
	h := sha256.Sum256([]byte(tok))
	return binary.BigEndian.Uint64(h[:8])
}

var defaultStopWords = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
	"of": true, "to": true, "in": true, "on": true, "for": true, "with": true,
	"is": true, "are": true, "was": true, "were": true, "be": true, "been": true,
	"as": true, "at": true, "by": true, "from": true, "this": true, "that": true,
	"these": true, "those": true, "it": true, "its": true, "if": true,
	"и": true, "в": true, "на": true, "с": true, "по": true, "к": true, "о": true,
	"не": true, "за": true, "из": true, "от": true, "для": true, "как": true,
}
