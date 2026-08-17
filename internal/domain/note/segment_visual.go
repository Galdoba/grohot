package note

import (
	"fmt"
	"strings"
)

// Display width constants for visual output.
const (
	headingDisplayWidth  = 20 // number of runes for heading text in tree view
	ownBlockDisplayWidth = 60 // number of runes for own block text in tree view
)

// Visual returns a pseudo-graphical string representation of the segment tree.
// The root is shown as "root". Headings include their level (h1-h6),
// and non-heading blocks show their type and a preview of the content.
func (st SegmentTree) Visual() string {
	var b strings.Builder
	b.WriteString("root")

	if len(st.Root.OwnBlocks) == 0 && len(st.Root.Children) == 0 {
		return b.String()
	}
	b.WriteByte('\n')
	visualSegment(st.Root, "", &b)
	return b.String()
}

// visualSegment recursively writes the tree structure of a segment.
func visualSegment(seg *Segment, prefix string, b *strings.Builder) {
	total := len(seg.OwnBlocks) + len(seg.Children)
	idx := 0

	for _, block := range seg.OwnBlocks {
		writeVisualBlock(b, prefix, chooseConnector(idx, total), block)
		idx++
	}

	for _, child := range seg.Children {
		writeVisualChild(b, prefix, chooseConnector(idx, total), child)
		newPrefix := prefix
		if idx == total-1 {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		visualSegment(child, newPrefix, b)
		idx++
	}
}

// writeVisualBlock writes a single own-block line to the builder.
func writeVisualBlock(b *strings.Builder, prefix, connector string, block ContentBlock) {
	text := formatOwnBlockText(block.RawText)
	line := fmt.Sprintf("%s%s[%s] %s\n", prefix, connector, block.Metadata.Type, text)
	b.WriteString(line)
}

// writeVisualChild writes a single child segment header line to the builder.
func writeVisualChild(b *strings.Builder, prefix, connector string, child *Segment) {
	headingText := "<no heading>"
	levelMarker := "??"
	if child.Header != nil {
		headingText = formatHeadingText(child.Header.RawText)
		level := headingLevel(child.Header.RawText)
		levelMarker = fmt.Sprintf("h%d", level)
	}
	headerLine := fmt.Sprintf("%s%s[%s] %s\n", prefix, connector, levelMarker, headingText)
	b.WriteString(headerLine)
}

// chooseConnector returns the tree connector for a given item index.
func chooseConnector(index, total int) string {
	if index == total-1 {
		return "└── "
	}
	return "├── "
}

// stripHeadingMarkers removes leading # and spaces from a heading line.
func stripHeadingMarkers(raw string) string {
	s := strings.TrimLeft(raw, "#")
	return strings.TrimSpace(s)
}

// headingLevel returns the number of # at the start of the raw heading line.
func headingLevel(raw string) int {
	count := 0
	for _, ch := range raw {
		if ch == '#' {
			count++
		} else {
			break
		}
	}
	return count
}

// truncatePad truncates a string to n runes and pads with spaces to exactly n runes
// (for monospaced alignment).
func truncatePad(s string, n int) string {
	runes := []rune(s)
	if len(runes) > n {
		runes = runes[:n]
	}
	padded := make([]rune, n)
	copy(padded, runes)
	for i := len(runes); i < n; i++ {
		padded[i] = ' '
	}
	return string(padded)
}

// formatHeadingText keeps the heading markers and truncates/pads to headingDisplayWidth.
func formatHeadingText(raw string) string {
	clean := strings.ReplaceAll(raw, "\r", "")
	return truncatePad(clean, headingDisplayWidth)
}

// formatOwnBlockText shows the first line of RawText truncated/padded to ownBlockDisplayWidth.
func formatOwnBlockText(raw string) string {
	clean := strings.ReplaceAll(raw, "\r", "")
	firstLine := clean
	if idx := strings.IndexByte(clean, '\n'); idx != -1 {
		firstLine = clean[:idx]
	}
	return truncatePad(firstLine, ownBlockDisplayWidth)
}
