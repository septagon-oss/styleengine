// variables.go owns custom-property declaration and reference diagnostics.
// Implements: REQ-011.
// Per: ADR-0029 (every file declares its purpose).
// Discipline: C-14.

package styleengine

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Var registers a CSS custom property to be emitted in :root. Names must
// match [a-z][a-z0-9-]*. The leading "--" is automatic. Re-registering a name
// is last-write-wins.
func (s *Sheet) Var(name, value string) *Sheet {
	if !validVarName(name) {
		panic(fmt.Sprintf("styleengine: invalid var name %q", name))
	}
	s.ensureRoot()
	rootIdx := s.ruleIdx[":root"]
	root := s.rules[rootIdx]
	prop := "--" + name
	for i, d := range root.Decls {
		if d.Property == prop {
			root.Decls[i] = Decl(prop, Literal(value))
			return s
		}
	}
	root.Decls = append(root.Decls, Decl(prop, Literal(value)))
	return s
}

func (s *Sheet) ensureRoot() {
	if _, ok := s.ruleIdx[":root"]; ok {
		return
	}
	s.rules = append(s.rules, &Rule{Selector: MustSelector(":root")})
	s.ruleIdx[":root"] = len(s.rules) - 1
}

// RegisteredVars returns the sorted list of custom-property names registered
// via Var (without the "--" prefix).
func (s *Sheet) RegisteredVars() []string {
	if s == nil {
		return nil
	}
	idx, ok := s.ruleIdx[":root"]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(s.rules[idx].Decls))
	for _, d := range s.rules[idx].Decls {
		if after, ok0 := strings.CutPrefix(d.Property, "--"); ok0 {
			out = append(out, after)
		}
	}
	sort.Strings(out)
	return out
}

// UndefinedVars returns var(--x) references in the Sheet that have no
// matching Var registration. Output is sorted. The scan covers:
//
//   - Top-level rules (excluding the :root rule itself)
//   - Nested rules inside [@media], [@layer], [@supports] at-rules
//   - Keyframe-stop declarations inside [@keyframes]
//
// Both typed VarRef values (constructed by the Builder) and var()
// references that came in through [Parse] as opaque [Literal] text are
// detected.
func (s *Sheet) UndefinedVars() []string {
	if s == nil {
		return nil
	}
	defined := map[string]bool{}
	for _, name := range s.RegisteredVars() {
		defined[name] = true
	}
	seen := map[string]bool{}
	s.collectUndefined(defined, seen)
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func (s *Sheet) collectUndefined(defined, seen map[string]bool) {
	if s == nil {
		return
	}
	for _, r := range s.rules {
		if r.Selector.String() == ":root" {
			continue
		}
		for _, d := range r.Decls {
			collectVarRefs(d.Value, defined, seen)
		}
	}
	for _, ar := range s.atRules {
		switch ar.kind {
		case atMedia, atLayer, atSupports:
			ar.inner.collectUndefined(defined, seen)
		case atKeyframes:
			if ar.keyframe == nil {
				continue
			}
			for _, stop := range ar.keyframe.stops {
				for _, d := range stop.decls {
					collectVarRefs(d.Value, defined, seen)
				}
			}
		}
	}
}

// varRefScanRE captures the name inside var(--name) or var(--name, fallback).
// The character class matches the one enforced by [validVarName] in types.go.
var varRefScanRE = regexp.MustCompile(`var\(\s*--([a-z][a-z0-9_-]*)`)

func collectVarRefs(v Value, defined, seen map[string]bool) {
	switch x := v.(type) {
	case varRefValue:
		if !defined[x.name] {
			seen[x.name] = true
		}
	case literalValue:
		// Parse stores declarations as Literal(rawText); scan that text for
		// embedded var(--name) references so values that arrived through
		// the parser are visible to UndefinedVars diagnostics. Strip CSS
		// string literals first so `content: "var(--copy)"` (where the
		// var() text is inside a quoted string) doesn't trigger a false
		// SE001. CSS comment markers are not stripped here because
		// styleengine/Sanitize already rejects them for untrusted input
		// and trusted callers do not embed them in values.
		scanText := stripCSSStrings(string(x))
		for _, m := range varRefScanRE.FindAllStringSubmatch(scanText, -1) {
			name := m[1]
			if !defined[name] {
				seen[name] = true
			}
		}
	}
}

// stripCSSStrings removes the contents of double-quoted and single-quoted
// CSS string literals, replacing them with empty quotes. Backslash escapes
// inside strings are honored so `"a\"b"` is correctly recognized as a
// single string. The result preserves the structure of the surrounding
// CSS while making string contents invisible to subsequent regex scans.
func stripCSSStrings(s string) string {
	if !strings.ContainsAny(s, `"'`) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		c := s[i]
		if c != '"' && c != '\'' {
			b.WriteByte(c)
			i++
			continue
		}
		// Found a string-literal opener; copy opener and skip to matching close.
		quote := c
		b.WriteByte(quote)
		i++
		for i < len(s) {
			cc := s[i]
			if cc == '\\' && i+1 < len(s) {
				i += 2
				continue
			}
			if cc == quote {
				b.WriteByte(quote)
				i++
				break
			}
			i++
		}
	}
	return b.String()
}
