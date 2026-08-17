package vault

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Galdoba/grohot/internal/domain/link"
	"github.com/Galdoba/grohot/internal/domain/note"
)

// BuildLinks iterates over all notes, clears old links, and builds new ones.
// It populates each segment's OutgoingLinks and IncomingLinks.
func (v *Vault) BuildLinks() error {
	if v.debug {
		fmt.Println("[vault] building links...")
	}

	v.clearAllLinks()

	for sourceRelPath, sourceNote := range v.notes {
		v.processNoteLinks(sourceRelPath, sourceNote)
	}

	if v.debug {
		fmt.Println("[vault] link building complete")
	}
	return nil
}

// clearAllLinks clears all outgoing and incoming links on every segment of every note.
func (v *Vault) clearAllLinks() {
	for _, n := range v.notes {
		v.clearNoteLinks(n)
	}
}

// clearNoteLinks clears links on all segments of a single note.
func (v *Vault) clearNoteLinks(n *note.Note) {
	for _, seg := range n.Segments() {
		seg.OutgoingLinks = nil
		seg.IncomingLinks = nil
	}
}

// processNoteLinks processes links for all segments of a note.
func (v *Vault) processNoteLinks(sourceRelPath string, sourceNote *note.Note) {
	for _, seg := range sourceNote.Segments() {
		v.processSegmentLinks(sourceRelPath, seg)
	}
}

// processSegmentLinks extracts links from a segment, resolves them, and updates link slices.
func (v *Vault) processSegmentLinks(sourceRelPath string, seg *note.Segment) {
	rawText := v.collectSegmentText(seg)
	links, err := link.Extract(rawText)
	if err != nil {
		fmt.Printf("[vault] found link error in %s (%s): %v\n", sourceRelPath, seg.ID, err)
	}
	for _, l := range links {
		resolved := v.resolveLink(sourceRelPath, seg, l)
		seg.OutgoingLinks = append(seg.OutgoingLinks, resolved)
		if resolved.LinkType != link.DeadLink {
			v.addBackLink(resolved)
		}
	}
	if v.debug && len(links) > 0 {
		fmt.Printf("[vault] found %d links in %s (%s)\n", len(links), sourceRelPath, seg.ID)
	}
}

// collectSegmentText concatenates header and code-own blocks (excluding code blocks) into a single string.
func (v *Vault) collectSegmentText(seg *note.Segment) string {
	var rawText strings.Builder
	if seg.Header != nil {
		rawText.WriteString(seg.Header.RawText)
		rawText.WriteString("\n")
	}
	for _, block := range seg.OwnBlocks {
		if block.Metadata.Type == note.TypeCode {
			continue
		}
		rawText.WriteString(block.RawText)
		rawText.WriteString("\n")
	}
	return rawText.String()
}

// addBackLink creates a backlink and appends it to the target segment's IncomingLinks.
func (v *Vault) addBackLink(l link.Link) {
	back := buildBackLink(l)
	targetSeg := v.findSegmentByID(back.TargetSegmentID)
	if targetSeg != nil {
		targetSeg.IncomingLinks = append(targetSeg.IncomingLinks, back)
	}
}

// resolveLink fills in Source* fields and resolves Target* fields for the given link.
func (v *Vault) resolveLink(sourceRelPath string, sourceSeg *note.Segment, l link.Link) link.Link {
	l.SourceFile = sourceRelPath
	l.SourceSegmentID = sourceSeg.ID
	l.TargetFile = v.normalizeTargetFile(l.TargetFile, sourceRelPath)
	l.TargetSegmentID = ""

	targetNote, found := v.notes[l.TargetFile]
	if !found {
		l.LinkType = link.DeadLink
		return l
	}

	if l.TargetHeading != "" {
		targetSeg := v.findSegmentByHeading(targetNote, l.TargetHeading)
		if targetSeg == nil {
			l.LinkType = link.DeadLink
		} else {
			l.TargetSegmentID = targetSeg.ID
			l.LinkType = link.Direct
		}
	} else {
		rootSeg := targetNote.Tree.Root
		l.TargetSegmentID = rootSeg.ID
		l.LinkType = link.Direct
	}
	return l
}

// normalizeTargetFile converts the target to a full relative path (with .md) from the vault root.
func (v *Vault) normalizeTargetFile(target, sourceRelPath string) string {
	target = filepath.ToSlash(strings.TrimSpace(target))
	if target == "" {
		return ""
	}
	if isAbsoluteOrRelativePath(target) {
		return v.normalizeFullPathTarget(target)
	}
	return v.resolveShortNameTarget(target, sourceRelPath)
}

// isAbsoluteOrRelativePath reports whether the target looks like a path (absolute, relative, or containing slash).
func isAbsoluteOrRelativePath(target string) bool {
	return strings.HasPrefix(target, pathSeparator) ||
		strings.HasPrefix(target, dotSlashPrefix) ||
		strings.Contains(target, pathSeparator)
}

// normalizeFullPathTarget ensures the target has .md extension and strips leading ./ and / if present.
func (v *Vault) normalizeFullPathTarget(target string) string {
	if !strings.HasSuffix(target, markdownExtension) {
		target += markdownExtension
	}
	target = strings.TrimPrefix(target, dotSlashPrefix)
	target = strings.TrimPrefix(target, pathSeparator) // remove leading slash
	return target
}

// resolveShortNameTarget resolves a short note name using byName index.
func (v *Vault) resolveShortNameTarget(target, sourceRelPath string) string {
	key := strings.ToLower(strings.TrimSuffix(target, markdownExtension))
	candidates := v.byName[key]
	if len(candidates) == 1 {
		return candidates[0]
	}
	if len(candidates) > 1 {
		return v.pickBestCandidate(candidates, sourceRelPath)
	}
	if !strings.HasSuffix(target, markdownExtension) {
		target += markdownExtension
	}
	return target
}

// pickBestCandidate chooses the best candidate from multiple notes with the same base name.
// Preference is given to notes in the same directory as the source.
func (v *Vault) pickBestCandidate(candidates []string, sourceRelPath string) string {
	sourceDir := filepath.ToSlash(filepath.Dir(sourceRelPath))
	for _, c := range candidates {
		if filepath.ToSlash(filepath.Dir(c)) == sourceDir {
			return c
		}
	}
	// Fallback to the first candidate (could be improved with logging)
	return candidates[0]
}

// findSegmentByHeading returns the segment with the given heading (case-insensitive), or nil if not found.
func (v *Vault) findSegmentByHeading(n *note.Note, heading string) *note.Segment {
	heading = strings.ToLower(strings.TrimSpace(heading))
	for _, seg := range n.Segments() {
		if seg.Header == nil {
			continue
		}
		text := strings.ToLower(strings.TrimSpace(seg.HeadingText()))
		if text == heading {
			return seg
		}
	}
	return nil
}

// findSegmentByID searches for a segment by its ID across all notes.
func (v *Vault) findSegmentByID(id string) *note.Segment {
	for _, n := range v.notes {
		if seg := findSegmentInNote(n, id); seg != nil {
			return seg
		}
	}
	return nil
}

// findSegmentInNote searches for a segment by ID within a single note.
func findSegmentInNote(n *note.Note, id string) *note.Segment {
	for _, seg := range n.Segments() {
		if seg.ID == id {
			return seg
		}
	}
	return nil
}

// buildBackLink creates a backlink from the given direct link.
func buildBackLink(l link.Link) link.Link {
	l.LinkType = link.BackLink
	return l
}

// removeString removes the first occurrence of target from slice and returns a new slice.
// It does not mutate the original slice.
func removeString(sl []string, target string) []string {
	for i, s := range sl {
		if s == target {
			result := make([]string, 0, len(sl)-1)
			result = append(result, sl[:i]...)
			result = append(result, sl[i+1:]...)
			return result
		}
	}
	return sl
}
