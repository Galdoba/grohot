package frontmatter

import (
	"errors"
	"strings"
	"testing"

	"github.com/Galdoba/grohot/internal/domain/note/frontmatter/property"
	"github.com/Galdoba/grohot/internal/domain/note/frontmatter/registry"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantProps []*property.Property // expected properties after Parse (before ResolveTypes)
		wantErr   error
	}{
		{
			name:    "empty frontmatter",
			input:   "---\n---\n",
			wantErr: nil,
		},
		{
			name:    "missing opening delimiter",
			input:   "title: hello\n---\n",
			wantErr: errors.New("frontmatter must start with '---'"),
		},
		{
			name:    "missing closing delimiter",
			input:   "---\ntitle: hello\n",
			wantErr: errors.New("frontmatter must end with '---'"),
		},
		{
			name:    "empty line inside",
			input:   "---\ntitle: hello\n\n---\n",
			wantErr: errors.New("empty line inside frontmatter"),
		},
		{
			name:  "single scalar property",
			input: "---\ntitle: My Note\n---\n",
			wantProps: []*property.Property{
				{Name: "title", Type: property.PropertyUndefined, Value: property.Value{Scalar: "My Note"}},
			},
		},
		{
			name:    "duplicate property",
			input:   "---\ntitle: first\ntitle: second\n---\n",
			wantErr: errors.New("duplicate property \"title\""),
		},
		{
			name:  "aliases with empty value (key only)",
			input: "---\naliases:\n---\n",
			wantProps: []*property.Property{
				{Name: "aliases", Type: property.PropertyAliases, Value: property.Value{Scalar: ""}},
			},
		},
		{
			name:  "aliases as list",
			input: "---\naliases:\n  - a\n  - b\n---\n",
			wantProps: []*property.Property{
				{Name: "aliases", Type: property.PropertyAliases, Value: property.Value{List: []string{"a", "b"}}},
			},
		},
		{
			name:  "list property (not aliases)",
			input: "---\ntags:\n  - go\n  - obsidian\n---\n",
			wantProps: []*property.Property{
				{Name: "tags", Type: property.PropertyUndefined, Value: property.Value{List: []string{"go", "obsidian"}}},
			},
		},
		{
			name:  "mixed properties",
			input: "---\nupdated: 2026-07-20T18:18:41.875+10:00\ncount: 7.5\nactive: true\n---\n",
			wantProps: []*property.Property{
				{Name: "updated", Type: property.PropertyUndefined, Value: property.Value{Scalar: "2026-07-20T18:18:41.875+10:00"}},
				{Name: "count", Type: property.PropertyUndefined, Value: property.Value{Scalar: "7.5"}},
				{Name: "active", Type: property.PropertyUndefined, Value: property.Value{Scalar: "true"}},
			},
		},
		{
			name:  "scalar with colon in name",
			input: "---\nkey:value: something\n---\n",
			wantProps: []*property.Property{
				{Name: "key:value", Type: property.PropertyUndefined, Value: property.Value{Scalar: "something"}},
			},
		},
		{
			name:    "list item without parent",
			input:   "---\n  - item\n---\n",
			wantErr: errors.New("list item without parent property"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.input, "\n")
			fm, err := Parse(lines)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(fm.Properties) != len(tt.wantProps) {
				t.Fatalf("property count: got %d, want %d", len(fm.Properties), len(tt.wantProps))
			}
			for i, p := range fm.Properties {
				w := tt.wantProps[i]
				if p.Name != w.Name || p.Type != w.Type || p.Value.Scalar != w.Value.Scalar || !stringSlicesEqual(p.Value.List, w.Value.List) {
					t.Errorf("property %d mismatch:\ngot  %+v\nwant %+v", i, p, w)
				}
			}
		})
	}
}

func stringSlicesEqual(a, b []string) bool {
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

func TestStringRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"simple scalar", "---\ntitle: Hello\n---\n", false},
		{"list", "---\ntags:\n  - a\n  - b\n---\n", false},
		{"aliases list", "---\naliases:\n  - note\n---\n", false},
		{"empty value", "---\ncheckbox:\n---\n", false},
		{"multiple properties", "---\nupdated: 2026-07-20T18:18:41.875+10:00\ncount: 7.5\nactive: true\n---\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.input, "\n")
			fm, err := Parse(lines)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parse error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			output := fm.String()
			fm2, err := Parse(strings.Split(output, "\n"))
			if err != nil {
				t.Fatalf("round-trip parse error: %v", err)
			}
			if len(fm.Properties) != len(fm2.Properties) {
				t.Fatalf("round-trip property count changed: %d -> %d", len(fm.Properties), len(fm2.Properties))
			}
			for i := range fm.Properties {
				a, b := fm.Properties[i], fm2.Properties[i]
				if a.Name != b.Name || a.Value.Scalar != b.Value.Scalar || !stringSlicesEqual(a.Value.List, b.Value.List) {
					t.Errorf("round-trip mismatch at index %d:\n  original: %+v\n  reparse:  %+v", i, a, b)
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		fm      *Frontmatter
		wantErr string
	}{
		{
			name:    "empty frontmatter is valid",
			fm:      New(),
			wantErr: "",
		},
		{
			name: "unnamed property",
			fm: New(
				&property.Property{Name: "", Type: property.PropertyText, Value: property.Value{Scalar: "x"}},
			),
			wantErr: "unnamed property detected: index=0",
		},
		{
			name: "duplicate property",
			fm: New(
				&property.Property{Name: "a", Type: property.PropertyText, Value: property.Value{Scalar: "1"}},
				&property.Property{Name: "a", Type: property.PropertyText, Value: property.Value{Scalar: "2"}},
			),
			wantErr: "property \"a\" is duplicated",
		},
		{
			name: "aliases with wrong type",
			fm: New(
				&property.Property{Name: "aliases", Type: property.PropertyText, Value: property.Value{Scalar: "x"}},
			),
			wantErr: "property \"aliases\": property 'aliases' must have type Aliases",
		},
		{
			name: "TypeAliases used for non-aliases name",
			fm: New(
				&property.Property{Name: "tags", Type: property.PropertyAliases, Value: property.Value{List: []string{"a"}}},
			),
			wantErr: "property \"tags\": type Aliases is reserved for property 'aliases'",
		},
		{
			name: "both scalar and list set",
			fm: New(
				&property.Property{Name: "x", Type: property.PropertyText, Value: property.Value{Scalar: "a", List: []string{"b"}}},
			),
			wantErr: "property \"x\": contains both list and scalar values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fm.Validate()
			if tt.wantErr == "" && err != nil {
				t.Errorf("unexpected error: %v", err)
			} else if tt.wantErr != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErr)
				} else if err.Error() != tt.wantErr {
					t.Errorf("expected error %q, got %q", tt.wantErr, err)
				}
			}
		})
	}
}

func TestResolveTypes(t *testing.T) {
	tests := []struct {
		name string
		in   *property.Property
		want property.Type
	}{
		{"aliases list", &property.Property{Name: "aliases", Value: property.Value{List: []string{"a"}}}, property.PropertyAliases},
		{"plain list", &property.Property{Name: "tags", Value: property.Value{List: []string{"a"}}}, property.PropertyMultitext},
		{"true checkbox", &property.Property{Name: "done", Value: property.Value{Scalar: "true"}}, property.PropertyCheckbox},
		{"false checkbox", &property.Property{Name: "done", Value: property.Value{Scalar: "false"}}, property.PropertyCheckbox},
		{"number integer", &property.Property{Name: "count", Value: property.Value{Scalar: "7"}}, property.PropertyNumber},
		{"number float", &property.Property{Name: "count", Value: property.Value{Scalar: "3.14"}}, property.PropertyNumber},
		{"date ISO", &property.Property{Name: "created", Value: property.Value{Scalar: "2026-07-21"}}, property.PropertyDate},
		{"date with dots", &property.Property{Name: "created", Value: property.Value{Scalar: "21.07.2026"}}, property.PropertyDate},
		{"datetime with time", &property.Property{Name: "updated", Value: property.Value{Scalar: "2026-07-21T12:34:56"}}, property.PropertyDateTime},
		{"datetime with timezone", &property.Property{Name: "updated", Value: property.Value{Scalar: "2026-07-21T12:34:56+03:00"}}, property.PropertyDateTime},
		{"text", &property.Property{Name: "desc", Value: property.Value{Scalar: "Hello, world!"}}, property.PropertyText},
		{"ambiguous empty scalar", &property.Property{Name: "checkbox", Value: property.Value{Scalar: ""}}, property.PropertyUndefined},
		{"ambiguous empty list", &property.Property{Name: "list", Value: property.Value{List: []string{}}}, property.PropertyMultitext},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := property.ResolvePropertyType(tt.in)
			if got != tt.want {
				t.Errorf("ResolvePropertyType(%+v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSetTypesFromRegistry(t *testing.T) {
	fm := New(
		&property.Property{Name: "a", Type: property.PropertyUndefined, Value: property.Value{Scalar: "x"}},
		&property.Property{Name: "b", Type: property.PropertyText, Value: property.Value{Scalar: "y"}},
	)
	reg := registry.Registry{"a": property.PropertyNumber, "b": property.PropertyCheckbox}
	fm.SetTypesFromRegistry(reg)
	if got := fm.Properties[0].Type; got != property.PropertyNumber {
		t.Errorf("property a: got %q, want %q", got, property.PropertyNumber)
	}
	if got := fm.Properties[1].Type; got != property.PropertyText {
		t.Errorf("property b type overwritten: got %q, want %q", got, property.PropertyText)
	}
}

func TestValidateTypesAgainstRegistry(t *testing.T) {
	tests := []struct {
		name    string
		fm      *Frontmatter
		reg     registry.Registry
		wantErr string
	}{
		{
			name: "all match",
			fm: New(
				&property.Property{Name: "a", Type: property.PropertyNumber},
			),
			reg:     registry.Registry{"a": property.PropertyNumber},
			wantErr: "",
		},
		{
			name: "mismatch",
			fm: New(
				&property.Property{Name: "a", Type: property.PropertyText},
			),
			reg:     registry.Registry{"a": property.PropertyNumber},
			wantErr: "property \"a\": type mismatch: expected \"number\" (registry), got \"text\"",
		},
		{
			name: "undefined ignored",
			fm: New(
				&property.Property{Name: "a", Type: property.PropertyUndefined},
			),
			reg:     registry.Registry{"a": property.PropertyNumber},
			wantErr: "",
		},
		{
			name:    "no matching registry key",
			fm:      New(&property.Property{Name: "a", Type: property.PropertyText}),
			reg:     registry.Registry{},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fm.ValidateTypesAgainstRegistry(tt.reg)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || err.Error() != tt.wantErr {
				t.Errorf("expected error %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestApplyRegistry(t *testing.T) {
	fm := New(
		&property.Property{Name: "a", Type: property.PropertyUndefined},
		&property.Property{Name: "b", Type: property.PropertyText},
	)
	reg := registry.Registry{"a": property.PropertyNumber, "b": property.PropertyCheckbox}
	err := fm.ApplyRegistry(reg)
	if err == nil {
		t.Fatalf("expected error due to type mismatch for b")
	}
	if err.Error() != "property \"b\": type mismatch: expected \"checkbox\" (registry), got \"text\"" {
		t.Errorf("error message mismatch: %v", err)
	}
	if fm.Properties[0].Type != property.PropertyNumber {
		t.Errorf("property a: got %q, want number", fm.Properties[0].Type)
	}
}

func TestPropertyControl(t *testing.T) {
	fm := New(
		&property.Property{Name: "title", Type: property.PropertyText, Value: property.Value{Scalar: "hello"}},
		&property.Property{Name: "count", Type: property.PropertyNumber, Value: property.Value{Scalar: "5"}},
		&property.Property{Name: "tags", Type: property.PropertyMultitext, Value: property.Value{List: []string{"a", "b"}}},
	)

	t.Run("Get", func(t *testing.T) {
		p := fm.Get("count")
		if p == nil || p.Name != "count" {
			t.Errorf("Get failed")
		}
		if fm.Get("nonexistent") != nil {
			t.Error("Get should return nil for missing property")
		}
	})

	t.Run("Has", func(t *testing.T) {
		if !fm.Has("title") || fm.Has("nope") {
			t.Error("Has failed")
		}
	})

	t.Run("IndexOf", func(t *testing.T) {
		if idx := fm.IndexOf("tags"); idx != 2 {
			t.Errorf("IndexOf tags: got %d, want 2", idx)
		}
		if idx := fm.IndexOf("missing"); idx != -1 {
			t.Errorf("IndexOf missing: got %d, want -1", idx)
		}
	})

	t.Run("Set (new property)", func(t *testing.T) {
		newProp := &property.Property{Name: "new", Type: property.PropertyText, Value: property.Value{Scalar: "val"}}
		idx := fm.Set(newProp)
		if idx != 3 || len(fm.Properties) != 4 || fm.Properties[3].Name != "new" {
			t.Errorf("Set new failed")
		}
		replace := &property.Property{Name: "title", Type: property.PropertyText, Value: property.Value{Scalar: "replaced"}}
		idx2 := fm.Set(replace)
		if idx2 != 0 || fm.Properties[0].Value.Scalar != "replaced" {
			t.Errorf("Set replace failed")
		}
	})

	t.Run("Insert", func(t *testing.T) {
		fm := New(
			&property.Property{Name: "a", Type: property.PropertyText, Value: property.Value{Scalar: "1"}},
			&property.Property{Name: "b", Type: property.PropertyText, Value: property.Value{Scalar: "2"}},
			&property.Property{Name: "c", Type: property.PropertyText, Value: property.Value{Scalar: "3"}},
		)
		p := &property.Property{Name: "b", Type: property.PropertyText, Value: property.Value{Scalar: "new_b"}}
		idx := fm.Insert(0, p)
		if idx != 0 {
			t.Errorf("Insert index: got %d, want 0", idx)
		}
		if len(fm.Properties) != 3 {
			t.Fatalf("expected 3 properties, got %d", len(fm.Properties))
		}
		if fm.Properties[0].Name != "b" || fm.Properties[0].Value.Scalar != "new_b" {
			t.Errorf("Insert failed: first property wrong")
		}
		if fm.Properties[1].Name != "a" || fm.Properties[2].Name != "c" {
			t.Errorf("Insert order wrong")
		}
		p2 := &property.Property{Name: "d", Type: property.PropertyText, Value: property.Value{Scalar: "4"}}
		idx2 := fm.Insert(10, p2)
		if idx2 != 3 || fm.Properties[3].Name != "d" {
			t.Errorf("Insert at end: got idx=%d, last=%s", idx2, fm.Properties[3].Name)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		fm := New(
			&property.Property{Name: "a", Type: property.PropertyText},
			&property.Property{Name: "b", Type: property.PropertyText},
			&property.Property{Name: "c", Type: property.PropertyText},
		)
		idx := fm.Delete("b")
		if idx != 1 {
			t.Errorf("Delete returned index %d, want 1", idx)
		}
		if len(fm.Properties) != 2 || fm.Properties[0].Name != "a" || fm.Properties[1].Name != "c" {
			t.Error("Delete failed")
		}
		idx = fm.Delete("nonexistent")
		if idx != -1 {
			t.Errorf("Delete non-existent should return -1, got %d", idx)
		}
	})

	t.Run("SwapByName", func(t *testing.T) {
		fm := New(
			&property.Property{Name: "a"},
			&property.Property{Name: "b"},
			&property.Property{Name: "c"},
		)
		err := fm.SwapByName("a", "c")
		if err != nil {
			t.Fatal(err)
		}
		names := []string{fm.Properties[0].Name, fm.Properties[1].Name, fm.Properties[2].Name}
		if names[0] != "c" || names[2] != "a" {
			t.Errorf("SwapByName failed: %v", names)
		}
		err = fm.SwapByName("a", "x")
		if err == nil {
			t.Error("expected error for missing property")
		}
	})

	t.Run("SwapByIndex", func(t *testing.T) {
		fm := New(
			&property.Property{Name: "a"},
			&property.Property{Name: "b"},
			&property.Property{Name: "c"},
		)
		err := fm.SwapByIndex(0, 2)
		if err != nil {
			t.Fatal(err)
		}
		if fm.Properties[0].Name != "c" || fm.Properties[2].Name != "a" {
			t.Error("SwapByIndex failed")
		}
		err = fm.SwapByIndex(-1, 0)
		if err == nil {
			t.Error("expected error for negative index")
		}
		err = fm.SwapByIndex(0, 5)
		if err == nil {
			t.Error("expected error for out-of-bounds index")
		}
	})
}
