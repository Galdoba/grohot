package registry

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"sort"

	"github.com/Galdoba/grohot/internal/domain/note/frontmatter/property"
)

// JSON keys and formatting constants used for serialization.
const (
	jsonKeyTypes  = "types"
	defaultIndent = "  "
	// filePerm defines the default file permission for saved registry files.
	filePerm = 0666
)

// Registry represents a mapping of property names to their types,
// compatible with Obsidian's frontmatter property registry structure.
// It is safe for concurrent reads, but not for concurrent writes.
type Registry map[string]property.Type

// Load reads a property registry from a JSON file.
// The expected JSON structure is an object with a single "types" key,
// whose value is an object mapping property names to type strings, e.g.:
//
//	{"types": {"aliases": "aliases", "tags": "tags"}}
func Load(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read registry file: %w", err)
	}

	var wrapper struct {
		Types Registry `json:"types"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("parse registry JSON: %w", err)
	}

	return &wrapper.Types, nil
}

// Save serialises the registry to a JSON file with the standard Obsidian wrapper
// {"types": {...}}.
func (reg Registry) Save(path string) error {
	wrapper := struct {
		Types Registry `json:"types"`
	}{Types: reg}

	data, err := json.MarshalIndent(wrapper, "", defaultIndent)
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}

	if err := os.WriteFile(path, data, filePerm); err != nil {
		return fmt.Errorf("write registry file: %w", err)
	}
	return nil
}

// Get returns the property type registered for the given name.
// The boolean indicates whether the name was found.
func (reg Registry) Get(name string) (property.Type, bool) {
	typ, ok := reg[name]
	return typ, ok
}

// Set assigns a property type to the given name, overwriting any existing entry.
func (reg Registry) Set(name string, pt property.Type) {
	reg[name] = pt
}

// Delete removes the given property name from the registry.
// It is a no-op if the name does not exist.
func (reg Registry) Delete(name string) {
	delete(reg, name)
}

// Has reports whether the registry contains the given property name.
func (reg Registry) Has(name string) bool {
	_, ok := reg[name]
	return ok
}

// Names returns a sorted slice of all property names currently in the registry.
func (reg Registry) Names() []string {
	names := make([]string, 0, len(reg))
	for name := range reg {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Merge copies all entries from another registry into this one.
// Existing entries with the same name will be overwritten.
func (reg Registry) Merge(other Registry) {
	maps.Copy(reg, other)
}
