package parser

import (
	"strings"

	"github.com/Galdoba/grohot/internal/domain/note"
)

// detectLineType returns the block type of a single line based on its syntax.
// Priority: heading > table > hr > list > callout > quote > paragraph.
func detectLineType(line string) note.BlockType {
	switch {
	case isHeading(line):
		return BlockTypeHeading
	case isTable(line):
		return BlockTypeTable
	case isHr(line):
		return BlockTypeHorizontalRule
	case isList(line):
		return BlockTypeList
	case isCallout(line):
		return BlockTypeCallout
	case isQuote(line):
		return BlockTypeQuote
	default:
		return BlockTypeParagraph
	}
}

// isHeading checks whether the line is an ATX heading (1–6 # characters followed by a space).
func isHeading(line string) bool {
	return headingRegex.MatchString(line)
}

// isHr checks if the line is a horizontal rule (---, ***, ___) and does not contain a pipe.
// The pipe exclusion prevents table lines from being misdetected.
func isHr(line string) bool {
	return !strings.Contains(line, "|") && hrRegex.MatchString(line)
}

// isTable checks if the line is a table row.
// A table row starts and ends with '|' (after trimming) and contains at least two pipe characters.
func isTable(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	return strings.HasPrefix(trimmed, "|") &&
		strings.HasSuffix(trimmed, "|") &&
		strings.Count(trimmed, "|") >= minTablePipeCount // at least two pipes
}

// isList returns true if the line starts with a Markdown list marker (-, *, +) or a numbered item.
func isList(line string) bool {
	return listRegex.MatchString(line)
}

// isCallout checks for an Obsidian callout marker (> [!...]).
func isCallout(line string) bool {
	return calloutRegex.MatchString(line)
}

// isQuote returns true if the line begins with '>' and is not a callout.
func isQuote(line string) bool {
	return quoteRegex.MatchString(line) && !isCallout(line)
}

// isEmpty reports whether the line consists only of whitespace.
func isEmpty(line string) bool {
	return strings.TrimSpace(line) == ""
}

// isCodeBlockStart returns true if the line is an opening fenced code block marker.
// It may include an optional language identifier (e.g., ```go).
func isCodeBlockStart(line string) bool {
	return codeBlockStartRegex.MatchString(line)
}

// isCodeBlockEnd returns true if the line is a closing fenced code block marker.
func isCodeBlockEnd(line string) bool {
	return codeBlockEndRegex.MatchString(line)
}