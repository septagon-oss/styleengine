// Implements: OSS-BB (library building block; see README).
// Per: ADR-0029 (every file declares its purpose).
// Discipline: C-14.

package styleengine

import (
	"strings"
	"testing"
)

func TestSheet_AddRule_PreservesInsertionOrder(t *testing.T) {
	s := NewSheet()
	s.AddRule(Rule{Selector: MustSelector(".a"), Decls: []Declaration{Decl("color", Literal("red"))}})
	s.AddRule(Rule{Selector: MustSelector(".b"), Decls: []Declaration{Decl("color", Literal("blue"))}})
	got := s.RenderPretty()
	aIdx := strings.Index(got, ".a")
	bIdx := strings.Index(got, ".b")
	if aIdx == -1 || bIdx == -1 || aIdx > bIdx {
		t.Fatalf("expected .a before .b, got:\n%s", got)
	}
}

func TestSheet_AddRule_DedupesByLastWrite(t *testing.T) {
	s := NewSheet()
	s.AddRule(Rule{Selector: MustSelector(".a"), Decls: []Declaration{Decl("color", Literal("red"))}})
	s.AddRule(Rule{Selector: MustSelector(".a"), Decls: []Declaration{Decl("color", Literal("blue"))}})
	got := s.RenderPretty()
	if strings.Count(got, ".a") != 1 {
		t.Fatalf("expected one .a rule, got:\n%s", got)
	}
	if !strings.Contains(got, "color: blue") {
		t.Fatalf("expected last write to win (color: blue), got:\n%s", got)
	}
	if strings.Contains(got, "color: red") {
		t.Fatalf("expected color: red to be deduped out, got:\n%s", got)
	}
}

func TestSheet_AddRule_MergesDifferentPropertiesUnderSameSelector(t *testing.T) {
	s := NewSheet()
	s.AddRule(Rule{Selector: MustSelector(".a"), Decls: []Declaration{Decl("color", Literal("red"))}})
	s.AddRule(Rule{Selector: MustSelector(".a"), Decls: []Declaration{Decl("font-size", Literal("16px"))}})
	got := s.RenderPretty()
	if !strings.Contains(got, "color: red") {
		t.Fatalf("expected color: red preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "font-size: 16px") {
		t.Fatalf("expected font-size: 16px merged in, got:\n%s", got)
	}
}

func TestSheet_Merge_PreservesOrderAndDedup(t *testing.T) {
	a := NewSheet()
	a.AddRule(Rule{Selector: MustSelector(".x"), Decls: []Declaration{Decl("color", Literal("red"))}})
	a.AddRule(Rule{Selector: MustSelector(".y"), Decls: []Declaration{Decl("color", Literal("blue"))}})

	b := NewSheet()
	b.AddRule(Rule{Selector: MustSelector(".x"), Decls: []Declaration{Decl("font-size", Literal("16px"))}})
	b.AddRule(Rule{Selector: MustSelector(".z"), Decls: []Declaration{Decl("color", Literal("green"))}})

	a.Merge(b)
	got := a.RenderPretty()

	if strings.Count(got, ".x") != 1 {
		t.Fatalf("expected .x deduped once, got:\n%s", got)
	}
	for _, want := range []string{"color: red", "font-size: 16px", ".y", ".z"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in merged output, got:\n%s", want, got)
		}
	}
}

// TestSheet_Merge_DeepCopiesAtRuleInnerSheet guards a regression where
// Merge shared the inner *Sheet pointer of at-rules with the source sheet;
// post-merge mutation of the source leaked into the merged Sheet.
func TestSheet_Merge_DeepCopiesAtRuleInnerSheet(t *testing.T) {
	src := NewSheet()
	src.Media("(min-width: 800px)", func(inner *Sheet) {
		inner.AddRule(Rule{
			Selector: MustSelector(".desktop"),
			Decls:    []Declaration{Decl("color", Literal("red"))},
		})
	})

	dst := NewSheet()
	dst.Merge(src)

	// Mutate the captured inner sheet via src's at-rule slice. A correct
	// implementation deep-copies, so dst must not observe this change.
	src.atRules[0].inner.AddRule(Rule{
		Selector: MustSelector(".desktop"),
		Decls:    []Declaration{Decl("color", Literal("blue"))},
	})
	mergedOut := dst.RenderPretty()
	if strings.Contains(mergedOut, "color: blue") {
		t.Fatalf("Merge shared inner sheet pointer; post-merge mutation leaked:\n%s", mergedOut)
	}
	if !strings.Contains(mergedOut, "color: red") {
		t.Fatalf("merged @media .desktop rule missing original color: red:\n%s", mergedOut)
	}
}

// TestSheet_Merge_DeepCopiesKeyframesBody guards the same invariant for
// the keyframesBody pointer carried inside @keyframes at-rules.
func TestSheet_Merge_DeepCopiesKeyframesBody(t *testing.T) {
	src := NewSheet()
	src.Keyframes("shift", func(k *KeyframesBuilder) {
		k.At("0%", Decl("transform", Literal("translateX(0)")))
		k.At("100%", Decl("transform", Literal("translateX(-50%)")))
	})

	dst := NewSheet()
	dst.Merge(src)

	src.atRules[0].keyframe.stops[0].decls[0] = Decl("transform", Literal("translateY(0)"))
	mergedOut := dst.RenderPretty()
	if strings.Contains(mergedOut, "translateY(0)") {
		t.Fatalf("Merge shared keyframesBody pointer; mutation leaked:\n%s", mergedOut)
	}
	if !strings.Contains(mergedOut, "translateX(0)") {
		t.Fatalf("merged @keyframes missing original translateX(0):\n%s", mergedOut)
	}
}

func TestSheet_Merge_NilSafe(t *testing.T) {
	a := NewSheet()
	a.Var("x", "1")
	a.Merge(nil)
	if got := a.RenderPretty(); !strings.Contains(got, "--x") {
		t.Fatalf("Merge(nil) should be a no-op, got:\n%s", got)
	}
}

func TestSheet_RenderPretty_EmptyIsEmpty(t *testing.T) {
	if got := NewSheet().RenderPretty(); got != "" {
		t.Fatalf("expected empty render, got %q", got)
	}
}
