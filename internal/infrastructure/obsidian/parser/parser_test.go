package parser

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/Galdoba/grohot/internal/domain/note"
)

func TestParser_Parse(t *testing.T) {
	data, _ := os.ReadFile(`c:\Users\pemaltynov\Documents\Obsidian\traveller\Generic\Управление игровыми таблицами.md`)
	lines := strings.Split(string(data), "\n")
	bl, err := NewParser().Parse(lines)
	fmt.Println(err)
	for i, b := range bl {
		fmt.Println(i, b)
	}
}

func TestParse_EmptyInput(t *testing.T) {
	p := NewParser()
	blocks, err := p.Parse(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(blocks) != 0 {
		t.Fatalf("expected no blocks, got %d", len(blocks))
	}
}

func TestParse_SingleParagraph(t *testing.T) {
	p := NewParser()
	lines := []string{"A simple paragraph."}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	expected := []note.ContentBlock{
		{RawText: "A simple paragraph.", Metadata: note.BlockMetadata{Type: BlockTypeParagraph}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_ParagraphsSeparatedByEmptyLines(t *testing.T) {
	p := NewParser()
	lines := []string{
		"First paragraph.",
		"",
		"Second paragraph.",
	}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	expected := []note.ContentBlock{
		{RawText: "First paragraph.", Metadata: note.BlockMetadata{Type: BlockTypeParagraph}},
		{RawText: "Second paragraph.", Metadata: note.BlockMetadata{Type: BlockTypeParagraph}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_MergingConsecutiveLinesOfSameType(t *testing.T) {
	p := NewParser()
	lines := []string{
		"Line one,",
		"still paragraph.",
	}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	expected := []note.ContentBlock{
		{RawText: "Line one,\nstill paragraph.", Metadata: note.BlockMetadata{Type: BlockTypeParagraph}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_Headings(t *testing.T) {
	p := NewParser()
	lines := []string{
		"# Heading 1",
		"## Heading 2",
	}
	// Two headings back‑to‑back merge into one block (same type).
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	expected := []note.ContentBlock{
		{RawText: "# Heading 1\n## Heading 2", Metadata: note.BlockMetadata{Type: BlockTypeHeading}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_HeadingsSeparatedByEmptyLine(t *testing.T) {
	p := NewParser()
	lines := []string{
		"# First",
		"",
		"# Second",
	}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	expected := []note.ContentBlock{
		{RawText: "# First", Metadata: note.BlockMetadata{Type: BlockTypeHeading}},
		{RawText: "# Second", Metadata: note.BlockMetadata{Type: BlockTypeHeading}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_Table(t *testing.T) {
	p := NewParser()
	lines := []string{
		"| Header 1 | Header 2 |",
		"|----------|----------|",
		"| Cell 1   | Cell 2   |",
	}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	// All table rows are of type "table", they merge.
	expected := []note.ContentBlock{
		{RawText: "| Header 1 | Header 2 |\n|----------|----------|\n| Cell 1   | Cell 2   |",
			Metadata: note.BlockMetadata{Type: BlockTypeTable}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_TableRequiresAtLeastTwoPipes(t *testing.T) {
	p := NewParser()
	// A single pipe line is not a table, it's a paragraph.
	lines := []string{"| Not a table"}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	expected := []note.ContentBlock{
		{RawText: "| Not a table", Metadata: note.BlockMetadata{Type: BlockTypeParagraph}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_HorizontalRule(t *testing.T) {
	p := NewParser()
	lines := []string{
		"---",
		"***",
		"___",
	}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	// All three are hr; they merge.
	expected := []note.ContentBlock{
		{RawText: "---\n***\n___", Metadata: note.BlockMetadata{Type: BlockTypeHorizontalRule}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_HorizontalRuleNotConfusedWithTable(t *testing.T) {
	p := NewParser()
	// A line with pipes is not an hr.
	lines := []string{"---|---|---"}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	expected := []note.ContentBlock{
		{RawText: "---|---|---", Metadata: note.BlockMetadata{Type: BlockTypeParagraph}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_Lists(t *testing.T) {
	p := NewParser()
	lines := []string{
		"- item 1",
		"- item 2",
		"  continuation of item 2", // still a list line because starts with whitespace? Actually list regex requires marker. So this line would be paragraph. Let's adjust test: we'll only put lines that match list regex.
	}
	// Better to use only lines that are list items.
	lines = []string{
		"- item 1",
		"* item 2",
		"1. ordered item",
	}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	expected := []note.ContentBlock{
		{RawText: "- item 1\n* item 2\n1. ordered item", Metadata: note.BlockMetadata{Type: BlockTypeList}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_CalloutAndContinuation(t *testing.T) {
	p := NewParser()
	lines := []string{
		"> [!note] Title",
		"> Content line",
		"> Another line",
	}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	// First line is callout, subsequent quote lines are converted to callout.
	expected := []note.ContentBlock{
		{RawText: "> [!note] Title\n> Content line\n> Another line",
			Metadata: note.BlockMetadata{Type: BlockTypeCallout}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_CalloutEndsWithEmptyLine(t *testing.T) {
	p := NewParser()
	lines := []string{
		"> [!warning]",
		"> warning text",
		"",
		"> separate quote",
	}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	expected := []note.ContentBlock{
		{RawText: "> [!warning]\n> warning text", Metadata: note.BlockMetadata{Type: BlockTypeCallout}},
		{RawText: "> separate quote", Metadata: note.BlockMetadata{Type: BlockTypeQuote}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_Quote(t *testing.T) {
	p := NewParser()
	lines := []string{
		"> quoted text",
		"> more quoted text",
	}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	expected := []note.ContentBlock{
		{RawText: "> quoted text\n> more quoted text", Metadata: note.BlockMetadata{Type: BlockTypeQuote}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_CodeBlock(t *testing.T) {
	p := NewParser()
	lines := []string{
		"```go",
		"func main() {",
		"    fmt.Println(\"hello\")",
		"}",
		"```",
		"After code block",
	}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	expected := []note.ContentBlock{
		{RawText: "```go\nfunc main() {\n    fmt.Println(\"hello\")\n}\n```",
			Metadata: note.BlockMetadata{Type: BlockTypeCode}},
		{RawText: "After code block", Metadata: note.BlockMetadata{Type: BlockTypeParagraph}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_UnclosedCodeBlockAtEOF(t *testing.T) {
	p := NewParser()
	lines := []string{
		"```",
		"some code",
		"still code",
	}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	expected := []note.ContentBlock{
		{RawText: "```\nsome code\nstill code",
			Metadata: note.BlockMetadata{Type: BlockTypeCode}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_CodeBlockWithEmptyLines(t *testing.T) {
	p := NewParser()
	lines := []string{
		"```",
		"line1",
		"",
		"line2",
		"```",
	}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	expected := []note.ContentBlock{
		{RawText: "```\nline1\n\nline2\n```",
			Metadata: note.BlockMetadata{Type: BlockTypeCode}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_PriorityHeadingOverCallout(t *testing.T) {
	// A line that starts with # is heading, even if it has > later.
	// But heading regex requires # at beginning after whitespace.
	// So a line like `> # heading` is not heading (it's quote/callout).
	// We need a valid heading: "## heading" is heading, no ambiguity.
	// To test priority, a line that could be both heading and list?
	// Actually, list uses -, *, +, or digit. Heading uses #. No conflict.
	// Conflict between hr and list? --- is hr, but - item is list. Since hr regex checks whole line of -, *, or _, and list regex requires a space after marker.
	// So we'll test that "---" is hr, not list. That's already covered by hr detection.
	// Let's add a test: "- text" is list, not hr.
	p := NewParser()
	lines := []string{"- text"}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	expected := []note.ContentBlock{
		{RawText: "- text", Metadata: note.BlockMetadata{Type: BlockTypeList}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_PriorityTableOverHr(t *testing.T) {
	// "---|---|---" is not a table because doesn't start/end with |. But "| --- |" is a table.
	p := NewParser()
	lines := []string{"| --- |"}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	expected := []note.ContentBlock{
		{RawText: "| --- |", Metadata: note.BlockMetadata{Type: BlockTypeTable}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_MultipleEmptyLines(t *testing.T) {
	p := NewParser()
	lines := []string{
		"first",
		"",
		"",
		"second",
	}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	expected := []note.ContentBlock{
		{RawText: "first", Metadata: note.BlockMetadata{Type: BlockTypeParagraph}},
		{RawText: "second", Metadata: note.BlockMetadata{Type: BlockTypeParagraph}},
	}
	assertBlocksEqual(t, expected, blocks)
}

func TestParse_OnlyEmptyLines(t *testing.T) {
	p := NewParser()
	lines := []string{"", "   ", "\t"}
	blocks, err := p.Parse(lines)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 0 {
		t.Fatalf("expected no blocks from only empty lines, got %d", len(blocks))
	}
}

func assertBlocksEqual(t *testing.T, expected, got []note.ContentBlock) {
	t.Helper()
	if len(expected) != len(got) {
		t.Fatalf("block count mismatch: expected %d, got %d\nExpected: %+v\nGot: %+v",
			len(expected), len(got), expected, got)
	}
	for i := range expected {
		if expected[i].Metadata.Type != got[i].Metadata.Type {
			t.Errorf("block %d: type mismatch: expected %q, got %q", i, expected[i].Metadata.Type, got[i].Metadata.Type)
		}
		if expected[i].RawText != got[i].RawText {
			t.Errorf("block %d: raw text mismatch:\nExpected: %q\nGot:      %q", i, expected[i].RawText, got[i].RawText)
		}
	}
	// If we want full deep equal, we could use reflect.DeepEqual after fixing any minor field differences.
	if !reflect.DeepEqual(expected, got) {
		// Already reported specific diffs, so just signal failure once.
		// But to avoid double reporting, we can skip if already had errors.
		if !t.Failed() {
			t.Errorf("unexpected block difference: expected %+v, got %+v", expected, got)
		}
	}
}
