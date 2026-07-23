// Package obsidian implements a BlockParser for Obsidian-flavored Markdown files.
// It performs the first pass: splitting lines into typed blocks (code, heading, table, etc.)
// without building hierarchy. Hierarchy is the responsibility of the second pass.
package parser

import (
	"strings"

	"github.com/Galdoba/grohot/internal/domain/note"
)

// Parser is a stateless implementation of note.BlockParser for Obsidian syntax.
type Parser struct{}

// NewParser creates a new Obsidian parser instance.
func NewParser() *Parser {
	return &Parser{}
}

// Parse implements note.BlockParser.
// Each non-empty line normally starts a new block; consecutive lines of the same type
// are merged until separated by an empty line. Code blocks (```) are collected verbatim.
func (p *Parser) Parse(lines []string) ([]note.ContentBlock, error) {
	state := &parseState{
		mode:   ModeNormal,
		buffer: []string{},
		blocks: []note.ContentBlock{},
	}

	for _, line := range lines {
		p.processLine(line, state)
	}

	// Flush an unclosed code block if it was the last element.
	if len(state.buffer) > 0 && state.mode == ModeMarker {
		p.emitCodeBlock(state)
	}

	return state.blocks, nil
}

// parseState holds the parser’s mutable state during parsing.
type parseState struct {
	mode        parseMode
	buffer      []string
	blocks      []note.ContentBlock
	blockClosed bool
	lastType    note.BlockType // type of the last processed line (used for callout context correction)
}

type parseMode int

const (
	ModeNormal parseMode = iota
	ModeMarker
)

// processLine dispatches the line to the handler corresponding to the current mode.
func (p *Parser) processLine(line string, state *parseState) {
	switch state.mode {
	case ModeNormal:
		p.processNormal(line, state)
	case ModeMarker:
		p.processMarker(line, state)
	}
}

// processNormal handles a line when the parser is in normal (non-code-block) mode.
func (p *Parser) processNormal(line string, state *parseState) {
	// Enter code block mode if the line opens a fenced code block.
	if isCodeBlockStart(line) {
		state.mode = ModeMarker
		state.buffer = []string{line}
		state.lastType = "" // reset context because we switch mode
		return
	}

	// Empty lines close the current block and reset context.
	if isEmpty(line) {
		p.closeCurrentBlock(state)
		return
	}

	detectedType := detectLineType(line)

	// Obsidian callout: once a callout block starts, subsequent lines beginning with '>'
	// are continuations of that callout, not separate quotes.
	if detectedType == BlockTypeQuote && state.lastType == BlockTypeCallout {
		detectedType = BlockTypeCallout
	}

	p.emitBlockLine(state, line, detectedType)
	state.lastType = detectedType
}

// closeCurrentBlock marks the current block as finished and resets the type context.
func (p *Parser) closeCurrentBlock(state *parseState) {
	state.blockClosed = true
	state.lastType = ""
}

// emitBlockLine appends a line to the active block. If the block type differs or the
// previous block was closed, a new block is created. Otherwise the line is appended
// to the existing block’s RawText with a newline separator.
func (p *Parser) emitBlockLine(state *parseState, line string, blockType note.BlockType) {
	needNewBlock := state.blockClosed ||
		len(state.blocks) == 0 ||
		state.blocks[len(state.blocks)-1].Metadata.Type != blockType

	if needNewBlock {
		state.blocks = append(state.blocks, note.ContentBlock{
			RawText: line,
			Metadata: note.BlockMetadata{
				Type: blockType,
			},
		})
		state.blockClosed = false
		return
	}

	// Append to the last block.
	last := &state.blocks[len(state.blocks)-1]
	if last.RawText == "" {
		last.RawText = line
	} else {
		last.RawText += "\n" + line
	}
}

// processMarker handles a line when the parser is inside a fenced code block.
func (p *Parser) processMarker(line string, state *parseState) {
	state.buffer = append(state.buffer, line)

	// Close the code block only if the current line is a closing fence and we have
	// more than just the opening fence in the buffer (len > 1).
	if isCodeBlockEnd(line) && len(state.buffer) > 1 {
		p.emitCodeBlock(state)
		state.mode = ModeNormal
		state.lastType = "" // reset to avoid merging with subsequent content
	}
}

// emitCodeBlock creates a "code" block from the accumulated buffer and clears it.
func (p *Parser) emitCodeBlock(state *parseState) {
	if len(state.buffer) == 0 {
		return
	}
	state.blocks = append(state.blocks, note.ContentBlock{
		RawText: strings.Join(state.buffer, "\n"),
		Metadata: note.BlockMetadata{
			Type: BlockTypeCode,
		},
	})
	state.buffer = []string{}
}
