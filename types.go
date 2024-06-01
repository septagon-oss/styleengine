// Implements: OSS-BB-001 (styleengine building block).
// Per: ADR-0029 (every file declares its purpose); Library Excellence Standard (see README provenance) (Phase 0 exemplar extraction).
// Discipline: C-14.

package styleengine

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Value is the typed RHS of a declaration. Values self-render to CSS.
type Value interface {
	CSS() string
	isValue()
}

// literalValue wraps an opaque CSS literal. Caller is responsible for safety.
type literalValue string

func (v literalValue) CSS() string { return string(v) }
func (literalValue) isValue()      {}

// Literal returns a Value that emits the given string verbatim. Use only with
// trusted input (token values, programmatically built strings). For untrusted
// input go through [Sanitize].
func Literal(s string) Value { return literalValue(s) }

// varRefValue is a var(--name[, fallback]) reference.
type varRefValue struct {
	name     string
	fallback string
}

func (v varRefValue) CSS() string {
	if v.fallback == "" {
		return "var(--" + v.name + ")"
	}
	return "var(--" + v.name + ", " + v.fallback + ")"
}
func (varRefValue) isValue() {}

// VarRef returns a Value referencing a CSS custom property. The leading "--"
// is added automatically. Panics if name is not a valid identifier, or if
// fallback contains characters that could escape the var() expression — the
// unmatched grouping/structural characters `)`, `(`, `{`, `}`, `;`, or a
// comment delimiter `/*`/`*/`. Callers that need a nested var() fallback
// must compose values through the [Builder]/[Value] API rather than embed
// raw CSS through this helper.
func VarRef(name, fallback string) Value {
	if !validVarName(name) {
		panic(fmt.Sprintf("styleengine: invalid var name %q (want [a-z][a-z0-9-]*)", name))
	}
	if err := validateVarFallback(fallback); err != nil {
		panic(fmt.Sprintf("styleengine: invalid VarRef fallback for %q: %v", name, err))
	}
	return varRefValue{name: name, fallback: fallback}
}

// validateVarFallback enforces that a fallback cannot break out of the
// surrounding var() expression or the enclosing rule.
func validateVarFallback(s string) error {
	if s == "" {
		return nil
	}
	if strings.ContainsAny(s, ")(;{}") {
		return fmt.Errorf("contains structural CSS character")
	}
	if strings.Contains(s, "/*") || strings.Contains(s, "*/") {
		return fmt.Errorf("contains comment delimiter")
	}
	return nil
}

// varNameRE matches a CSS custom-property name (the part after "--").
// Intentionally narrower than the CSS Syntax level 3 ident grammar:
//   - lowercase first character ([a-z])
//   - additional [a-z0-9_-] allowed
//
// CSS technically requires dots and other punctuation in idents to be
// escaped (`\2E`); the engine keeps emission ident-clean by rejecting them
// outright at this boundary. Token registries upstream MUST supply
// ident-clean keys — there is no normalization shim.
var varNameRE = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

func validVarName(s string) bool { return varNameRE.MatchString(s) }

// Declaration is property: value [!important].
type Declaration struct {
	Property  string
	Value     Value
	important bool
}

// Decl is a constructor for Declaration.
func Decl(property string, value Value) Declaration {
	return Declaration{Property: property, Value: value}
}

// Important returns a copy of the declaration with !important set.
func (d Declaration) Important() Declaration { d.important = true; return d }

func (d Declaration) CSS() string {
	s := d.Property + ": " + d.Value.CSS()
	if d.important {
		s += " !important"
	}
	return s
}

// Selector is a normalized CSS selector. Use [NewSelector] or [MustSelector].
type Selector struct {
	raw string // normalized form
}

// NewSelector parses and normalizes a selector. Returns error if empty.
func NewSelector(s string) (Selector, error) {
	s = normalizeSelector(s)
	if s == "" {
		return Selector{}, fmt.Errorf("styleengine: empty selector")
	}
	return Selector{raw: s}, nil
}

// MustSelector panics on error.
func MustSelector(s string) Selector {
	sel, err := NewSelector(s)
	if err != nil {
		panic(err)
	}
	return sel
}

func (s Selector) String() string { return s.raw }

func normalizeSelector(s string) string {
	s = strings.TrimSpace(s)
	// collapse runs of whitespace
	s = strings.Join(strings.Fields(s), " ")
	// ensure single space after commas
	parts := strings.Split(s, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return strings.Join(parts, ", ")
}

// Rule is a (selector, []declaration) pair.
type Rule struct {
	Selector Selector
	Decls    []Declaration
}

// CSSPretty emits an indented multi-line form. Custom-property declarations
// (--foo) are emitted first in alphabetical order so that diffs are stable;
// other declarations preserve insertion order.
func (r Rule) CSSPretty() string {
	decls := append([]Declaration(nil), r.Decls...)
	sort.SliceStable(decls, func(i, j int) bool {
		iVar := strings.HasPrefix(decls[i].Property, "--")
		jVar := strings.HasPrefix(decls[j].Property, "--")
		if iVar != jVar {
			return iVar
		}
		if iVar && jVar {
			return decls[i].Property < decls[j].Property
		}
		return false
	})
	var b strings.Builder
	b.WriteString(r.Selector.String())
	b.WriteString(" {")
	for _, d := range decls {
		b.WriteString("\n  ")
		b.WriteString(d.CSS())
		b.WriteByte(';')
	}
	b.WriteString("\n}")
	return b.String()
}
