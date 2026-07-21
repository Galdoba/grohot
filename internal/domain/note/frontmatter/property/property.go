package property

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

const (
	AliasesName = "aliases"
)

// Type represents the type of a frontmatter property.
type Type string

const (
	PropertyAliases   Type = "aliases"
	PropertyCheckbox  Type = "checkbox"
	PropertyDate      Type = "date"
	PropertyDateTime  Type = "datetime"
	PropertyMultitext Type = "multitext"
	PropertyNumber    Type = "number"
	PropertyText      Type = "text"
	PropertyUndefined Type = "undefined"
)

// Property is a single frontmatter key-value pair with an associated type.
type Property struct {
	Name  string `json:"name"`
	Type  Type   `json:"type"`
	Value Value  `json:"value"`
}

// Validate checks the invariants of a property, including its value.
func (p *Property) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("name must not be empty")
	}
	if p.Name == AliasesName && p.Type != PropertyAliases {
		return fmt.Errorf("property 'aliases' must have type Aliases")
	}
	if p.Type == PropertyAliases && p.Name != AliasesName {
		return fmt.Errorf("type Aliases is reserved for property 'aliases'")
	}
	return p.Value.validate()
}

// PropertyValue holds the value of a property – either a scalar or a list of
// strings.
type Value struct {
	Scalar string   `json:"scalar,omitempty"`
	List   []string `json:"list,omitempty"`
}

// validate ensures that scalar and list are not both set.
func (pv *Value) validate() error {
	if len(pv.List) > 0 && pv.Scalar != "" {
		return errors.New("contains both list and scalar values")
	}
	return nil
}

var dateFormats = []string{
	"2006-01-02",
	"2006-1-2",
	"02.01.2006",
	"2.1.2006",
	"01/02/2006",
	"1/2/2006",
}

var dateTimeFormats = []string{
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	time.RFC3339,
	time.RFC3339Nano,
}

// ResolvePropertyType returns the concrete type of a property when it can be
// unambiguously inferred from its value. Otherwise it returns TypeUndefined.
func ResolvePropertyType(p *Property) Type {
	if tryAliases(p) {
		return PropertyAliases
	}
	if tryList(p) {
		return PropertyMultitext
	}
	if tryCheckbox(p) {
		return PropertyCheckbox
	}
	if tryNumber(p) {
		return PropertyNumber
	}
	if tryDate(p) {
		return PropertyDate
	}
	if tryDateTime(p) {
		return PropertyDateTime
	}
	if tryText(p) {
		return PropertyText
	}
	return PropertyUndefined
}

func tryAliases(p *Property) bool {
	return p.Name == AliasesName && p.Value.List != nil && p.Value.Scalar == ""
}

func tryList(p *Property) bool {
	return p.Name != AliasesName && p.Value.List != nil && p.Value.Scalar == ""
}

func tryCheckbox(p *Property) bool {
	if p.Value.List != nil {
		return false
	}
	return p.Value.Scalar == "true" || p.Value.Scalar == "false"
}

func tryNumber(p *Property) bool {
	if p.Value.List != nil {
		return false
	}
	f, err := strconv.ParseFloat(p.Value.Scalar, 64)
	if err != nil {
		return false
	}
	if math.IsInf(f, 64) || math.IsNaN(f) {
		return false
	}
	return true
}

func tryDate(p *Property) bool {
	if p.Value.List != nil || p.Value.Scalar == "" {
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
	return true
}
