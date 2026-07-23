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
	Name        string
	Filepath    string
	Frontmatter *frontmatter.Frontmatter
	Blocks      []ContentBlock
}

// New creates a Note from already parsed data (useful for testing or manual construction).
func New(name, filepath string, fm *frontmatter.Frontmatter, blocks []ContentBlock) (*Note, error) {
	if name == "" {
		return nil, fmt.Errorf("note name cannot be empty")
	}
	if filepath == "" {
		return nil, fmt.Errorf("filepath cannot be empty")
	}
	n := &Note{
		Name:        name,
		Filepath:    filepath,
		Frontmatter: fm,
		Blocks:      blocks,
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
	lines := strings.Split(string(data), "\n")

	fm, restLines, err := frontmatter.Parse(lines)
	if err != nil {
		return nil, fmt.Errorf("frontmatter parsing failed: %w", err)
	}

	// First pass: raw blocks from the body
	blocks, err := parser.Parse(restLines)
	if err != nil {
		return nil, fmt.Errorf("body parsing failed: %w", err)
	}

	// Create note without hierarchy
	note, err := New(
		strings.TrimSuffix(filepath.Base(path), ".md"),
		path,
		fm,
		blocks,
	)
	if err != nil {
		return nil, err
	}

	// Second pass: build hierarchy (Path, Sequence, Depth, Index)
	noteWithHierarchy, err := note.BuildHierarchy(builder)
	if err != nil {
		return nil, fmt.Errorf("hierarchy build failed: %w", err)
	}

	return noteWithHierarchy, nil
}

// BuildHierarchy applies the HierarchyBuilder to the note's blocks,
// returning a new Note with updated metadata.
func (n *Note) BuildHierarchy(builder HierarchyBuilder) (*Note, error) {
	updatedBlocks, err := builder.Build(n.Blocks)
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
