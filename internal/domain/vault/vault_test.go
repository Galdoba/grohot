package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Galdoba/grohot/internal/domain/link"
	"github.com/Galdoba/grohot/internal/domain/note"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// createVaultRoot creates a valid Obsidian vault root directory structure.
func createVaultRoot(t *testing.T, root string) {
	t.Helper()
	obsidianDir := filepath.Join(root, ".obsidian")
	if err := os.Mkdir(obsidianDir, 0o755); err != nil {
		t.Fatalf("failed to create .obsidian: %v", err)
	}
	if err := os.WriteFile(filepath.Join(obsidianDir, "types.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(obsidianDir, "graph.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeFile writes content to path, creating parent directories as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// makeTestSegment creates a segment with an optional heading and own blocks.
func makeTestSegment(id string, heading string, blocks ...note.ContentBlock) *note.Segment {
	seg := &note.Segment{ID: id}
	if heading != "" {
		seg.Header = &note.ContentBlock{RawText: "# " + heading + "\n"}
	}
	seg.OwnBlocks = blocks
	return seg
}

// makeTestNoteWithSegments creates a note with the given segments.
// The first segment becomes the root, others are attached as children.
func makeTestNoteWithSegments(name string, root *note.Segment, children ...*note.Segment) *note.Note {
	root.Children = children
	for _, c := range children {
		c.Parent = root
	}
	return &note.Note{
		Name:     name,
		Filepath: name + ".md",
		Tree:     note.SegmentTree{Root: root},
	}
}

// mockParser implements note.BlockParser for tests.
type mockParser struct {
	blocks []note.ContentBlock
	err    error
}

func (m *mockParser) Parse(lines []string) ([]note.ContentBlock, error) {
	return m.blocks, m.err
}

// mockBuilder implements note.HierarchyBuilder for tests.
type mockBuilder struct {
	blocks []note.ContentBlock
	err    error
}

func (m *mockBuilder) Build(blocks []note.ContentBlock, filepath string) ([]note.ContentBlock, error) {
	return m.blocks, m.err
}

// mockChunker implements note.Chunker for tests.
type mockChunker struct {
	chunks map[string][]note.Chunk
	err    error
}

func (m *mockChunker) Chunk(seg *note.Segment) ([]note.Chunk, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.chunks[seg.ID], nil
}

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

// ---------------------------------------------------------------------------
// Tests for methods that depend on note package
// ---------------------------------------------------------------------------

func TestNoteAndAllNotes(t *testing.T) {
	v := &Vault{
		notes:  make(map[string]*note.Note),
		byName: make(map[string][]string),
	}
	n := makeTestNoteWithSegments("test", makeTestSegment("seg1", "Test"))
	relPath := "folder/test.md"
	v.notes[relPath] = n

	gotNote, ok := v.Note(relPath)
	if !ok || gotNote != n {
		t.Fatalf("Note(%q) = %v, %v; want %v, true", relPath, gotNote, ok, n)
	}

	gotNote, ok = v.Note("folder\\test.md")
	if !ok || gotNote != n {
		t.Fatalf("Note with backslashes failed: got %v, %v", gotNote, ok)
	}

	all := v.AllNotes()
	if len(all) != 1 || all[0] != n {
		t.Fatalf("AllNotes returned %v, want [%v]", all, n)
	}
}

func TestCollectSegmentText(t *testing.T) {
	v := &Vault{}
	seg := &note.Segment{
		Header:    &note.ContentBlock{RawText: "# Heading\n"},
		OwnBlocks: []note.ContentBlock{
			{RawText: "line1"},
			{RawText: "line2"},
		},
	}
	text := v.collectSegmentText(seg)
	want := "# Heading\n\nline1\nline2\n"
	if text != want {
		t.Errorf("collectSegmentText = %q, want %q", text, want)
	}

	segWithCode := &note.Segment{
		Header:    &note.ContentBlock{RawText: "# H\n"},
		OwnBlocks: []note.ContentBlock{
			{RawText: "text"},
			{RawText: "code", Metadata: note.BlockMetadata{Type: note.TypeCode}},
		},
	}
	text = v.collectSegmentText(segWithCode)
	want = "# H\n\ntext\n"
	if text != want {
		t.Errorf("collectSegmentText with code = %q, want %q", text, want)
	}
}

func TestFindSegmentByHeading(t *testing.T) {
	v := &Vault{}
	root := makeTestSegment("root", "")
	child1 := makeTestSegment("child1", "First")
	child2 := makeTestSegment("child2", "Second")
	n := makeTestNoteWithSegments("test", root, child1, child2)

	seg := v.findSegmentByHeading(n, "first")
	if seg == nil || seg.ID != "child1" {
		t.Errorf("expected child1, got %v", seg)
	}

	seg = v.findSegmentByHeading(n, "nonexistent")
	if seg != nil {
		t.Error("expected nil for unknown heading")
	}
}

func TestFindSegmentByID(t *testing.T) {
	v := &Vault{
		notes: make(map[string]*note.Note),
	}
	n1 := makeTestNoteWithSegments("note1", makeTestSegment("seg1", "One"))
	n2 := makeTestNoteWithSegments("note2", makeTestSegment("seg2", "Two"))
	v.notes["note1.md"] = n1
	v.notes["note2.md"] = n2

	seg := v.findSegmentByID("seg2")
	if seg == nil || seg.ID != "seg2" {
		t.Errorf("expected seg2, got %v", seg)
	}
	if v.findSegmentByID("nonexistent") != nil {
		t.Error("expected nil for unknown ID")
	}
}

func TestBuildBackLink(t *testing.T) {
	l := link.Link{LinkType: link.Direct, SourceFile: "s", TargetFile: "t"}
	back := buildBackLink(l)
	if back.LinkType != link.BackLink {
		t.Errorf("expected BackLink, got %v", back.LinkType)
	}
	if back.SourceFile != "s" || back.TargetFile != "t" {
		t.Errorf("backlink fields not preserved: %+v", back)
	}
}

func TestPickBestCandidate(t *testing.T) {
	v := &Vault{}
	candidates := []string{"folder1/dup.md", "folder2/dup.md"}

	got := v.pickBestCandidate(candidates, "folder1/source.md")
	if got != "folder1/dup.md" {
		t.Errorf("pickBestCandidate = %q, want %q", got, "folder1/dup.md")
	}

	got = v.pickBestCandidate(candidates, "other/source.md")
	if got != "folder1/dup.md" {
		t.Errorf("pickBestCandidate fallback = %q, want %q", got, "folder1/dup.md")
	}
}

func TestScanAndLoad(t *testing.T) {
	root := t.TempDir()
	createVaultRoot(t, root)

	// Create markdown files, including hidden dirs
	writeFile(t, filepath.Join(root, "note1.md"), "dummy")
	writeFile(t, filepath.Join(root, "sub", "note2.md"), "dummy")
	writeFile(t, filepath.Join(root, ".obsidian", "ignored.md"), "dummy")
	writeFile(t, filepath.Join(root, "sub", ".hidden", "note3.md"), "dummy")

	sampleBlock := note.ContentBlock{RawText: "text", Metadata: note.BlockMetadata{Type: note.TypeParagraph, Depth: 0}}
	parser := &mockParser{blocks: []note.ContentBlock{sampleBlock}}
	builder := &mockBuilder{blocks: []note.ContentBlock{sampleBlock}}

	v, err := NewVault(root, Dependencies{Parser: parser, Builder: builder})
	if err != nil {
		t.Fatalf("NewVault failed: %v", err)
	}
	v.debug = false

	err = v.ScanAndLoad()
	if err != nil {
		t.Fatalf("ScanAndLoad returned error: %v", err)
	}

	expectedPaths := []string{"note1.md", "sub/note2.md"}
	for _, p := range expectedPaths {
		if _, ok := v.notes[p]; !ok {
			t.Errorf("missing note %q in vault.notes", p)
		}
	}
	if _, ok := v.notes[".obsidian/ignored.md"]; ok {
		t.Error("should not load .obsidian/ignored.md")
	}
	if _, ok := v.notes["sub/.hidden/note3.md"]; ok {
		t.Error("should not load sub/.hidden/note3.md")
	}

	if _, ok := v.byName["note1"]; !ok {
		t.Error("byName missing note1")
	}
	if _, ok := v.byName["note2"]; !ok {
		t.Error("byName missing note2")
	}
}

func TestReloadNote(t *testing.T) {
	root := t.TempDir()
	createVaultRoot(t, root)
	writeFile(t, filepath.Join(root, "note.md"), "dummy")

	sampleBlock := note.ContentBlock{RawText: "old", Metadata: note.BlockMetadata{Type: note.TypeParagraph}}
	parser := &mockParser{blocks: []note.ContentBlock{sampleBlock}}
	builder := &mockBuilder{blocks: []note.ContentBlock{sampleBlock}}

	v, err := NewVault(root, Dependencies{Parser: parser, Builder: builder})
	if err != nil {
		t.Fatalf("NewVault failed: %v", err)
	}
	v.debug = false

	err = v.ScanAndLoad()
	if err != nil {
		t.Fatalf("ScanAndLoad failed: %v", err)
	}
	oldNote := v.notes["note.md"]
	if oldNote == nil {
		t.Fatal("expected note.md to be loaded")
	}

	// Change mock to return new block
	newBlock := note.ContentBlock{RawText: "new", Metadata: note.BlockMetadata{Type: note.TypeParagraph}}
	parser.blocks = []note.ContentBlock{newBlock}
	builder.blocks = []note.ContentBlock{newBlock}

	err = v.ReloadNote("note.md")
	if err != nil {
		t.Fatalf("ReloadNote returned error: %v", err)
	}

	newNote := v.notes["note.md"]
	if newNote == oldNote {
		t.Error("expected note to be replaced")
	}
	if got := v.byName["note"]; len(got) != 1 || got[0] != "note.md" {
		t.Errorf("byName not updated correctly: %v", got)
	}
}

func TestBuildLinks(t *testing.T) {
	// Target note with root and child segment
	targetChild := makeTestSegment("target-child", "Child")
	targetRoot := makeTestSegment("target-root", "")
	targetNote := makeTestNoteWithSegments("target", targetRoot, targetChild)

	// Source note with link to target and target#Child
	sourceSeg := makeTestSegment("source-root", "", note.ContentBlock{
		RawText:  "Link to [[target]] and [[target#Child]]",
		Metadata: note.BlockMetadata{Type: note.TypeParagraph},
	})
	sourceNote := makeTestNoteWithSegments("source", sourceSeg)

	v := &Vault{
		notes: map[string]*note.Note{
			"source.md": sourceNote,
			"target.md": targetNote,
		},
		byName: map[string][]string{
			"source": {"source.md"},
			"target": {"target.md"},
		},
		debug: false,
	}

	// Pre-populate to test clearing
	sourceSeg.OutgoingLinks = []link.Link{{LinkType: link.Direct}}
	targetRoot.IncomingLinks = []link.Link{{LinkType: link.BackLink}}

	err := v.BuildLinks()
	if err != nil {
		t.Fatalf("BuildLinks returned error: %v", err)
	}

	// Check source outgoing links
	if len(sourceSeg.OutgoingLinks) != 2 {
		t.Fatalf("expected 2 outgoing links, got %d", len(sourceSeg.OutgoingLinks))
	}
	first := sourceSeg.OutgoingLinks[0]
	if first.SourceFile != "source.md" || first.SourceSegmentID != "source-root" ||
		first.TargetFile != "target.md" || first.TargetSegmentID != "target-root" ||
		first.LinkType != link.Direct {
		t.Errorf("first link incorrect: %+v", first)
	}
	second := sourceSeg.OutgoingLinks[1]
	if second.TargetSegmentID != "target-child" || second.LinkType != link.Direct {
		t.Errorf("second link incorrect: %+v", second)
	}

	// Check target incoming links
	if len(targetRoot.IncomingLinks) != 1 {
		t.Errorf("expected 1 incoming link to root, got %d", len(targetRoot.IncomingLinks))
	} else if targetRoot.IncomingLinks[0].LinkType != link.BackLink {
		t.Errorf("expected BackLink, got %v", targetRoot.IncomingLinks[0].LinkType)
	}
	if len(targetChild.IncomingLinks) != 1 {
		t.Errorf("expected 1 incoming link to child, got %d", len(targetChild.IncomingLinks))
	} else if targetChild.IncomingLinks[0].LinkType != link.BackLink {
		t.Errorf("expected BackLink, got %v", targetChild.IncomingLinks[0].LinkType)
	}
}

func TestResolveLink(t *testing.T) {
	targetChild := makeTestSegment("target-child", "Child")
	targetRoot := makeTestSegment("target-root", "")
	targetNote := makeTestNoteWithSegments("target", targetRoot, targetChild)

	sourceSeg := makeTestSegment("source-root", "")
	v := &Vault{
		notes: map[string]*note.Note{"target.md": targetNote},
		byName: map[string][]string{"target": {"target.md"}},
	}

	// Direct link to note (no heading)
	l := link.Link{TargetFile: "target"}
	resolved := v.resolveLink("source.md", sourceSeg, l)
	if resolved.LinkType != link.Direct || resolved.TargetFile != "target.md" || resolved.TargetSegmentID != "target-root" {
		t.Errorf("direct link incorrect: %+v", resolved)
	}

	// Link to heading
	l = link.Link{TargetFile: "target", TargetHeading: "Child"}
	resolved = v.resolveLink("source.md", sourceSeg, l)
	if resolved.LinkType != link.Direct || resolved.TargetSegmentID != "target-child" {
		t.Errorf("heading link incorrect: %+v", resolved)
	}

	// Dead link: file not found
	l = link.Link{TargetFile: "nonexistent"}
	resolved = v.resolveLink("source.md", sourceSeg, l)
	if resolved.LinkType != link.DeadLink {
		t.Errorf("expected DeadLink, got %v", resolved.LinkType)
	}

	// Dead link: heading not found
	l = link.Link{TargetFile: "target", TargetHeading: "Missing"}
	resolved = v.resolveLink("source.md", sourceSeg, l)
	if resolved.LinkType != link.DeadLink {
		t.Errorf("expected DeadLink for missing heading, got %v", resolved.LinkType)
	}
}

func TestChunkAll(t *testing.T) {
	seg1 := makeTestSegment("seg1", "One")
	seg2 := makeTestSegment("seg2", "Two")
	n := makeTestNoteWithSegments("test", seg1, seg2)

	v := &Vault{
		notes:   map[string]*note.Note{"test.md": n},
		chunker: &mockChunker{chunks: map[string][]note.Chunk{
			"seg1": {{ID: "c1"}, {ID: "c2"}},
			"seg2": {{ID: "c3"}},
		}},
	}

	chunks, err := v.ChunkAll()
	if err != nil {
		t.Fatalf("ChunkAll returned error: %v", err)
	}
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}

	// Error case
	v.chunker = &mockChunker{err: fmt.Errorf("chunk error")}
	_, err = v.ChunkAll()
	if err == nil {
		t.Error("expected error from chunker")
	}
}