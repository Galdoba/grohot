// File: hierarchy/builder_test.go
package hierarchy_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Galdoba/grohot/internal/domain/note"
	"github.com/Galdoba/grohot/internal/infrastructure/obsidian/hierarchy"
	"github.com/Galdoba/grohot/internal/infrastructure/obsidian/parser"
)

func Test_Live(t *testing.T) {
	path := `c:\Users\pemaltynov\Documents\Obsidian\traveller\Generic\Управление игровыми таблицами.md`
	data, _ := os.ReadFile(path)
	lines := strings.Split(string(data), "\n")
	bl, err := parser.NewParser().Parse(lines)
	fmt.Println(err)
	for i, b := range bl {
		fmt.Println(i, b)
	}
	hBl, err2 := hierarchy.NewBuilder().Build(bl, path)
	fmt.Println(err2)
	for i, b := range hBl {
		fmt.Println(i, b.Metadata)
	}
}

func buildHeading(raw string) note.ContentBlock {
	return note.ContentBlock{
		RawText: raw,
		Metadata: note.BlockMetadata{
			Type: note.TypeHeading,
		},
	}
}

func buildParagraph(raw string) note.ContentBlock {
	return note.ContentBlock{
		RawText: raw,
		Metadata: note.BlockMetadata{
			Type: note.TypeParagraph,
		},
	}
}

func TestBuilder_EmptyBlocks(t *testing.T) {
	b := hierarchy.NewBuilder()
	_, err := b.Build(nil, "test.md")
	if err == nil {
		t.Fatal("expected error for nil blocks")
	}
	_, err = b.Build([]note.ContentBlock{}, "")
	if err == nil {
		t.Fatal("expected error for empty filepath")
	}
}

func TestBuilder_NoHeadings(t *testing.T) {
	b := hierarchy.NewBuilder()
	blocks := []note.ContentBlock{
		buildParagraph("intro"),
		buildParagraph("body"),
	}
	result, err := b.Build(blocks, "note.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(result))
	}
	// All blocks belong to the root path (empty string).
	for i, blk := range result {
		if blk.Metadata.Path != "" {
			t.Errorf("block %d: expected empty path, got %q", i, blk.Metadata.Path)
		}
		if blk.Metadata.Depth != 0 {
			t.Errorf("block %d: depth should be 0, got %d", i, blk.Metadata.Depth)
		}
		if blk.Metadata.Index != i {
			t.Errorf("block %d: index expected %d, got %d", i, i, blk.Metadata.Index)
		}
	}
	// Sequence must be continuous starting at 1.
	if result[0].Metadata.Sequence != 1 || result[1].Metadata.Sequence != 2 {
		t.Errorf("unexpected sequence numbers: %d, %d",
			result[0].Metadata.Sequence, result[1].Metadata.Sequence)
	}
}

func TestBuilder_HeadingHierarchy(t *testing.T) {
	b := hierarchy.NewBuilder()
	blocks := []note.ContentBlock{
		buildHeading("# Chapter 1"),
		buildParagraph("text under chapter 1"),
		buildHeading("## Section 1.1"),
		buildParagraph("section text"),
		buildHeading("# Chapter 2"),
		buildParagraph("chapter 2 text"),
	}
	result, err := b.Build(blocks, "doc.md")
	if err != nil {
		t.Fatal(err)
	}

	// Verify paths and depths
	tests := []struct {
		idx   int
		path  string
		depth int
		seq   int
	}{
		{0, "Chapter 1", 0, 1},               // heading itself
		{1, "Chapter 1", 1, 2},               // non-heading under Chapter 1
		{2, "Chapter 1 > Section 1.1", 1, 1}, // heading Section 1.1
		{3, "Chapter 1 > Section 1.1", 2, 2}, // text under Section 1.1
		{4, "Chapter 2", 0, 1},               // Chapter 2 heading
		{5, "Chapter 2", 1, 2},               // text under Chapter 2
	}
	for _, tc := range tests {
		blk := result[tc.idx]
		if blk.Metadata.Path != tc.path {
			t.Errorf("block %d: path: expected %q, got %q", tc.idx, tc.path, blk.Metadata.Path)
		}
		if blk.Metadata.Depth != tc.depth {
			t.Errorf("block %d: depth: expected %d, got %d", tc.idx, tc.depth, blk.Metadata.Depth)
		}
		if blk.Metadata.Sequence != tc.seq {
			t.Errorf("block %d: sequence: expected %d, got %d", tc.idx, tc.seq, blk.Metadata.Sequence)
		}
		if blk.Metadata.Filepath != "doc.md" {
			t.Errorf("block %d: filepath not set", tc.idx)
		}
	}
}

func TestBuilder_ResetSequencePerPath(t *testing.T) {
	b := hierarchy.NewBuilder()
	blocks := []note.ContentBlock{
		buildHeading("# A"),
		buildParagraph("A1"),
		buildHeading("## B"),
		buildParagraph("B1"),
		buildHeading("# C"),
		buildParagraph("C1"),
	}
	result, _ := b.Build(blocks, "t.md")
	// sequence for "A" -> 1 (heading), 2 (paragraph)
	// sequence for "A > B" -> 1 (heading), 2 (paragraph)
	// sequence for "C" -> 1 (heading), 2 (paragraph)
	if result[0].Metadata.Sequence != 1 {
		t.Errorf("expected seq 1 for #A, got %d", result[0].Metadata.Sequence)
	}
	if result[1].Metadata.Sequence != 2 {
		t.Errorf("expected seq 2 for A1, got %d", result[1].Metadata.Sequence)
	}
	if result[2].Metadata.Sequence != 1 {
		t.Errorf("expected seq 1 for ##B, got %d", result[2].Metadata.Sequence)
	}
	if result[3].Metadata.Sequence != 2 {
		t.Errorf("expected seq 2 for B1, got %d", result[3].Metadata.Sequence)
	}
	if result[4].Metadata.Sequence != 1 {
		t.Errorf("expected seq 1 for #C, got %d", result[4].Metadata.Sequence)
	}
}

func TestBuilder_InvalidHeading(t *testing.T) {
	b := hierarchy.NewBuilder()
	blocks := []note.ContentBlock{
		{RawText: "not a heading", Metadata: note.BlockMetadata{Type: note.TypeHeading}},
	}
	_, err := b.Build(blocks, "file.md")
	if err == nil {
		t.Fatal("expected error for invalid heading")
	}
}

func TestBuilder_CustomSeparator(t *testing.T) {
	b := hierarchy.NewBuilder().WithSeparator(" / ")
	blocks := []note.ContentBlock{
		buildHeading("# Top"),
		buildHeading("## Child"),
	}
	result, _ := b.Build(blocks, "f.md")
	if result[1].Metadata.Path != "Top / Child" {
		t.Errorf("expected path with custom separator, got %q", result[1].Metadata.Path)
	}
}
