package vault

import (
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// Tests for pure functions (no dependency on note package)
// ---------------------------------------------------------------------------

func TestIsVaultRoot(t *testing.T) {
	root := t.TempDir()
	if isVaultRoot(root) {
		t.Fatal("should not be vault root without .obsidian")
	}

	obsidianDir := filepath.Join(root, ".obsidian")
	if err := os.Mkdir(obsidianDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if isVaultRoot(root) {
		t.Fatal("should not be vault root without required files")
	}

	if err := os.WriteFile(filepath.Join(obsidianDir, "types.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(obsidianDir, "graph.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !isVaultRoot(root) {
		t.Fatal("should be vault root with .obsidian and required files")
	}
}

func TestFindVaultRoot(t *testing.T) {
	root := t.TempDir()
	obsidianDir := filepath.Join(root, ".obsidian")
	if err := os.Mkdir(obsidianDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(obsidianDir, "types.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(obsidianDir, "graph.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	found, err := FindVaultRoot(nested)
	if err != nil {
		t.Fatalf("FindVaultRoot returned error: %v", err)
	}
	if found != root {
		t.Errorf("FindVaultRoot = %q, want %q", found, root)
	}

	// Error case
	outside := t.TempDir()
	if _, err := FindVaultRoot(outside); err == nil {
		t.Error("expected error when vault root not found")
	}
}

func TestNormalizeTargetFile(t *testing.T) {
	v := &Vault{
		byName: map[string][]string{
			"note":    {"note.md"},
			"dup":     {"folder1/dup.md", "folder2/dup.md"},
			"withext": {"withext.md"},
		},
	}

	tests := []struct {
		name          string
		target        string
		sourceRelPath string
		want          string
	}{
		{"empty target", "", "any.md", ""},
		{"absolute path", "/some/path.md", "any.md", "some/path.md"},
		{"relative with ./", "./some/path", "any.md", "some/path.md"},
		{"relative with slash", "some/path", "any.md", "some/path.md"},
		{"short name unique", "note", "any.md", "note.md"},
		{"short name with .md", "note.md", "any.md", "note.md"},
		{"short name multiple, source same dir", "dup", "folder1/source.md", "folder1/dup.md"},
		{"short name multiple, source different dir", "dup", "other/source.md", "folder1/dup.md"},
		{"short name not found", "missing", "any.md", "missing.md"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := v.normalizeTargetFile(tc.target, tc.sourceRelPath)
			if got != tc.want {
				t.Errorf("normalizeTargetFile(%q, %q) = %q, want %q", tc.target, tc.sourceRelPath, got, tc.want)
			}
		})
	}
}

func TestRemoveString(t *testing.T) {
	cases := []struct {
		name   string
		input  []string
		target string
		want   []string
	}{
		{"remove middle", []string{"a", "b", "c"}, "b", []string{"a", "c"}},
		{"remove first", []string{"a", "b", "c"}, "a", []string{"b", "c"}},
		{"remove last", []string{"a", "b", "c"}, "c", []string{"a", "b"}},
		{"not found", []string{"a", "b", "c"}, "d", []string{"a", "b", "c"}},
		{"empty slice", []string{}, "a", []string{}},
		{"single element", []string{"a"}, "a", []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			original := append([]string(nil), tc.input...)
			got := removeString(tc.input, tc.target)
			if !equalStringSlices(got, tc.want) {
				t.Errorf("removeString(%v, %q) = %v, want %v", original, tc.target, got, tc.want)
			}
			if !equalStringSlices(tc.input, original) {
				t.Errorf("original slice mutated: %v -> %v", original, tc.input)
			}
		})
	}
}

func TestIsHiddenDir(t *testing.T) {
	if !isHiddenDir(".obsidian") {
		t.Error(".obsidian should be hidden")
	}
	if isHiddenDir("visible") {
		t.Error("visible should not be hidden")
	}
	if isHiddenDir(".") {
		t.Error("dot should not be considered hidden")
	}
}

func TestIsMarkdownFile(t *testing.T) {
	if !isMarkdownFile("note.md") {
		t.Error("note.md should be markdown")
	}
	if !isMarkdownFile("NOTE.MD") {
		t.Error("case insensitive")
	}
	if isMarkdownFile("note.txt") {
		t.Error("txt should not be markdown")
	}
}

func TestNoteBaseName(t *testing.T) {
	cases := map[string]string{
		"folder/Note.md":     "note",
		"Note.MD":            "note",
		"path/to/my.note.md": "my.note",
		"noext":              "noext",
	}
	for input, want := range cases {
		got := noteBaseName(input)
		if got != want {
			t.Errorf("noteBaseName(%q) = %q, want %q", input, got, want)
		}
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
