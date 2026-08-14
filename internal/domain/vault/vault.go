// File: internal/infrastructure/vault/vault.go
package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Galdoba/grohot/internal/domain/link"
	"github.com/Galdoba/grohot/internal/domain/note"
)

// Vault представляет собой Obsidian-хранилище.
type Vault struct {
	rootDir string
	notes   map[string]*note.Note // ключ: относительный путь (с .md), нормализованный с '/'
	byName  map[string][]string   // ключ: нижний регистр имени файла без .md → список относительных путей
	parser  note.BlockParser
	builder note.HierarchyBuilder
	chunker note.Chunker
	debug   bool
}

// NewVault создаёт новый Vault для указанной корневой директории.
// Возвращает ошибку, если rootDir не является корнем Obsidian-хранилища
// (отсутствует директория .obsidian).
func NewVault(rootDir string, parser note.BlockParser, builder note.HierarchyBuilder, chunker note.Chunker) (*Vault, error) {
	if !isVaultRoot(rootDir) {
		return nil, fmt.Errorf("%s is not an Obsidian vault root", rootDir)
	}
	return &Vault{
		rootDir: rootDir,
		notes:   make(map[string]*note.Note),
		byName:  make(map[string][]string),
		parser:  parser,
		builder: builder,
		chunker: chunker,
		debug:   true,
	}, nil
}

// isVaultRoot проверяет, что директория dir является корнем Obsidian-хранилища.
// Для этого она должна содержать поддиректорию .obsidian, а внутри неё —
// файлы types.json и graph.json.
func isVaultRoot(dir string) bool {
	obsidianDir := filepath.Join(dir, ".obsidian")
	info, err := os.Stat(obsidianDir)
	if err != nil || !info.IsDir() {
		return false
	}

	// Проверяем обязательные файлы внутри .obsidian
	requiredFiles := []string{"types.json", "graph.json"}
	for _, name := range requiredFiles {
		if _, err := os.Stat(filepath.Join(obsidianDir, name)); err != nil {
			return false
		}
	}
	return true
}

// FindVaultRoot поднимается вверх от startPath до тех пор, пока не найдёт папку .obsidian.
func FindVaultRoot(startPath string) (string, error) {
	abs, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}
	for {
		if isVaultRoot(abs) {
			return abs, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			break
		}
		abs = parent
	}
	return "", fmt.Errorf("vault root not found from %s", startPath)
}

// ScanAndLoad рекурсивно обходит все .md-файлы в хранилище и загружает их.
// Предыдущие данные очищаются.
func (v *Vault) ScanAndLoad() error {
	if v.debug {
		fmt.Println("[vault] scanning .md files...")
	}
	v.notes = make(map[string]*note.Note)
	v.byName = make(map[string][]string)

	fileCount := 0
	err := filepath.Walk(v.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Пропускаем скрытые директории и .obsidian
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(path), ".md") {
			return nil
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(v.rootDir, absPath)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath) // нормализуем на '/'

		if v.debug {
			fmt.Printf("[vault] loading %s\n", relPath)
		}
		n, err := note.Load(absPath, v.parser, v.builder)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", relPath, err)
		}
		v.notes[relPath] = n
		name := strings.ToLower(strings.TrimSuffix(filepath.Base(relPath), ".md"))
		v.byName[name] = append(v.byName[name], relPath)
		fileCount++
		return nil
	})

	if v.debug {
		fmt.Printf("[vault] scan complete: %d notes loaded\n", fileCount)
	}
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}
	return nil
}

// ReloadNote перезагружает одну заметку по относительному пути.
// После перезагрузки необходимо вызвать BuildLinks() для перестроения связей.
func (v *Vault) ReloadNote(relPath string) error {
	if v.debug {
		fmt.Printf("[vault] reloading %s\n", relPath)
	}
	relPath = filepath.ToSlash(relPath)
	absPath := filepath.Join(v.rootDir, filepath.FromSlash(relPath))
	n, err := note.Load(absPath, v.parser, v.builder)
	if err != nil {
		return fmt.Errorf("failed to reload %s: %w", relPath, err)
	}

	// Удаляем старую запись из byName (по точному значению)
	name := strings.ToLower(strings.TrimSuffix(filepath.Base(relPath), ".md"))
	v.byName[name] = removeString(v.byName[name], relPath)

	// Обновляем основное хранилище
	v.notes[relPath] = n

	// Добавляем новую запись в byName
	v.byName[name] = append(v.byName[name], relPath)
	if v.debug {
		fmt.Printf("[vault] reloaded %s\n", relPath)
	}
	return nil
}

// Note возвращает заметку по относительному пути (с .md). Второе значение — false, если не найдена.
func (v *Vault) Note(relPath string) (*note.Note, bool) {
	n, ok := v.notes[filepath.ToSlash(relPath)]
	return n, ok
}

// AllNotes возвращает срез всех загруженных заметок.
func (v *Vault) AllNotes() []*note.Note {
	notes := make([]*note.Note, 0, len(v.notes))
	for _, n := range v.notes {
		notes = append(notes, n)
	}
	return notes
}

// BuildLinks проходит по всем заметкам, очищает старые связи и строит новые.
// Этот метод заполняет у каждого сегмента OutgoingLinks и IncomingLinks.
func (v *Vault) BuildLinks() error {
	if v.debug {
		fmt.Println("[vault] building links...")

	}
	// 1. Очистить ссылки во всех сегментах
	for _, n := range v.notes {
		for _, seg := range n.Segments() {
			seg.OutgoingLinks = nil
			seg.IncomingLinks = nil
		}
	}

	// 2. Извлечь и разрешить ссылки
	for sourceRelPath, sourceNote := range v.notes {
		for _, seg := range sourceNote.Segments() {
			// Собираем текст из всех OwnBlocks (ссылки могут быть в любом блоке)
			var rawText strings.Builder
			if seg.Header != nil {
				rawText.WriteString(seg.Header.RawText)
				rawText.WriteString("\n")
			}
			for _, block := range seg.OwnBlocks {
				if block.Metadata.Type == note.TypeCode {
					continue
				}
				rawText.WriteString(block.RawText)
				rawText.WriteString("\n")
			}
			links, err := link.Extract(rawText.String())
			if err != nil {
				fmt.Printf("[vault] found link error in %s (%s): %v\n", sourceRelPath, seg.ID, err)
			}

			for _, l := range links {
				resolved := v.resolveLink(sourceRelPath, seg, l)
				seg.OutgoingLinks = append(seg.OutgoingLinks, resolved)

				// Если ссылка разрешена, добавляем обратную
				if resolved.LinkType != link.DeadLink {
					back := buildBackLink(resolved)
					targetSeg := v.findSegmentByID(resolved.TargetSegmentID)
					if targetSeg != nil {
						targetSeg.IncomingLinks = append(targetSeg.IncomingLinks, back)
					}
				}
			}
			if v.debug && len(links) > 0 {
				fmt.Printf("[vault] found %d links in %s (%s)\n", len(links), sourceRelPath, seg.ID)
			}
		}
	}

	if v.debug {
		fmt.Println("[vault] link building complete")
	}
	return nil
}

// resolveLink преобразует извлечённую ссылку в полную, подставляя Source* и разрешая Target*.
func (v *Vault) resolveLink(sourceRelPath string, sourceSeg *note.Segment, l link.Link) link.Link {
	l.SourceFile = sourceRelPath
	l.SourceSegmentID = sourceSeg.ID
	l.TargetFile = v.normalizeTargetFile(l.TargetFile, sourceRelPath)
	l.TargetSegmentID = ""

	// Проверяем существование целевого файла
	targetNote, found := v.notes[l.TargetFile]
	if !found {
		l.LinkType = link.DeadLink
		return l
	}

	// Если цель — сегмент (есть заголовок)
	if l.TargetHeading != "" {
		targetSeg := v.findSegmentByHeading(targetNote, l.TargetHeading)
		if targetSeg == nil {
			l.LinkType = link.DeadLink
		} else {
			l.TargetSegmentID = targetSeg.ID
			l.LinkType = link.Direct
		}
	} else {
		// Ссылка на заметку → целевой корневой сегмент
		rootSeg := targetNote.Tree.Root
		l.TargetSegmentID = rootSeg.ID
		l.LinkType = link.Direct
	}
	return l
}

// normalizeTargetFile приводит TargetFile к полному относительному пути (с .md) от корня Vault.
func (v *Vault) normalizeTargetFile(target string, sourceRelPath string) string {
	target = filepath.ToSlash(strings.TrimSpace(target))
	if target == "" {
		return ""
	}
	// Если уже содержит путь от корня (начинается с '/', './' или содержит '/'), оставляем как есть
	if strings.HasPrefix(target, "/") || strings.HasPrefix(target, "./") || strings.Contains(target, "/") {
		// добавляем .md при необходимости
		if !strings.HasSuffix(target, ".md") {
			target += ".md"
		}
		return strings.TrimPrefix(target, "./")
	}
	// Короткое имя — ищем в byName
	key := strings.ToLower(target)
	key = strings.TrimSuffix(key, ".md")
	candidates := v.byName[key]
	if len(candidates) == 1 {
		return candidates[0]
	}
	if len(candidates) > 1 {
		// Пытаемся найти в той же директории
		dir := filepath.ToSlash(filepath.Dir(sourceRelPath))
		for _, c := range candidates {
			if filepath.ToSlash(filepath.Dir(c)) == dir {
				return c
			}
		}
		// Если не нашли, возвращаем первый (с логированием можно расширить)
		return candidates[0]
	}
	// Не найдено — возвращаем имя с .md как «мёртвую» цель
	if !strings.HasSuffix(target, ".md") {
		target += ".md"
	}
	return target
}

// buildBackLink создаёт обратную ссылку на основе прямой.
func buildBackLink(l link.Link) link.Link {
	l.LinkType = link.BackLink
	return l
}

// findSegmentByHeading ищет сегмент с указанным заголовком (регистро-независимо).
func (v *Vault) findSegmentByHeading(n *note.Note, heading string) *note.Segment {
	heading = strings.ToLower(strings.TrimSpace(heading))
	for _, seg := range n.Segments() {
		if seg.Header == nil {
			continue
		}
		text := strings.ToLower(strings.TrimSpace(seg.HeadingText()))
		if text == heading {
			return seg
		}
	}
	return nil
}

// findSegmentByID ищет сегмент по ID среди всех заметок.
func (v *Vault) findSegmentByID(id string) *note.Segment {
	for _, n := range v.notes {
		for _, seg := range n.Segments() {
			if seg.ID == id {
				return seg
			}
		}
	}
	return nil
}

func removeString(sl []string, target string) []string {
	for i, s := range sl {
		if s == target {
			// Возвращаем копию, чтобы не мутировать исходный слайс
			result := make([]string, 0, len(sl)-1)
			result = append(result, sl[:i]...)
			result = append(result, sl[i+1:]...)
			return result
		}
	}
	return sl
}

// ChunkAll применяет настроенный чанкер ко всем сегментам всех заметок
// и возвращает общий список чанков.
func (v *Vault) ChunkAll() ([]note.Chunk, error) {
	if v.chunker == nil {
		return nil, fmt.Errorf("chunker is not set")
	}
	var all []note.Chunk
	for _, n := range v.notes {
		for _, seg := range n.Segments() {
			chunks, err := v.chunker.Chunk(seg)
			if err != nil {
				return nil, fmt.Errorf("chunk segment %s: %w", seg.ID, err)
			}
			all = append(all, chunks...)
		}
	}
	return all, nil
}
