package note

import (
	"fmt"
	"strings"
)

// String возвращает подробное человеко-читаемое представление чанка.
func (c Chunk) String() string {
	var b strings.Builder
	b.WriteString("═══ Chunk ═══\n")
	fmt.Fprintf(&b, "ID:        %s\n", c.ID)

	if len(c.Embedding) == 0 {
		b.WriteString("Embedding: <empty>\n")
	} else {
		first := c.Embedding[0]
		last := c.Embedding[len(c.Embedding)-1]
		fmt.Fprintf(&b, "Embedding: [%d floats] first=%v last=%v\n", len(c.Embedding), first, last)
	}

	b.WriteString("Text:\n")
	b.WriteString("----\n")
	b.WriteString(c.Text)
	b.WriteString("\n----\n")

	b.WriteString("Metadata:\n")
	writeChunkMetadata(&b, c.Metadata)
	return b.String()
}

func writeChunkMetadata(b *strings.Builder, m ChunkMetadata) {
	fmt.Fprintf(b, "  SegmentID:    %s\n", m.SegmentID)
	fmt.Fprintf(b, "  ChunkIndex:   %d\n", m.ChunkIndex)
	fmt.Fprintf(b, "  TotalChunks:  %d\n", m.TotalChunks)
	fmt.Fprintf(b, "  FilePath:     %s\n", m.FilePath)
	fmt.Fprintf(b, "  Ancestors:    %v\n", m.Ancestors)
	fmt.Fprintf(b, "  HeadingText:  %q\n", m.HeadingText)
	fmt.Fprintf(b, "  HeadingLevel: %d\n", m.HeadingLevel)
	fmt.Fprintf(b, "  DomType:      %s\n", m.DomType)
	fmt.Fprintf(b, "  TokenCount:   %d\n", m.TokenCount)
	fmt.Fprintf(b, "  CharCount:    %d\n", m.CharCount)
	fmt.Fprintf(b, "  Tags:         %v\n", m.Tags)
	// fmt.Fprintf(b, "  Created:      %q\n", m.Created)
	// fmt.Fprintf(b, "  Updated:      %q\n", m.Updated)

	if len(m.Links) == 0 {
		b.WriteString("  Links:        []\n")
	} else {
		b.WriteString("  Links:\n")
		for i, l := range m.Links {
			fmt.Fprintf(b, "    [%d] %s\n", i, l.String())
		}
	}
}
