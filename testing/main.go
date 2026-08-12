package main

import (
	"fmt"

	"github.com/Galdoba/grohot/internal/domain/note"
	"github.com/Galdoba/grohot/internal/infrastructure/obsidian/hierarchy"
	"github.com/Galdoba/grohot/internal/infrastructure/obsidian/parser"
)

func main() {
	path := `c:\Users\pemaltynov\Documents\Obsidian\nounamental_rhizome\Projects\books\_tests\test TExt.md`
	// data, _ := os.ReadFile(path)
	// lines := strings.Split(string(data), "\n")
	parser := parser.NewParser()
	// bl, err := parser.NewParser().Parse(lines)
	// fmt.Println(err)
	// for i, b := range bl {
	// 	fmt.Println(i, b)
	// }
	bldr := hierarchy.NewBuilder()
	// hBl, err2 := hierarchy.NewBuilder().Build(bl, path)
	// fmt.Println(err2)
	// for i, b := range hBl {
	// 	fmt.Println(i, b.Metadata)
	// }
	n, err := note.Load(path, parser, bldr)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(n.Filepath)
	// fmt.Println(n.Frontmatter.String())
	// for _, bl := range n.Blocks {
	// 	fmt.Println(bl)
	// }

	fmt.Println(n.Tree.Visual())
	fmt.Println("")

	segments := n.Segments()
	for _, seg := range segments {
		fmt.Println(seg.String())
	}
	fmt.Println("")
	fmt.Printf("total %d segments", len(segments))
}

// func buildHeading(raw string) note.ContentBlock {
// 	return note.ContentBlock{
// 		RawText: raw,
// 		Metadata: note.BlockMetadata{
// 			Type: note.TypeHeading,
// 		},
// 	}
// }

// func buildParagraph(raw string) note.ContentBlock {
// 	return note.ContentBlock{
// 		RawText: raw,
// 		Metadata: note.BlockMetadata{
// 			Type: note.TypeParagraph,
// 		},
// 	}
// }
