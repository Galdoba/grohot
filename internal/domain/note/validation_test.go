// validation_test.go
package note

import (
	"strings"
	"testing"
)

func TestIsValidType(t *testing.T) {
	valid := []BlockType{TypeHeading, TypeParagraph, TypeList, TypeCode, TypeTable, TypeQuote, TypeCallout, TypeHr}
	for _, typ := range valid {
		if !isValidType(typ) {
			t.Errorf("isValidType(%q) should be true", typ)
		}
	}
	invalid := []BlockType{"", "unknown", "heading "}
	for _, typ := range invalid {
		if isValidType(typ) {
			t.Errorf("isValidType(%q) should be false", typ)
		}
	}
}

func TestValidate(t *testing.T) {
	validNote := &Note{
		Name:     "note",
		Filepath: "/vault/note.md",
		Blocks: []ContentBlock{
			{Metadata: BlockMetadata{Filepath: "/vault/note.md", Type: TypeParagraph, Path: "P", Sequence: 1}},
			{Metadata: BlockMetadata{Filepath: "/vault/note.md", Type: TypeHeading, Path: "H", Sequence: 1}},
		},
	}
	if err := validNote.Validate(); err != nil {
		t.Errorf("valid note should pass validation, got error: %v", err)
	}

	tests := []struct {
		name    string
		note    *Note
		wantErr string
	}{
		{
			name:    "empty name",
			note:    &Note{Name: "", Filepath: "/vault/note.md"},
			wantErr: "note name cannot be empty",
		},
		{
			name:    "empty filepath",
			note:    &Note{Name: "n", Filepath: ""},
			wantErr: "filepath cannot be empty",
		},
		{
			name: "mismatched filepath",
			note: &Note{
				Name:     "n",
				Filepath: "/vault/note.md",
				Blocks: []ContentBlock{
					{Metadata: BlockMetadata{Filepath: "/other/note.md", Type: TypeParagraph}},
				},
			},
			wantErr: "mismatched filepath",
		},
		{
			name: "invalid type",
			note: &Note{
				Name:     "n",
				Filepath: "/vault/note.md",
				Blocks: []ContentBlock{
					{Metadata: BlockMetadata{Filepath: "/vault/note.md", Type: BlockType("bad")}},
				},
			},
			wantErr: "invalid type",
		},
		{
			name: "duplicate ID",
			note: &Note{
				Name:     "n",
				Filepath: "/vault/note.md",
				Blocks: []ContentBlock{
					{Metadata: BlockMetadata{Filepath: "/vault/note.md", Type: TypeParagraph, Path: "P", Sequence: 1}},
					{Metadata: BlockMetadata{Filepath: "/vault/note.md", Type: TypeParagraph, Path: "P", Sequence: 1}},
				},
			},
			wantErr: "duplicate block ID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.note.Validate()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want contains %q", err.Error(), tt.wantErr)
			}
		})
	}
}
