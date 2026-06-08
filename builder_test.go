// Implements: OSS-BB-001 (styleengine building block; tests).
// Per: ADR-0029 (every file declares its purpose); Library Excellence Standard (see README provenance).
// Discipline: C-14.

package styleengine

import (
	"strings"
	"testing"
)

func TestBuilder_RuleChain(t *testing.T) {
	sheet := New().
		Var("font-display", "Fraunces, Georgia, serif").
		Rule(":root").
		Decl("color-scheme", Literal("light dark")).
		Done().
		Rule(".pke-editorial h1").
		Decl("font-family", VarRef("font-display", "")).
		Done().
		Build()
	got := sheet.RenderPretty()
	if !strings.Contains(got, "--font-display:") {
		t.Fatalf("expected var, got:\n%s", got)
	}
	if !strings.Contains(got, "color-scheme: light dark") {
		t.Fatalf("expected color-scheme, got:\n%s", got)
	}
	if !strings.Contains(got, ".pke-editorial h1") {
		t.Fatalf("expected editorial rule, got:\n%s", got)
	}
}

func TestBuilder_BuildIsIdempotent(t *testing.T) {
	b := New().Var("a", "1")
	s1 := b.Build()
	s2 := b.Build()
	if s1 != s2 {
		t.Fatalf("Build should return the same Sheet pointer twice")
	}
}
