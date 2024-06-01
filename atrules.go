// Implements: OSS-BB-001 (styleengine building block).
// Per: ADR-0029 (every file declares its purpose); Library Excellence Standard (see README provenance) (Phase 0 exemplar extraction).
// Discipline: C-14.

package styleengine

import (
	"sort"
	"strings"
)

// atRuleKind enumerates the @-rule families styleengine v1 supports.
type atRuleKind int

const (
	atMedia atRuleKind = iota
	atKeyframes
	atLayer
	atSupports
)

// AtRule is opaque; consumers construct via Sheet.Media/Keyframes/Layer/Supports.
type AtRule struct {
	kind     atRuleKind
	prelude  string
	inner    *Sheet
	keyframe *keyframesBody
}

// keyframesBody carries ordered animation stops.
type keyframesBody struct {
	name  string
	stops []keyframeStop
}

type keyframeStop struct {
	offset string
	decls  []Declaration
}

// Media nests a @media block. The closure receives a child Sheet that may
// contain rules, vars, and nested at-rules.
func (s *Sheet) Media(query string, fn func(*Sheet)) *Sheet {
	inner := NewSheet()
	fn(inner)
	s.atRules = append(s.atRules, AtRule{kind: atMedia, prelude: query, inner: inner})
	return s
}

// Layer nests a @layer block.
func (s *Sheet) Layer(name string, fn func(*Sheet)) *Sheet {
	inner := NewSheet()
	fn(inner)
	s.atRules = append(s.atRules, AtRule{kind: atLayer, prelude: name, inner: inner})
	return s
}

// Supports nests a @supports block.
func (s *Sheet) Supports(condition string, fn func(*Sheet)) *Sheet {
	inner := NewSheet()
	fn(inner)
	s.atRules = append(s.atRules, AtRule{kind: atSupports, prelude: condition, inner: inner})
	return s
}

// KeyframesBuilder accumulates animation stops.
type KeyframesBuilder struct {
	body *keyframesBody
}

// At adds a keyframe stop ("0%", "50%", "from", "to", etc.).
func (k *KeyframesBuilder) At(offset string, decls ...Declaration) *KeyframesBuilder {
	k.body.stops = append(k.body.stops, keyframeStop{offset: offset, decls: decls})
	return k
}

// Keyframes adds a @keyframes block.
func (s *Sheet) Keyframes(name string, fn func(*KeyframesBuilder)) *Sheet {
	body := &keyframesBody{name: name}
	fn(&KeyframesBuilder{body: body})
	s.atRules = append(s.atRules, AtRule{kind: atKeyframes, keyframe: body})
	return s
}

// FontFaceDecl is the structured form of an @font-face block.
type FontFaceDecl struct {
	Family       string
	Src          string
	Weight       string
	Style        string
	Display      string
	UnicodeRange string
}

// AddFontFace appends an @font-face block.
func (s *Sheet) AddFontFace(ff FontFaceDecl) *Sheet {
	s.fontFaces = append(s.fontFaces, FontFace{decl: ff})
	return s
}

// FontFace is the runtime IR for @font-face.
type FontFace struct {
	decl FontFaceDecl
}

// CSSPretty emits the @font-face block in pretty form.
func (f FontFace) CSSPretty() string {
	var b strings.Builder
	b.WriteString("@font-face {")
	if f.decl.Family != "" {
		b.WriteString("\n  font-family: ")
		b.WriteString(quoteFontFamily(f.decl.Family))
		b.WriteByte(';')
	}
	if f.decl.Src != "" {
		b.WriteString("\n  src: ")
		b.WriteString(f.decl.Src)
		b.WriteByte(';')
	}
	if f.decl.Weight != "" {
		b.WriteString("\n  font-weight: ")
		b.WriteString(f.decl.Weight)
		b.WriteByte(';')
	}
	if f.decl.Style != "" {
		b.WriteString("\n  font-style: ")
		b.WriteString(f.decl.Style)
		b.WriteByte(';')
	}
	if f.decl.Display != "" {
		b.WriteString("\n  font-display: ")
		b.WriteString(f.decl.Display)
		b.WriteByte(';')
	}
	if f.decl.UnicodeRange != "" {
		b.WriteString("\n  unicode-range: ")
		b.WriteString(f.decl.UnicodeRange)
		b.WriteByte(';')
	}
	b.WriteString("\n}")
	return b.String()
}

// quoteFontFamily wraps the family name in double quotes, escaping internal
// double quotes. Bare CSS keywords like serif/sans-serif/monospace should be
// passed un-quoted in the Src list, not as Family.
func quoteFontFamily(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}

// CSSPretty emits the at-rule in pretty form.
func (a AtRule) CSSPretty() string {
	switch a.kind {
	case atMedia:
		return wrapAtRulePretty("@media "+a.prelude, a.inner.RenderPretty())
	case atLayer:
		return wrapAtRulePretty("@layer "+a.prelude, a.inner.RenderPretty())
	case atSupports:
		return wrapAtRulePretty("@supports "+a.prelude, a.inner.RenderPretty())
	case atKeyframes:
		return keyframesPretty(a.keyframe)
	}
	return ""
}

func wrapAtRulePretty(header, body string) string {
	if body == "" {
		return header + " {\n}"
	}
	return header + " {\n" + indentBlock(body, "  ") + "\n}"
}

func indentBlock(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

func keyframesPretty(body *keyframesBody) string {
	var b strings.Builder
	b.WriteString("@keyframes ")
	b.WriteString(body.name)
	b.WriteString(" {")
	stops := append([]keyframeStop(nil), body.stops...)
	sort.SliceStable(stops, func(i, j int) bool {
		return keyframeOffsetOrder(stops[i].offset) < keyframeOffsetOrder(stops[j].offset)
	})
	for _, k := range stops {
		b.WriteString("\n  ")
		b.WriteString(k.offset)
		b.WriteString(" {")
		for _, d := range k.decls {
			b.WriteString("\n    ")
			b.WriteString(d.CSS())
			b.WriteByte(';')
		}
		b.WriteString("\n  }")
	}
	b.WriteString("\n}")
	return b.String()
}

func keyframeOffsetOrder(off string) int {
	switch off {
	case "from":
		return 0
	case "to":
		return 100
	}
	if strings.HasSuffix(off, "%") {
		n := 0
		for _, r := range off[:len(off)-1] {
			if r < '0' || r > '9' {
				return 9999
			}
			n = n*10 + int(r-'0')
		}
		return n
	}
	return 9999
}
