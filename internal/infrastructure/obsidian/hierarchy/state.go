// File: hierarchy/state.go
package hierarchy

import (
	"fmt"
	"strings"

	"github.com/Galdoba/grohot/internal/domain/note"
)

// buildState holds the mutable data needed while processing blocks sequentially.
type buildState struct {
	stack       []string          // current heading stack (texts without #)
	pathCounter map[string]int    // sequence number per path
	separator   string            // path separator (e.g. " > ")
}

// processBlock takes a raw block (from the first pass) and fills its hierarchy metadata
// based on the current heading stack. It returns a new fully-populated block.
func (s *buildState) processBlock(block note.ContentBlock, index int, filepath string) (note.ContentBlock, error) {
	out := block
	out.Metadata.Filepath = filepath
	out.Metadata.Index = index

	if block.Metadata.Type == note.TypeHeading {
		if err := s.handleHeading(&out); err != nil {
			return note.ContentBlock{}, err
		}
	} else {
		s.handleNonHeading(&out)
	}

	s.assignPathAndSequence(&out)
	return out, nil
}

// handleHeading extracts heading level and text, updates the stack, and sets Depth.
func (s *buildState) handleHeading(block *note.ContentBlock) error {
	level, text, err := extractHeading(block.RawText)
	if err != nil {
		return fmt.Errorf("invalid heading %q: %w", block.RawText, err)
	}

	// Trim the stack to keep only the first level-1 elements (parents).
	// If the heading jumps more than one level deeper (e.g. h2 -> h4), the missing
	// intermediate levels are simply absent; the stack is truncated to what exists.
	if level-1 < len(s.stack) {
		s.stack = s.stack[:level-1]
	}
	// Append the current heading text.
	s.stack = append(s.stack, text)
	block.Metadata.Depth = len(s.stack) - 1 // number of ancestors
	return nil
}

// handleNonHeading sets Depth and Path based on the current heading stack.
func (s *buildState) handleNonHeading(block *note.ContentBlock) {
	block.Metadata.Depth = len(s.stack) // number of parent headings
	block.Metadata.Path = strings.Join(s.stack, s.separator)
}

// assignPathAndSequence computes the block's Path (for headings this is the new full path)
// and assigns a monotonically increasing Sequence within that path.
func (s *buildState) assignPathAndSequence(block *note.ContentBlock) {
	path := strings.Join(s.stack, s.separator)
	block.Metadata.Path = path

	s.pathCounter[path]++
	block.Metadata.Sequence = s.pathCounter[path]
}