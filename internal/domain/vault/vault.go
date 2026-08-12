package vault

import "github.com/Galdoba/grohot/internal/domain/note"

// Vault как структура данных это папка имеющая в себе 2 маркерных файла
// ./obsidian/types.json
// ./obsidian/graph.json
//
// Её имя это название папки.
type Vault struct {
	Root  string
	Name  string
	Notes map[string]*note.Note
}

func New(path string) *Vault {
	return &Vault{}
}

// Index индексирует заметку превращая её в сегменты/чанки.
func (v *Vault) Index(path string) error {
	return nil
}
