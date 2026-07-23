// File: hierarchy/heading.go
package hierarchy

import (
	"fmt"
	"regexp"
	"strings"
)

// headingRegex matches an ATX heading line and captures the level (1-6 #) and the text.
// It allows leading whitespace and at least one space after the #s.
var headingRegex = regexp.MustCompile(`^[ \t]*(#{1,6})[ \t]+(.*)`)

// extractHeading parses a raw heading string (e.g. "### My Title") and returns
// the level (number of #) and the heading text. An error is returned if the string
// does not match the expected heading format.
func extractHeading(raw string) (level int, text string, err error) {
	matches := headingRegex.FindStringSubmatch(raw)
	if matches == nil {
		return 0, "", fmt.Errorf("not a valid ATX heading")
	}
	level = len(matches[1])                // number of # characters
	text = strings.TrimSpace(matches[2])  // heading text, may be empty
	return level, text, nil
}