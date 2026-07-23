package parser

import (
	"regexp"

	"github.com/Galdoba/grohot/internal/domain/note"
)

// Block type constants – use these instead of raw strings to avoid typos and improve clarity.
// These values mirror the domain's BlockType constants; they are kept here to make the package
// self‑contained and to avoid import cycles. When domain constants are available, prefer
// referencing them directly.
const (
	BlockTypeHeading        note.BlockType = "heading"
	BlockTypeTable          note.BlockType = "table"
	BlockTypeHorizontalRule note.BlockType = "hr"
	BlockTypeList           note.BlockType = "list"
	BlockTypeCallout        note.BlockType = "callout"
	BlockTypeQuote          note.BlockType = "quote"
	BlockTypeParagraph      note.BlockType = "paragraph"
	BlockTypeCode           note.BlockType = "code"
)

// Internal constants to eliminate magic numbers.
const minTablePipeCount = 2 // a table row requires at least two pipe symbols

// Precompiled regular expressions.
// All are immutable and safe for concurrent reads.
var (
	// headingRegex matches ATX headings (1 to 6 '#' followed by a space).
	// Immutable: safe for concurrent read access.
	headingRegex = regexp.MustCompile(`^[ \t]*(#{1,6}) +`)

	// hrRegex matches horizontal rules (---, ***, ___) with optional surrounding whitespace.
	hrRegex = regexp.MustCompile(`^[ \t]*(?:---+|\*\*\*+|___+)[ \t]*$`)

	// listRegex matches unordered (-, *, +) and ordered (1.) list markers.
	listRegex = regexp.MustCompile(`^[ \t]*([-*+]|\d+\.)[ \t]+`)

	// calloutRegex matches Obsidian callout markers (> [!<type>]).
	calloutRegex = regexp.MustCompile(`^[ \t]*>[ \t]*\[![a-zA-Z]`)

	// quoteRegex matches lines starting with '>' (after optional whitespace).
	quoteRegex = regexp.MustCompile(`^[ \t]*>[ \t]*`)

	// codeBlockStartRegex matches an opening fenced code block with an optional language tag.
	codeBlockStartRegex = regexp.MustCompile(`^[ \t]*` + "```" + `[ \t]*[a-zA-Z0-9_\-]*[ \t]*$`)

	// codeBlockEndRegex matches a closing fenced code block (exactly three backticks and optional whitespace).
	codeBlockEndRegex = regexp.MustCompile(`^[ \t]*` + "```" + `[ \t]*$`)
)