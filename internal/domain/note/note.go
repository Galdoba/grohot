package note

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Galdoba/grohot/internal/domain/note/frontmatter"
)

// Note is the aggregate root representing a single note.
type Note struct {
	Name        string
	Filepath    string
	Frontmatter *frontmatter.Frontmatter
	Blocks      []ContentBlock
	Tree        SegmentTree
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
	}
	if err := n.Validate(); err != nil {
		return nil, err
	}
	return n, nil
}

// Load reads a file, parses frontmatter, parses the body into blocks,
// and builds the hierarchy. Returns a fully populated Note.
// This is the recommended entry point.
func Load(path string, parser BlockParser, builder HierarchyBuilder) (*Note, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	data = bytes.ReplaceAll(data, []byte("\r"), []byte("\n"))
	lines := strings.Split(string(data), "\n")

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
	}, nil
}

// Segments returns all segments of the note as a flat slice.
func (n *Note) Segments() []*Segment {
	return n.Tree.Flatten()
}
