package frontmatter

import (
	"fmt"

	"github.com/Galdoba/grohot/internal/domain/note/frontmatter/property"
)

func (f *Frontmatter) Get(name string) *property.Property {
	for _, p := range f.Properties {
		if p.Name == name {
			return p
		}
	}
	return nil
}

func (f *Frontmatter) Has(name string) bool {
	for _, p := range f.Properties {
		if p.Name == name {
			return true
		}
	}
	return false
}

func (f *Frontmatter) IndexOf(name string) int {
	for i, p := range f.Properties {
		if p.Name == name {
			return i
		}
	}
	return -1
}

func (f *Frontmatter) Set(p *property.Property) int {
	for i, has := range f.Properties {
		if p.Name == has.Name {
			f.Properties[i] = p
			return i
		}
	}
	f.Properties = append(f.Properties, p)
	return len(f.Properties) - 1
}

func (f *Frontmatter) Insert(index int, p *property.Property) int {
	f.Delete(p.Name)
	index = minmax(index, 0, len(f.Properties))

	f.Properties = append(f.Properties, &property.Property{})
	copy(f.Properties[index+1:], f.Properties[index:])
	f.Properties[index] = p

	return index
}

func minmax(i, min, max int) int {
	if max < min {
		panic("minmax: min > max")
	}
	if i < min {
		return min
	}
	if i > max {
		return max
	}
	return i
}

func (f *Frontmatter) Delete(name string) int {
	deleted := -1
	filtered := f.Properties[:0]
	for i, prop := range f.Properties {
		if prop.Name != name {
			filtered = append(filtered, prop)
		} else {
			deleted = i
		}
	}
	f.Properties = filtered
	return deleted
}

func (f *Frontmatter) SwapByName(name1, name2 string) error {
	if !f.Has(name1) {
		return fmt.Errorf("property %q is absent", name1)
	}
	if !f.Has(name2) {
		return fmt.Errorf("property %q is absent", name2)
	}

	f.Properties[f.IndexOf(name1)], f.Properties[f.IndexOf(name2)] = f.Properties[f.IndexOf(name2)], f.Properties[f.IndexOf(name1)]
	return nil
}

func (f *Frontmatter) SwapByIndex(idx1, idx2 int) error {
	if idx1 < 0 {
		return fmt.Errorf("idx1 is to low")
	}
	if idx2 < 0 {
		return fmt.Errorf("idx2 is to low")
	}
	if idx1 > len(f.Properties)-1 {
		return fmt.Errorf("idx1 is to high")
	}
	if idx2 > len(f.Properties)-1 {
		return fmt.Errorf("idx2 is to high")
	}

	f.Properties[idx1], f.Properties[idx2] = f.Properties[idx2], f.Properties[idx1]
	return nil
}
