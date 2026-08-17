// Package link provides parsing and representation of Obsidian wiki links.
package link

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Type represents the kind of a link.
type Type string

const (
	// Direct is an outgoing link whose target has been found.
	Direct Type = "direct"
	// BackLink is an incoming link (the reverse side of Direct).
	BackLink Type = "backlink"
	// DeadLink is an outgoing link whose target has not been found.
	DeadLink Type = "deadlink"
)

// Link represents a single Obsidian wiki link.
// Source fields are populated by the Vault after resolution.
type Link struct {
	SourceFile      string `json:"source_file"`                 // relative path including .md extension
	SourceSegmentID string `json:"source_segment_id"`           // ID of the source segment
	TargetFile      string `json:"target_file"`                 // relative path including .md extension
	TargetHeading   string `json:"target_heading,omitempty"`    // heading of the target segment
	TargetSegmentID string `json:"target_segment_id,omitempty"` // ID of the target segment
	DisplayText     string `json:"display_text,omitempty"`
	IsTransclusion  bool   `json:"is_transclusion,omitempty"`
	LinkType        Type   `json:"link_type"`
}

const (
	transclusionPrefix  = "!"
	linkOpenDelimiter   = "[["
	linkCloseDelimiter  = "]]"
	pipeDelimiter       = "|"
	headingDelimiter    = "#"
	headingPrefix       = " #"
	truncatedIDLength   = 8
	truncationSuffix    = "..."
	parseErrorFormat    = "%q: %v"
	parseErrorSeparator = "; "
	parseErrorMessage   = "some links failed to parse: %s"
	linkStringFormat    = "Link{Type:%s, Source:%s|%s, Target:%s|%s%s, Display:%q, Transclusion:%t}"

	msgInvalidWikilinkFormat = "invalid wikilink format"
	msgEmptyLinkTarget       = "link target cannot be empty"
)

var (
	// ErrInvalidWikilinkFormat indicates that a string does not match the expected [[...]] shape.
	ErrInvalidWikilinkFormat = errors.New(msgInvalidWikilinkFormat)
	// ErrEmptyLinkTarget indicates that a parsed link has neither file nor heading.
	ErrEmptyLinkTarget = errors.New(msgEmptyLinkTarget)
)

// wikiLinkRegex matches any potential wiki-link candidate, including
// malformed ones (unclosed, empty). The actual validation happens in parseLink.
var wikiLinkRegex = regexp.MustCompile(`!?\[\[[^\]]*\]?\]?`)

// Extract returns all wiki links found in rawText.
// If some links cannot be parsed, it returns the successfully parsed links
// together with an error describing all failures.
func Extract(rawText string) ([]Link, error) {
	links := make([]Link, 0)
	var parseErrors []string

	for _, rawLink := range extractFromText(rawText) {
		link, err := parseLink(rawLink)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf(parseErrorFormat, rawLink, err))
			continue
		}
		links = append(links, link)
	}

	if len(parseErrors) > 0 {
		return links, fmt.Errorf(parseErrorMessage, strings.Join(parseErrors, parseErrorSeparator))
	}
	return links, nil
}

// extractFromText returns every wiki-link substring found in rawText.
func extractFromText(rawText string) []string {
	return wikiLinkRegex.FindAllString(rawText, -1)
}

// parseLink parses a single wiki-link substring such as "[[Target|Display]]"
// or "![[Target]]". It returns a Link with LinkType Direct; the Vault may
// later change it to DeadLink if the target is not found. Source fields are
// left empty for the caller to populate.
func parseLink(rawLink string) (Link, error) {
	isTransclusion := strings.HasPrefix(rawLink, transclusionPrefix)
	text := strings.TrimPrefix(rawLink, transclusionPrefix)

	inner, ok := extractInnerLinkText(text)
	if !ok {
		return Link{}, ErrInvalidWikilinkFormat
	}

	target, display := splitDisplay(inner)
	filePart, headingPart := splitHeading(target)

	filePart = strings.TrimSpace(filePart)
	headingPart = strings.TrimSpace(headingPart)
	display = strings.TrimSpace(display)

	// Normalize Windows-style path separators to forward slashes.
	filePart = strings.ReplaceAll(filePart, "\\", "/")

	if filePart == "" && headingPart == "" {
		return Link{}, ErrEmptyLinkTarget
	}

	return Link{
		TargetFile:     filePart,
		TargetHeading:  headingPart,
		DisplayText:    display,
		IsTransclusion: isTransclusion,
		LinkType:       Direct,
	}, nil
}

// extractInnerLinkText removes the [[ ]] delimiters from text and returns the
// inner content. It uses the first closing delimiter pair after the opening
// delimiter, matching Obsidian's parsing behavior for nested brackets.
// ok is false if text does not start with [[ or has no closing ]].
func extractInnerLinkText(text string) (inner string, ok bool) {
	if !strings.HasPrefix(text, linkOpenDelimiter) {
		return "", false
	}
	innerStart := len(linkOpenDelimiter)
	idxClose := strings.Index(text[innerStart:], linkCloseDelimiter)
	if idxClose < 0 {
		return "", false
	}
	innerEnd := innerStart + idxClose
	return text[innerStart:innerEnd], true
}

// splitDisplay separates the target part from the optional display text.
// The display text is everything after the first pipe delimiter.
func splitDisplay(inner string) (target, display string) {
	target = inner
	if before, after, ok := strings.Cut(inner, pipeDelimiter); ok {
		target = before
		display = after
	}
	return target, display
}

// splitHeading separates the file part from the optional heading part.
// The heading is everything after the first # delimiter.
func splitHeading(target string) (filePart, headingPart string) {
	filePart = target
	if before, after, ok := strings.Cut(target, headingDelimiter); ok {
		filePart = before
		headingPart = after
	}
	return filePart, headingPart
}

// String returns a human-readable representation of the link for debugging.
func (l Link) String() string {
	targetHeading := ""
	if l.TargetHeading != "" {
		targetHeading = headingPrefix + l.TargetHeading
	}
	return fmt.Sprintf(
		linkStringFormat,
		l.LinkType,
		l.SourceFile, truncateID(l.SourceSegmentID),
		l.TargetFile, truncateID(l.TargetSegmentID),
		targetHeading,
		l.DisplayText,
		l.IsTransclusion,
	)
}

// truncateID shortens long identifiers for debugging output.
func truncateID(id string) string {
	if len(id) > truncatedIDLength {
		return id[:truncatedIDLength] + truncationSuffix
	}
	return id
}
