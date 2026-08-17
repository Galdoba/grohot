package note

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Galdoba/grohot/internal/domain/note/frontmatter"
)

// Note is the aggregate root representing a single note.
type Note struct {
	Name        string                   // human‑readable note name (without extension)
	Filepath    string                   // absolute or relative path to the note file
	Frontmatter *frontmatter.Frontmatter // parsed frontmatter properties
	Blocks      []ContentBlock           // flat list of content blocks with hierarchy metadata
	Tree        SegmentTree              // hierarchical segment tree built from Blocks
}

// NoteConfig holds the parameters for constructing a Note.
// It is used to keep the argument count of New within the limit.
type NoteConfig struct {
	Name        string
	Filepath    string
	Frontmatter *frontmatter.Frontmatter
	Blocks      []ContentBlock
}

// New creates a Note from the provided configuration.
// It validates the resulting note before returning.
func New(cfg NoteConfig) (*Note, error) {
	if strings.TrimSpace(cfg.Name) == "" {
		return nil, fmt.Errorf("note name cannot be empty")
	}
	if cfg.Filepath == "" {
		return nil, fmt.Errorf("filepath cannot be empty")
	}
	n := &Note{
		Name:        cfg.Name,
		Filepath:    cfg.Filepath,
		Frontmatter: cfg.Frontmatter,
		Blocks:      cfg.Blocks,
		Tree:        SegmentTree{Root: &Segment{}}, // ensure Tree is never nil
	}
	if err := n.Validate(); err != nil {
		return nil, err
	}
	return n, nil
}

// Load reads a file, parses frontmatter, parses the body into blocks,
// builds the hierarchy, and populates the segment tree.
// This is the recommended entry point for creating a fully populated Note.
func Load(path string, parser BlockParser, builder HierarchyBuilder) (*Note, error) {
	lines, err := readNoteFile(path)
	if err != nil {
		return nil, err
	}

	fm, restLines, err := frontmatter.Parse(lines)
	if err != nil {
		return nil, fmt.Errorf("frontmatter parsing failed: %w", err)
	}

	blocks, err := parser.Parse(restLines)
	if err != nil {
		return nil, fmt.Errorf("body parsing failed: %w", err)
	}

	note, err := New(NoteConfig{
		Name:        strings.TrimSuffix(filepath.Base(path), ".md"),
		Filepath:    path,
		Frontmatter: fm,
		Blocks:      blocks,
	})
	if err != nil {
		return nil, err
	}

	noteWithHierarchy, err := note.BuildHierarchy(builder)
	if err != nil {
		return nil, fmt.Errorf("hierarchy build failed: %w", err)
	}

	noteWithHierarchy.Tree = CreateSegmentsTree(noteWithHierarchy.Blocks)
	noteWithHierarchy.Tree.PopulateMetadata(path, noteWithHierarchy.Frontmatter)

	return noteWithHierarchy, nil
}

// readNoteFile reads the file at path, normalizes line endings to "\n",
// and returns the lines as a slice of strings.
func readNoteFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	normalized := strings.ReplaceAll(string(data), "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	return strings.Split(normalized, "\n"), nil
}

// BuildHierarchy applies the HierarchyBuilder to the note's blocks,
// returning a new Note with updated metadata.
func (n *Note) BuildHierarchy(builder HierarchyBuilder) (*Note, error) {
	updatedBlocks, err := builder.Build(n.Blocks, n.Filepath)
	if err != nil {
		return nil, err
	}
	return &Note{
		Name:        n.Name,
		Filepath:    n.Filepath,
		Frontmatter: n.Frontmatter,
		Blocks:      updatedBlocks,
		Tree:        SegmentTree{Root: &Segment{}},
	}, nil
}

// Segments returns all segments of the note as a flat slice.
// It is safe to call even if the segment tree has not been built.
func (n *Note) Segments() []*Segment {
	if n.Tree.Root == nil {
		return nil
	}
	return n.Tree.Flatten()
}
