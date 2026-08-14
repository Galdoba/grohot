// File: internal/domain/note/link/link.go
package link

import (
	"fmt"
	"regexp"
	"strings"
)

// Type представляет тип ссылки.
type Type string

const (
	// Direct — исходящая ссылка, цель найдена.
	Direct Type = "direct"
	// BackLink — входящая ссылка (обратная сторона Direct).
	BackLink Type = "backlink"
	// DeadLink — исходящая ссылка, цель не найдена.
	DeadLink Type = "deadlink"
)

// Link представляет одну вики-ссылку Obsidian.
// Source* поля заполняются Vault'ом после разрешения.
type Link struct {
	SourceFile      string `json:"source_file"`                 // относительный путь (с .md)
	SourceSegmentID string `json:"source_segment_id"`           // ID сегмента-источника
	TargetFile      string `json:"target_file"`                 // относительный путь (с .md)
	TargetHeading   string `json:"target_heading,omitempty"`    // заголовок целевого сегмента
	TargetSegmentID string `json:"target_segment_id,omitempty"` // ID целевого сегмента
	DisplayText     string `json:"display_text,omitempty"`
	IsTransclusion  bool   `json:"is_transclusion,omitempty"`
	LinkType        Type   `json:"link_type"`
}

// linkRegex находит все вики-ссылки в тексте, включая трансклюзии (![[...]]).
var linkRegex = regexp.MustCompile(`!?\[\[([^\[\]]+)\]\]`)

// Extract возвращает все ссылки, найденные в rawText.
// При обнаружении невалидной ссылки возвращает ошибку.
func Extract(rawText string) ([]Link, error) {
	links := make([]Link, 0)
	var parseErrors []string
	for _, linkText := range extractFromText(rawText) {
		l, err := parseLink(linkText)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("%q: %v", linkText, err))
			continue
		}
		links = append(links, l)
	}
	if len(parseErrors) > 0 {
		return links, fmt.Errorf("some links failed to parse: %s", strings.Join(parseErrors, "; "))
	}
	return links, nil
}

// extractFromText возвращает срез всех подстрок, содержащих вики-ссылки.
func extractFromText(rawText string) []string {
	return linkRegex.FindAllString(rawText, -1)
}

// parseLink разбирает текст вида "[[Target|Display]]" или "![[Target]]".
// Возвращает Link с LinkType=Direct; DeadLink будет установлен Vault'ом,
// если цель не будет найдена. Source* поля остаются пустыми и заполняются позже.
func parseLink(linkText string) (Link, error) {
	// Определяем трансклюзию
	isTransclusion := false
	if len(linkText) > 0 && linkText[0] == '!' {
		isTransclusion = true
		linkText = linkText[1:] // убираем '!'
	}

	// Убираем [[ и ]]
	inner := linkText
	if len(inner) >= 4 && inner[:2] == "[[" && inner[len(inner)-2:] == "]]" {
		inner = inner[2 : len(inner)-2]
	} else {
		return Link{}, fmt.Errorf("invalid wikilink format")
	}

	// Разделяем на цель и отображаемый текст
	display := ""
	target := inner
	if idx := indexOfPipe(inner); idx != -1 {
		target = inner[:idx]
		display = inner[idx+1:]
	}

	// Разделяем цель на файл и заголовок
	filePart := ""
	headingPart := ""
	if idx := indexOfHash(target); idx != -1 {
		filePart = target[:idx]
		headingPart = target[idx+1:]
	} else {
		filePart = target
	}

	// Trim пробелы
	filePart = trimSpace(filePart)
	headingPart = trimSpace(headingPart)
	display = trimSpace(display)

	// Проверяем, что есть хотя бы цель или заголовок
	if filePart == "" && headingPart == "" {
		return Link{}, fmt.Errorf("link target cannot be empty")
	}

	return Link{
		TargetFile:     filePart,
		TargetHeading:  headingPart,
		DisplayText:    display,
		IsTransclusion: isTransclusion,
		LinkType:       Direct, // пока, Vault может изменить на DeadLink
	}, nil
}

// Вспомогательные функции для работы с байтами (чтобы не импортировать strings)

func indexOfPipe(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '|' {
			return i
		}
	}
	return -1
}

func indexOfHash(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '#' {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// String возвращает подробное человекочитаемое представление ссылки.
func (l Link) String() string {
	targetHeading := ""
	if l.TargetHeading != "" {
		targetHeading = " #" + l.TargetHeading
	}
	return fmt.Sprintf("Link{Type:%s, Source:%s|%s, Target:%s|%s%s, Display:%q, Transclusion:%t}",
		l.LinkType,
		l.SourceFile, truncateID(l.SourceSegmentID),
		l.TargetFile, truncateID(l.TargetSegmentID),
		targetHeading,
		l.DisplayText,
		l.IsTransclusion)
}

// truncateID сокращает длинный идентификатор для отладки.
func truncateID(id string) string {
	if len(id) > 8 {
		return id[:8] + "..."
	}
	return id
}
