// Package frontmatter provides parsing, type resolution and validation
// for Obsidian-style YAML frontmatter.
package frontmatter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Galdoba/grohot/internal/domain/note/frontmatter/property"
	"github.com/Galdoba/grohot/internal/domain/note/frontmatter/registry"
)

const (
	delimiter      = "---"
	listItemPrefix = "  - "
	keyValueSep    = ": "
	keyOnlySuffix  = ":"
	aliasesName    = "aliases"
)

// Frontmatter represents an ordered set of note properties.
type Frontmatter struct {
	Properties []*property.Property
}

// New creates a new Frontmatter with the given properties.
func New(pr ...*property.Property) *Frontmatter {
	f := &Frontmatter{}
	f.Properties = append(f.Properties, pr...)
	return f
}

// Validate checks the syntactic invariants of all properties, including their values.
func (f *Frontmatter) Validate() error {
	seen := make(map[string]bool)

	for i := range f.Properties {
		name := f.Properties[i].Name
		if name == "" {
			return fmt.Errorf("unnamed property detected: index=%d", i)
		}
		if seen[name] {
			return fmt.Errorf("property %q is duplicated", name)
		}
		if err := f.Properties[i].Validate(); err != nil {
			return fmt.Errorf("property %q: %w", f.Properties[i].Name, err)
		}
		seen[name] = true

	}
	return nil
}

// SetTypesFromRegistry assigns types from the registry to all properties
// that currently have TypeUndefined. Properties not in the registry are
// left untouched. This method never returns an error.
func (f *Frontmatter) SetTypesFromRegistry(types registry.Registry) {
	for i := range f.Properties {
		p := f.Properties[i]
		t, ok := types[p.Name]
		if !ok {
			continue
		}
		if p.Type == property.PropertyUndefined {
			p.Type = t
		}
	}
}

// ValidateTypesAgainstRegistry ensures that every property that has a
// type other than Undefined matches the type recorded in the registry
// for that name. Properties not found in the registry are ignored.
func (f *Frontmatter) ValidateTypesAgainstRegistry(types registry.Registry) error {
	for _, p := range f.Properties {
		t, ok := types[p.Name]
		if !ok {
			continue
		}
		if p.Type != property.PropertyUndefined && p.Type != t {
			return fmt.Errorf("property %q: type mismatch: expected %q (registry), got %q", p.Name, t, p.Type)
		}
	}
	return nil
}

// ApplyRegistry is a convenience wrapper that first sets types from the
// registry for Undefined properties, and then validates all non‑Undefined
// types against the registry. It is kept for backward compatibility;
// prefer using SetTypesFromRegistry and ValidateTypesAgainstRegistry
// separately for finer control.
func (f *Frontmatter) ApplyRegistry(types registry.Registry) error {
	f.SetTypesFromRegistry(types)
	return f.ValidateTypesAgainstRegistry(types)
}

// ResolveTypes automatically determines concrete types for every property
// that still has TypeUndefined.
func (f *Frontmatter) ResolveTypes() {
	for i := range f.Properties {
		if f.Properties[i].Type == property.PropertyUndefined {
			f.Properties[i].Type = property.ResolvePropertyType(f.Properties[i])
		}
	}
}

// Parse extracts a YAML frontmatter block from a slice of lines (including
// the surrounding delimiter lines). Properties are created with TypeAliases
// for the "aliases" key, and TypeUndefined for everything else.
// func Parse(lines []string) (*Frontmatter, error) {
// 	body, err := findDelimiters(lines)
// 	if err != nil {
// 		return nil, err
// 	}
// 	props, err := parseBody(body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &Frontmatter{Properties: props}, nil
// }

// Parse extracts YAML frontmatter from a slice of lines.
// It returns the parsed Frontmatter, the remaining lines after the frontmatter,
// and an error if any.
// If no frontmatter is present, it returns (nil, allLines, nil).
func Parse(lines []string) (*Frontmatter, []string, error) {
	start, end, err := findFrontmatterRange(lines)
	if err != nil {
		return nil, nil, err
	}
	if start == -1 {
		return nil, lines, nil
	}
	body := lines[start+1 : end]
	props, err := parseBody(body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse frontmatter body: %w", err)
	}
	fm := &Frontmatter{Properties: props}
	rest := lines[end+1:]
	return fm, rest, nil
}

// findFrontmatterRange returns the start and end indices of the frontmatter
// delimiters ("---"). Returns (-1, -1, nil) if no frontmatter is found.
// Returns an error if an opening delimiter is found but no closing delimiter.
func findFrontmatterRange(lines []string) (start, end int, err error) {
	if len(lines) == 0 {
		return -1, -1, nil
	}
	// Obsidian usually has no leading spaces, but for safety reasons.
	if strings.TrimSpace(lines[0]) != "---" {
		return -1, -1, nil
	}
	start = 0
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			return start, end, nil
		}
	}
	return -1, -1, fmt.Errorf("unclosed frontmatter: opening delimiter found but no closing delimiter")
}

// findDelimiters locates the opening and closing delimiter lines and returns
// the lines between them.
func findDelimiters(lines []string) ([]string, error) {
	if len(lines) < 2 || lines[0] != delimiter {
		return nil, errors.New("frontmatter must start with '---'")
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == delimiter {
			end = i
			break
		}
	}
	if end == -1 {
		return nil, errors.New("frontmatter must end with '---'")
	}
	return lines[1:end], nil
}

// parseBody processes the inner lines of the frontmatter and builds the
// property list.
func parseBody(body []string) ([]*property.Property, error) {
	state := &parseState{
		seen: make(map[string]bool),
	}
	for _, line := range body {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			return nil, errors.New("empty line inside frontmatter")
		}
		switch {
		case strings.HasPrefix(line, listItemPrefix):
			item := strings.TrimPrefix(line, listItemPrefix)
			if err := state.appendListItem(item); err != nil {
				return nil, err
			}
		case strings.Contains(line, keyValueSep):
			before, after, _ := strings.Cut(line, keyValueSep)
			name := before
			scalar := after
			if err := state.addScalarProperty(name, scalar); err != nil {
				return nil, err
			}
		case strings.HasSuffix(line, keyOnlySuffix):
			name := strings.TrimSuffix(line, keyOnlySuffix)
			if err := state.addKeyOnlyProperty(name); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("invalid line: %q", line)
		}
	}
	return state.properties, nil
}

// parseState holds the mutable parsing state.
type parseState struct {
	properties []*property.Property
	seen       map[string]bool
}

// appendListItem adds an item to the last property, converting it to a list if
// necessary.
func (s *parseState) appendListItem(item string) error {
	if len(s.properties) == 0 {
		return errors.New("list item without parent property")
	}
	last := s.properties[len(s.properties)-1]
	// Convert an empty scalar property into a list property.
	if last.Value.List == nil && last.Value.Scalar == "" {
		last.Value = property.Value{List: []string{}}
	}
	if last.Value.List == nil {
		return fmt.Errorf("list item for scalar property %q", last.Name)
	}
	last.Value.List = append(last.Value.List, item)
	return nil
}

// addScalarProperty registers a key: value pair, failing on duplicates.
func (s *parseState) addScalarProperty(name, scalar string) error {
	if s.seen[name] {
		return fmt.Errorf("duplicate property %q", name)
	}
	s.seen[name] = true
	propType := property.PropertyUndefined
	if name == aliasesName {
		propType = property.PropertyAliases
	}
	s.properties = append(s.properties, &property.Property{
		Name: name,
		Type: propType,
		Value: property.Value{
			Scalar: scalar,
		},
	})
	return nil
}

// addKeyOnlyProperty registers a key with no value (yet), e.g. "tags:".
func (s *parseState) addKeyOnlyProperty(name string) error {
	if name == "" {
		return errors.New("empty property name")
	}
	if s.seen[name] {
		return fmt.Errorf("duplicate property %q", name)
	}
	s.seen[name] = true
	propType := property.PropertyUndefined
	if name == aliasesName {
		propType = property.PropertyAliases
	}
	// Always created as an empty scalar; subsequent list items will convert it.
	s.properties = append(s.properties, &property.Property{
		Name: name,
		Type: propType,
		Value: property.Value{
			Scalar: "",
		},
	})
	return nil
}

// String returns the YAML representation of the frontmatter, including the
// surrounding delimiter lines.
func (f *Frontmatter) String() string {
	var b strings.Builder
	b.WriteString(delimiter)
	b.WriteByte('\n')
	for _, p := range f.Properties {
		if p.Value.List != nil {
			b.WriteString(p.Name)
			b.WriteString(keyOnlySuffix)
			b.WriteByte('\n')
			for _, item := range p.Value.List {
				b.WriteString(listItemPrefix)
				b.WriteString(item)
				b.WriteByte('\n')
			}
		} else {
			b.WriteString(p.Name)
			b.WriteString(keyValueSep)
			b.WriteString(p.Value.Scalar)
			b.WriteByte('\n')
		}
	}
	b.WriteString(delimiter)
	b.WriteByte('\n')
	return b.String()
}

// ParseBytes is a Parse convinience wrapper.
func ParseBytes(data []byte) (*Frontmatter, error) {
	fm, _, err := Parse(strings.Split(string(data), "\n"))
	return fm, err
}

// LineCount is a convinience func that tells which line in file is last for frontmatter.
func (f *Frontmatter) LineCount() int {
	count := 2
	for _, p := range f.Properties {
		if p.Value.List != nil {
			count++                    // key line
			count += len(p.Value.List) // list elements
		} else {
			count++
		}
	}
	return count
}
