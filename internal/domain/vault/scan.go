package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Galdoba/grohot/internal/domain/note"
)

const (
	markdownExtension = ".md"
	dotSlashPrefix    = "./"
	pathSeparator     = "/"
)

// ScanAndLoad recursively scans all .md files in the vault and loads them.
// Previous data is cleared.
func (v *Vault) ScanAndLoad() error {
	if v.debug {
		fmt.Println("[vault] scanning .md files...")
	}
	v.notes = make(map[string]*note.Note)
	v.byName = make(map[string][]string)

	fileCount := 0
	walkFn := func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			if isHiddenDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !isMarkdownFile(path) {
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
		relPath = filepath.ToSlash(relPath)

		if v.debug {
			fmt.Printf("[vault] loading %s\n", relPath)
		}
		if err := v.addNoteFromFile(absPath, relPath); err != nil {
			return err
		}
		fileCount++
		return nil
	}

	err := filepath.Walk(v.rootDir, walkFn)
	if v.debug {
		fmt.Printf("[vault] scan complete: %d notes loaded\n", fileCount)
	}
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}
	return nil
}

// ReloadNote reloads a single note by relative path.
// After reloading, BuildLinks must be called to rebuild links.
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

	name := noteBaseName(relPath)
	v.byName[name] = removeString(v.byName[name], relPath)
	v.notes[relPath] = n
	v.byName[name] = append(v.byName[name], relPath)

	if v.debug {
		fmt.Printf("[vault] reloaded %s\n", relPath)
	}
	return nil
}

// addNoteFromFile loads a note from disk and adds it to the vault indexes.
func (v *Vault) addNoteFromFile(absPath, relPath string) error {
	n, err := note.Load(absPath, v.parser, v.builder)
	if err != nil {
		return fmt.Errorf("failed to load %s: %w", relPath, err)
	}
	v.notes[relPath] = n
	name := noteBaseName(relPath)
	v.byName[name] = append(v.byName[name], relPath)
	return nil
}

// isHiddenDir reports whether the directory name is hidden (starts with a dot and is not ".").
func isHiddenDir(name string) bool {
	return strings.HasPrefix(name, ".") && name != "."
}

// isMarkdownFile reports whether the file has a .md extension.
func isMarkdownFile(path string) bool {
	return strings.EqualFold(filepath.Ext(path), markdownExtension)
}

// noteBaseName returns the lowercase base filename without extension.
func noteBaseName(relPath string) string {
	base := strings.ToLower(filepath.Base(relPath))
	return strings.TrimSuffix(base, markdownExtension)
}
