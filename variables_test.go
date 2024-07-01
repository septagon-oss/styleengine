// Implements: OSS-BB (library building block; see README).
// Per: ADR-0029 (every file declares its purpose).
// Discipline: C-14.

package styleengine

import (
	"sort"
	"strings"
	"testing"
)

func TestSheet_Var_EmitsInRoot(t *testing.T) {
	s := NewSheet()
	s.Var("font-display", "Fraunces, Georgia, serif")
	got := s.RenderPretty()
	if !strings.Contains(got, ":root") {
		t.Fatalf("expected :root block, got:\n%s", got)
	}
	if !strings.Contains(got, "--font-display: Fraunces, Georgia, serif") {
		t.Fatalf("expected var emission, got:\n%s", got)
	}
}

func TestSheet_Var_RejectsBadName(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic on invalid var name")
		}
	}()
	NewSheet().Var("Bad", "x")
}

func TestSheet_Var_SortedAlphabeticallyInRoot(t *testing.T) {
	s := NewSheet()
	s.Var("zeta", "1")
	s.Var("alpha", "2")
	s.Var("mu", "3")
	got := s.RenderPretty()
	aIdx := strings.Index(got, "--alpha")
	mIdx := strings.Index(got, "--mu")
	zIdx := strings.Index(got, "--zeta")
	if !(aIdx < mIdx && mIdx < zIdx) {
		t.Fatalf("expected alphabetical, got:\n%s", got)
	}
}

func TestSheet_UndefinedVars(t *testing.T) {
	s := NewSheet()
	s.Var("defined", "x")
	s.AddRule(Rule{
		Selector: MustSelector(".a"),
		Decls: []Declaration{
			Decl("font", VarRef("defined", "")),
			Decl("color", VarRef("missing", "")),
		},
	})
	undef := s.UndefinedVars()
	sort.Strings(undef)
	if len(undef) != 1 || undef[0] != "missing" {
		t.Fatalf("expected [missing], got %v", undef)
	}
}

// TestSheet_UndefinedVars_RecursesIntoAtRules guards a regression where
// var() references nested inside @media/@layer/@supports/@keyframes were
// invisible to UndefinedVars because the scan only walked top-level rules.
func TestSheet_UndefinedVars_RecursesIntoAtRules(t *testing.T) {
	s := NewSheet()
	s.Var("defined", "x")
	s.Media("(min-width: 800px)", func(inner *Sheet) {
		inner.AddRule(Rule{
			Selector: MustSelector(".desktop"),
			Decls:    []Declaration{Decl("color", VarRef("missing-media", ""))},
		})
	})
	s.Keyframes("shift", func(k *KeyframesBuilder) {
		k.At("0%", Decl("transform", VarRef("missing-keyframe", "")))
	})
	s.Supports("(display: grid)", func(inner *Sheet) {
		inner.AddRule(Rule{
			Selector: MustSelector(".grid"),
			Decls:    []Declaration{Decl("font", VarRef("missing-supports", ""))},
		})
	})
	undef := s.UndefinedVars()
	want := map[string]bool{
		"missing-media":    true,
		"missing-keyframe": true,
		"missing-supports": true,
	}
	if len(undef) != len(want) {
		t.Fatalf("UndefinedVars = %v, want keys %v", undef, want)
	}
	for _, name := range undef {
		if !want[name] {
			t.Errorf("unexpected name %q", name)
		}
	}
}

// TestSheet_UndefinedVars_IgnoresVarRefsInsideStrings guards a regression
// where the regex scanner treated text inside CSS string literals as actual
// var() references — e.g., `content: "var(--copy)"` would falsely emit
// SE001 for var(--copy) even though it is just literal text.
func TestSheet_UndefinedVars_IgnoresVarRefsInsideStrings(t *testing.T) {
	s, err := Parse(`.x { content: "var(--copy)"; quotes: 'and var(--also)'; color: var(--real); }`)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	s.Var("real", "red")
	undef := s.UndefinedVars()
	for _, name := range undef {
		if name == "copy" || name == "also" {
			t.Errorf("UndefinedVars reported %q which was inside a CSS string literal: %v", name, undef)
		}
	}
	s2, _ := Parse(`.x { color: var(--missing); }`)
	if got := s2.UndefinedVars(); len(got) != 1 || got[0] != "missing" {
		t.Fatalf("expected [missing], got %v", got)
	}
}

// TestStripCSSStrings_ExactBehavior pins the helper's contract so changes
// to the scanner stay safe.
func TestStripCSSStrings_ExactBehavior(t *testing.T) {
	cases := []struct{ in, want string }{
		{`color: red`, `color: red`},
		{`content: "var(--x)"`, `content: ""`},
		{`content: 'var(--x)'`, `content: ''`},
		{`content: "a\"b"`, `content: ""`},
		{`a: "x"; b: "y"`, `a: ""; b: ""`},
	}
	for _, tc := range cases {
		if got := stripCSSStrings(tc.in); got != tc.want {
			t.Errorf("stripCSSStrings(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestSheet_UndefinedVars_DetectsParsedLiteralRefs guards a regression where
// var() references that arrived through Parse — stored as Literal(rawText)
// — were invisible to UndefinedVars diagnostics.
func TestSheet_UndefinedVars_DetectsParsedLiteralRefs(t *testing.T) {
	s, err := Parse(".x { color: var(--missing); font: var(--also-missing, 16px); padding: var(--registered); }")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	s.Var("registered", "1rem")
	undef := s.UndefinedVars()
	want := map[string]bool{"missing": true, "also-missing": true}
	if len(undef) != len(want) {
		t.Fatalf("UndefinedVars = %v, want keys %v", undef, want)
	}
	for _, name := range undef {
		if !want[name] {
			t.Errorf("unexpected undefined var %q", name)
		}
	}
}
