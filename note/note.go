package note

type Note struct {
	Name        string //name of note, used in [[internal links]]
	Filepath    string //path to note file
	Frontmatter map[string]any
	Nodes       []ContentNode
}

type ContentNode struct {
	NodeID string
	Type   string
	Text   string
}

