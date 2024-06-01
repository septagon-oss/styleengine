// Implements: OSS-BB-001 (styleengine building block).
// Per: ADR-0029 (every file declares its purpose); Library Excellence Standard (see README provenance) (Phase 0 exemplar extraction).
// Discipline: C-14.

package styleengine

import (
	"strings"
)

// Sheet is the root IR. It is an ordered, deduped, mergeable collection of
// Rules and AtRules.
//
// Sheet is NOT safe for concurrent mutation. The design intent is
// single-writer construction: a Sheet is constructed in one goroutine,
// possibly merged with other partial Sheets, then rendered. If a future
// consumer needs concurrent mutation, wrap calls in a mutex at the caller —
// Sheet itself stays lock-free to keep render hot-paths fast.
type Sheet struct {
	rules     []*Rule        // ordered by insertion (per unique selector)
	ruleIdx   map[string]int // selector -> index in rules
	atRules   []AtRule
	fontFaces []FontFace
}

// NewSheet returns an empty Sheet.
func NewSheet() *Sheet {
	return &Sheet{
		ruleIdx: map[string]int{},
	}
}

// AddRule adds (or merges) a Rule. If a rule with the same normalized selector
// already exists, the new declarations merge into it; duplicates within a
// (selector, property, !important) triple are resolved last-write-wins.
func (s *Sheet) AddRule(r Rule) {
	key := r.Selector.String()
	if idx, ok := s.ruleIdx[key]; ok {
		s.rules[idx] = mergeRule(s.rules[idx], &r)
		return
	}
	cp := r
	s.rules = append(s.rules, &cp)
	s.ruleIdx[key] = len(s.rules) - 1
}

func mergeRule(into, from *Rule) *Rule {
	idx := map[string]int{}
	for i, d := range into.Decls {
		idx[d.Property+importantKey(d.important)] = i
	}
	for _, d := range from.Decls {
		k := d.Property + importantKey(d.important)
		if i, ok := idx[k]; ok {
			into.Decls[i] = d
			continue
		}
		idx[k] = len(into.Decls)
		into.Decls = append(into.Decls, d)
	}
	return into
}

func importantKey(b bool) string {
	if b {
		return "!"
	}
	return ""
}

// Merge appends every rule, at-rule, and font-face from other into s,
// preserving Sheet's dedup invariants. The other Sheet is unchanged.
//
// Merge deep-copies AtRule and keyframesBody contents so that subsequent
// mutation of the source sheet (or its captured inner builders) does not
// leak into the merged result. Composing several partial Sheets (e.g.,
// token-derived + layered overrides) is safe even if the
// partial Sheets are reused after merging.
func (s *Sheet) Merge(other *Sheet) *Sheet {
	if s == nil || other == nil {
		return s
	}
	for _, r := range other.rules {
		s.AddRule(Rule{Selector: r.Selector, Decls: append([]Declaration(nil), r.Decls...)})
	}
	for _, ar := range other.atRules {
		s.atRules = append(s.atRules, cloneAtRule(ar))
	}
	s.fontFaces = append(s.fontFaces, other.fontFaces...)
	return s
}

// cloneSheet returns a deep copy of src suitable for embedding in a merged
// AtRule. The returned sheet has its own rules/atRules/fontFaces slices and
// can be mutated independently of src.
func cloneSheet(src *Sheet) *Sheet {
	if src == nil {
		return nil
	}
	dst := NewSheet()
	for _, r := range src.rules {
		dst.AddRule(Rule{Selector: r.Selector, Decls: append([]Declaration(nil), r.Decls...)})
	}
	for _, ar := range src.atRules {
		dst.atRules = append(dst.atRules, cloneAtRule(ar))
	}
	dst.fontFaces = append(dst.fontFaces, src.fontFaces...)
	return dst
}

// cloneAtRule deep-copies an AtRule, including its inner Sheet (for
// @media/@layer/@supports) and its keyframesBody (for @keyframes). FontFace
// is a value type and is safe to copy directly via the value-typed return.
func cloneAtRule(ar AtRule) AtRule {
	out := AtRule{kind: ar.kind, prelude: ar.prelude}
	if ar.inner != nil {
		out.inner = cloneSheet(ar.inner)
	}
	if ar.keyframe != nil {
		body := &keyframesBody{name: ar.keyframe.name}
		body.stops = make([]keyframeStop, len(ar.keyframe.stops))
		for i, stop := range ar.keyframe.stops {
			body.stops[i] = keyframeStop{
				offset: stop.offset,
				decls:  append([]Declaration(nil), stop.decls...),
			}
		}
		out.keyframe = body
	}
	return out
}

// RenderPretty emits the Sheet as indented CSS. AtRules and FontFaces
// render after top-level rules in insertion order.
func (s *Sheet) RenderPretty() string {
	if s == nil {
		return ""
	}
	var b strings.Builder
	for i, r := range s.rules {
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(r.CSSPretty())
	}
	for _, ff := range s.fontFaces {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(ff.CSSPretty())
	}
	for _, ar := range s.atRules {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(ar.CSSPretty())
	}
	return b.String()
}
