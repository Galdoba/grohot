// Package frontmatter provides parsing, type resolution and validation
// for Obsidian-style YAML frontmatter.
package frontmatter

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	delimiter      = "---"
	listItemPrefix = "  - "
	keyValueSep    = ": "
	keyOnlySuffix  = ":"
	aliasesName    = "aliases"
)

// PropertyType represents the type of a frontmatter property.
type PropertyType string

const (
	TypeAliases   PropertyType = "Aliases"
	TypeCheckbox  PropertyType = "Checkbox"
	TypeDate      PropertyType = "Date"
	TypeDateTime  PropertyType = "Date & Time"
	TypeList      PropertyType = "List"
	TypeNumber    PropertyType = "Number"
	TypeText      PropertyType = "Text"
	TypeUndefined PropertyType = "Undefined"
)

// Immutable: safe for concurrent read access.
var dateFormats = []string{
	"2006-01-02",
	"2006-1-2",
	"02.01.2006",
	"2.1.2006",
	"01/02/2006",
	"1/2/2006",
}

// Immutable: safe for concurrent read access.
var dateTimeFormats = []string{
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	time.RFC3339,
	time.RFC3339Nano,
}

// ---------------------------------------------------------------------------
// Frontmatter
// ---------------------------------------------------------------------------

// Frontmatter represents an ordered set of note properties.
type Frontmatter struct {
	Properties []Property
}

// New creates a new Frontmatter with the given properties.
func New(pr ...Property) *Frontmatter {
	f := &Frontmatter{}
	f.Properties = append(f.Properties, pr...)
	return f
}

// Validate checks the syntactic invariants of all properties, including their values.
func (f *Frontmatter) Validate() error {
	for i := range f.Properties {
		if err := f.Properties[i].validate(); err != nil {
			return fmt.Errorf("property %q: %w", f.Properties[i].Name, err)
		}
	}
	return nil
}

// SetTypesFromRegistry assigns types from the registry to all properties
// that currently have TypeUndefined. Properties not in the registry are
// left untouched. This method never returns an error.
func (f *Frontmatter) SetTypesFromRegistry(types map[string]PropertyType) {
	for i := range f.Properties {
		p := &f.Properties[i]
		t, ok := types[p.Name]
		if !ok {
			continue
		}
		if p.Type == TypeUndefined {
			p.Type = t
		}
	}
}

// ValidateTypesAgainstRegistry ensures that every property that has a
// type other than Undefined matches the type recorded in the registry
// for that name. Properties not found in the registry are ignored.
func (f *Frontmatter) ValidateTypesAgainstRegistry(types map[string]PropertyType) error {
	for _, p := range f.Properties {
		t, ok := types[p.Name]
		if !ok {
			continue
		}
		if p.Type != TypeUndefined && p.Type != t {
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
func (f *Frontmatter) ApplyRegistry(types map[string]PropertyType) error {
	f.SetTypesFromRegistry(types)
	return f.ValidateTypesAgainstRegistry(types)
}

// ResolveTypes automatically determines concrete types for every property
// that still has TypeUndefined.
func (f *Frontmatter) ResolveTypes() {
	for i := range f.Properties {
		if f.Properties[i].Type == TypeUndefined {
			f.Properties[i].Type = ResolvePropertyType(&f.Properties[i])
		}
	}
}

// ---------------------------------------------------------------------------
// Type resolution
// ---------------------------------------------------------------------------

// ResolvePropertyType returns the concrete type of a property when it can be
// unambiguously inferred from its value. Otherwise it returns TypeUndefined.
func ResolvePropertyType(p *Property) PropertyType {
	if tryAliases(p) {
		return TypeAliases
	}
	if tryList(p) {
		return TypeList
	}
	if tryCheckbox(p) {
		return TypeCheckbox
	}
	if tryNumber(p) {
		return TypeNumber
	}
	if tryDate(p) {
		return TypeDate
	}
	if tryDateTime(p) {
		return TypeDateTime
	}
	if tryText(p) {
		return TypeText
	}
	return TypeUndefined
}

// --- Unambiguous type predicates ---

func tryAliases(p *Property) bool {
	return p.Name == aliasesName && p.Value.List != nil && p.Value.Scalar == ""
}

func tryList(p *Property) bool {
	return p.Name != aliasesName && p.Value.List != nil && p.Value.Scalar == ""
}

func tryCheckbox(p *Property) bool {
	if len(p.Value.List) > 0 {
		return false
	}
	s := strings.ToLower(p.Value.Scalar)
	return s == "true" || s == "false"
}

func tryNumber(p *Property) bool {
	if len(p.Value.List) > 0 {
		return false
	}
	f, err := strconv.ParseFloat(p.Value.Scalar, 64)
	if err != nil {
		return false
	}
	if math.IsInf(f, 64) || math.IsNaN(f) {
		return false
	}
	return err == nil
}

func tryDate(p *Property) bool {
	if len(p.Value.List) > 0 {
		return false
	}
	if p.Value.Scalar == "" {
		return false
	}
	s := p.Value.Scalar
	for _, layout := range dateFormats {
		t, err := time.Parse(layout, s)
		if err != nil {
			continue
		}
		// Check that the parsed time has no time-of-day component.
		y, m, d := t.Date()
		if t.Equal(time.Date(y, m, d, 0, 0, 0, 0, t.Location())) {
			return true
		}
	}
	return false
}

func tryDateTime(p *Property) bool {
	if p.Value.List != nil || p.Value.Scalar == "" {
		return false
	}
	s := p.Value.Scalar
	for _, layout := range dateTimeFormats {
		t, err := time.Parse(layout, s)
		if err != nil {
			continue
		}
		// A date-time must contain a non-zero time part.
		y, m, d := t.Date()
		if !t.Equal(time.Date(y, m, d, 0, 0, 0, 0, t.Location())) {
			return true
		}
	}
	return false
}

func tryText(p *Property) bool {
	if p.Value.List != nil || p.Value.Scalar == "" {
		return false
	}
	// Text is the fallback: it must not be any other known scalar type.
	if tryCheckbox(p) || tryNumber(p) || tryDate(p) || tryDateTime(p) {
		return false
	}
	return true
}

// ---------------------------------------------------------------------------
// Parsing
// ---------------------------------------------------------------------------

// Parse extracts a YAML frontmatter block from a slice of lines (including
// the surrounding delimiter lines). Properties are created with TypeAliases
// for the "aliases" key, and TypeUndefined for everything else.
func Parse(lines []string) (*Frontmatter, error) {
	body, err := findDelimiters(lines)
	if err != nil {
		return nil, err
	}
	props, err := parseBody(body)
	if err != nil {
		return nil, err
	}
	return &Frontmatter{Properties: props}, nil
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
func parseBody(body []string) ([]Property, error) {
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
			idx := strings.Index(line, keyValueSep)
			name := line[:idx]
			scalar := line[idx+len(keyValueSep):]
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
	properties []Property
	seen       map[string]bool
}

// appendListItem adds an item to the last property, converting it to a list if
// necessary.
func (s *parseState) appendListItem(item string) error {
	if len(s.properties) == 0 {
		return errors.New("list item without parent property")
	}
	last := &s.properties[len(s.properties)-1]
	// Convert an empty scalar property into a list property.
	if last.Value.List == nil && last.Value.Scalar == "" {
		last.Value = PropertyValue{List: []string{}}
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
	propType := TypeUndefined
	if name == aliasesName {
		propType = TypeAliases
	}
	s.properties = append(s.properties, Property{
		Name: name,
		Type: propType,
		Value: PropertyValue{
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
	propType := TypeUndefined
	if name == aliasesName {
		propType = TypeAliases
	}
	// Always created as an empty scalar; subsequent list items will convert it.
	s.properties = append(s.properties, Property{
		Name: name,
		Type: propType,
		Value: PropertyValue{
			Scalar: "",
		},
	})
	return nil
}

// ---------------------------------------------------------------------------
// String representation
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Property types
// ---------------------------------------------------------------------------

// Property is a single frontmatter key-value pair with an associated type.
type Property struct {
	Name  string        `json:"name"`
	Type  PropertyType  `json:"type"`
	Value PropertyValue `json:"value"`
}

// validate checks the invariants of a property, including its value.
func (p *Property) validate() error {
	if p.Name == "" {
		return fmt.Errorf("name must not be empty")
	}
	if p.Name == aliasesName && p.Type != TypeAliases {
		return fmt.Errorf("property 'aliases' must have type Aliases")
	}
	if p.Type == TypeAliases && p.Name != aliasesName {
		return fmt.Errorf("type Aliases is reserved for property 'aliases'")
	}
	return p.Value.validate()
}

// PropertyValue holds the value of a property – either a scalar or a list of
// strings.
type PropertyValue struct {
	Scalar string   `json:"scalar,omitempty"`
	List   []string `json:"list,omitempty"`
}

// validate ensures that scalar and list are not both set.
func (pv *PropertyValue) validate() error {
	if len(pv.List) > 0 && pv.Scalar != "" {
		return errors.New("both list and scalar values are set")
	}
	return nil
}
