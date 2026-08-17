package note

import "fmt"

// Validate checks the note's invariants.
func (n *Note) Validate() error {
	if n.Name == "" {
		return fmt.Errorf("note name cannot be empty")
	}
	if n.Filepath == "" {
		return fmt.Errorf("filepath cannot be empty")
	}
	seenIDs := make(map[string]bool)
	for i, block := range n.Blocks {
		if err := validateBlock(block, n.Filepath, seenIDs); err != nil {
			return fmt.Errorf("block at index %d: %w", i, err)
		}
	}
	return nil
}

// validateBlock checks a single block for consistency with the note.
// It returns an error if the block is invalid or has a duplicate ID.
func validateBlock(block ContentBlock, noteFilepath string, seenIDs map[string]bool) error {
	meta := block.Metadata
	if meta.Filepath != "" && meta.Filepath != noteFilepath {
		return fmt.Errorf("mismatched filepath: %s (expected %s)", meta.Filepath, noteFilepath)
	}
	if !isValidType(meta.Type) {
		return fmt.Errorf("invalid type: %q", meta.Type)
	}
	if meta.Path != "" && meta.Sequence > 0 {
		id := GenerateBlockID(meta)
		if seenIDs[id] {
			return fmt.Errorf("duplicate block ID: %s", id)
		}
		seenIDs[id] = true
	}
	return nil
}

// isValidType checks if the type is one of the defined constants.
func isValidType(typ BlockType) bool {
	switch typ {
	case TypeHeading, TypeParagraph, TypeList, TypeCode, TypeTable, TypeQuote, TypeCallout, TypeHr:
		return true
	default:
		return false
	}
}
