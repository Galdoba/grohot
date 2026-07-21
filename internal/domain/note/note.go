package note

import "github.com/Galdoba/grohot/internal/domain/note/frontmatter"

type Note struct {
	Name        string //name of note, used in [[internal links]]
	Filepath    string //path to note file
	Frontmatter frontmatter.Frontmatter
	Nodes       []ContentNode
}

type ContentNode struct {
	NodeID string
	Type   string
	Text   string
}
