// note_test.go
package note

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewNote(t *testing.T) {
	tests := []struct {
		name    string
		cfg     NoteConfig
		wantErr bool
	}{
		{
			name:    "valid",
			cfg:     NoteConfig{Name: "note", Filepath: "/vault/note.md"},
			wantErr: false,
		},
		{
			name:    "empty name",
			cfg:     NoteConfig{Name: " ", Filepath: "/vault/note.md"},
			wantErr: true,
		},
		{
			name:    "empty filepath",
			cfg:     NoteConfig{Name: "note", Filepath: ""},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// mock parser and builder for Load tests
type mockParser struct {
	blocks []ContentBlock
	err    error
}

func (m *mockParser) Parse(lines []string) ([]ContentBlock, error) {
	return m.blocks, m.err
}

type mockBuilder struct {
	blocks []ContentBlock
	err    error
}

func (m *mockBuilder) Build(blocks []ContentBlock, filepath string) ([]ContentBlock, error) {
	if m.blocks != nil {
		return m.blocks, m.err
	}
	// default: return same blocks
	return blocks, m.err
}

func TestLoad(t *testing.T) {
	// Create a temp file
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	content := "---\ntitle: Test Note\n---\n# Heading\nSome text\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := &mockParser{blocks: []ContentBlock{
		{Metadata: BlockMetadata{Type: TypeHeading, Depth: 0, Path: "Heading", Sequence: 1}},
		{Metadata: BlockMetadata{Type: TypeParagraph, Depth: 0, Path: "Heading", Sequence: 2}},
	}}
	builder := &mockBuilder{} // returns same blocks

	note, err := Load(path, parser, builder)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if note.Name != "test" {
		t.Errorf("Name = %q, want test", note.Name)
	}
	if note.Filepath != path {
		t.Errorf("Filepath = %q, want %q", note.Filepath, path)
	}
	if note.Frontmatter == nil {
		t.Error("Frontmatter should not be nil")
	} else {
		if title := note.Frontmatter.Get("title"); title == nil || title.Value.Scalar != "Test Note" {
			t.Errorf("frontmatter title not parsed correctly")
		}
	}
	if len(note.Blocks) != 2 {
		t.Errorf("len(Blocks) = %d, want 2", len(note.Blocks))
	}
	if note.Tree.Root == nil {
		t.Error("Tree.Root should not be nil")
	}
	if len(note.Segments()) == 0 {
		t.Error("Segments() should return at least one segment")
	}
}

func TestBuildHierarchy(t *testing.T) {
	n := &Note{
		Name:     "test",
		Filepath: "/vault/test.md",
		Blocks:   []ContentBlock{{Metadata: BlockMetadata{Type: TypeParagraph}}},
		Tree:     SegmentTree{Root: &Segment{}},
	}
	builder := &mockBuilder{blocks: []ContentBlock{
		{Metadata: BlockMetadata{Type: TypeParagraph, Path: "P", Sequence: 1}},
	}}
	updated, err := n.BuildHierarchy(builder)
	if err != nil {
		t.Fatal(err)
	}
	if updated == n {
		t.Error("BuildHierarchy should return a new note")
	}
	if len(updated.Blocks) != 1 || updated.Blocks[0].Metadata.Path != "P" {
		t.Errorf("BuildHierarchy did not use builder result")
	}
}

func TestSegments(t *testing.T) {
	t.Run("root nil", func(t *testing.T) {
		n := &Note{Tree: SegmentTree{Root: nil}}
		segs := n.Segments()
		if segs != nil {
			t.Errorf("Segments() with nil root should return nil, got %v", segs)
		}
	})
	t.Run("root set", func(t *testing.T) {
		n := &Note{Tree: SegmentTree{Root: &Segment{}}}
		segs := n.Segments()
		if len(segs) != 1 {
			t.Errorf("Segments() returned %d segments, want 1", len(segs))
		}
	})
}
