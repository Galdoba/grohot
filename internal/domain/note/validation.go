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
		meta := block.Metadata
		if meta.Filepath != "" && meta.Filepath != n.Filepath {
			return fmt.Errorf("block at index %d has mismatched filepath: %s (expected %s)",
				i, meta.Filepath, n.Filepath)
		}
		if !isValidType(meta.Type) {
			return fmt.Errorf("block at index %d has invalid type: %q", i, meta.Type)
		}
		if meta.Path != "" && meta.Sequence > 0 {
			id := GenerateBlockID(meta)
			if seenIDs[id] {
				return fmt.Errorf("duplicate block ID: %s", id)
			}
			seenIDs[id] = true
		}
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
