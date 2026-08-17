package vault

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Galdoba/grohot/internal/domain/note"
)

const (
	obsidianDirName   = ".obsidian"
	obsidianTypesFile = "types.json"
	obsidianGraphFile = "graph.json"
)

// Dependencies holds the external collaborators required by Vault.
type Dependencies struct {
	Parser  note.BlockParser
	Builder note.HierarchyBuilder
	Chunker note.Chunker
}

// Vault represents an Obsidian vault.
type Vault struct {
	rootDir string
	notes   map[string]*note.Note // key: relative path (with .md), normalized with '/'
	byName  map[string][]string   // key: lowercase base filename without .md -> list of relative paths
	parser  note.BlockParser
	builder note.HierarchyBuilder
	chunker note.Chunker
	debug   bool
}

// NewVault creates a new Vault for the given root directory.
// It returns an error if rootDir is not an Obsidian vault root.
func NewVault(rootDir string, deps Dependencies) (*Vault, error) {
	if !isVaultRoot(rootDir) {
		return nil, fmt.Errorf("%s is not an Obsidian vault root", rootDir)
	}
	return &Vault{
		rootDir: rootDir,
		notes:   make(map[string]*note.Note),
		byName:  make(map[string][]string),
		parser:  deps.Parser,
		builder: deps.Builder,
		chunker: deps.Chunker,
		debug:   true,
	}, nil
}

// isVaultRoot checks that dir is the root of an Obsidian vault.
// It verifies the presence of .obsidian directory and required files inside it.
func isVaultRoot(dir string) bool {
	obsidianDir := filepath.Join(dir, obsidianDirName)
	info, err := os.Stat(obsidianDir)
	if err != nil || !info.IsDir() {
		return false
	}
	if !fileExists(filepath.Join(obsidianDir, obsidianTypesFile)) {
		return false
	}
	if !fileExists(filepath.Join(obsidianDir, obsidianGraphFile)) {
		return false
	}
	return true
}

// fileExists reports whether the file at path exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FindVaultRoot climbs up from startPath until it finds a directory containing .obsidian.
// It returns the vault root directory or an error if not found.
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

// Note returns the note for the given relative path (with .md).
// The second return value indicates whether the note was found.
func (v *Vault) Note(relPath string) (*note.Note, bool) {
	n, ok := v.notes[filepath.ToSlash(relPath)]
	return n, ok
}

// AllNotes returns a slice of all loaded notes.
func (v *Vault) AllNotes() []*note.Note {
	notes := make([]*note.Note, 0, len(v.notes))
	for _, n := range v.notes {
		notes = append(notes, n)
	}
	return notes
}