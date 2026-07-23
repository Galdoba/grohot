// File: hierarchy/builder.go
package hierarchy

import (
	"fmt"

	"github.com/Galdoba/grohot/internal/domain/note"
)

const defaultSeparator = " > "

// Builder implements note.HierarchyBuilder, filling metadata fields
// (Path, Depth, Sequence, Index, Filepath) based on heading structure.
type Builder struct {
	separator string
}

// NewBuilder creates a Builder with the default path separator " > ".
func NewBuilder() *Builder {
	return &Builder{separator: defaultSeparator}
}

// WithSeparator allows configuring a custom path separator (e.g. " / ").
func (b *Builder) WithSeparator(sep string) *Builder {
	b.separator = sep
	return b
}

// Build processes a slice of ContentBlocks and returns a new slice where each block
// has its hierarchy metadata populated. The original blocks are not modified.
func (b *Builder) Build(blocks []note.ContentBlock, filepath string) ([]note.ContentBlock, error) {
	if err := validateInputs(blocks, filepath); err != nil {
		return nil, err
	}

	state := newBuildState(b.separator)
	result := make([]note.ContentBlock, len(blocks))

	for i, block := range blocks {
		filled, err := state.processBlock(block, i, filepath)
		if err != nil {
			return nil, fmt.Errorf("block %d: %w", i, err)
		}
		result[i] = filled
	}

	return result, nil
}

// newBuildState initialises the mutable build state.
func newBuildState(separator string) *buildState {
	return &buildState{
		stack:       []string{},
		pathCounter: make(map[string]int),
		separator:   separator,
	}
}

// validateInputs checks the preconditions for Build.
func validateInputs(blocks []note.ContentBlock, filepath string) error {
	if blocks == nil {
		return fmt.Errorf("blocks slice cannot be nil")
	}
	if filepath == "" {
		return fmt.Errorf("filepath cannot be empty")
	}
	return nil
}