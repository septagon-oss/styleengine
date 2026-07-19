// atrules_test.go validates supported CSS at-rule construction and rendering.
// Validates: REQ-011.
// Per: ADR-0029 (every file declares its purpose).
// Discipline: C-14.

package styleengine

import (
	"strings"
	"testing"
)

func TestSheet_Media(t *testing.T) {
	s := NewSheet()
	s.Media("(prefers-reduced-motion: reduce)", func(m *Sheet) {
		m.AddRule(Rule{
			Selector: MustSelector(":root"),
			Decls:    []Declaration{Decl("--motion-duration-marquee", Literal("0ms"))},
		})
	})
	got := s.RenderPretty()
	if !strings.Contains(got, "@media (prefers-reduced-motion: reduce)") {
		t.Fatalf("expected @media, got:\n%s", got)
	}
	if !strings.Contains(got, "--motion-duration-marquee: 0ms") {
		t.Fatalf("expected nested decl, got:\n%s", got)
	}
}

func TestSheet_Keyframes(t *testing.T) {
	s := NewSheet()
	s.Keyframes("marquee-shift", func(k *KeyframesBuilder) {
		k.At("0%", Decl("transform", Literal("translateX(0)")))
		k.At("100%", Decl("transform", Literal("translateX(-50%)")))
	})
	got := s.RenderPretty()
	if !strings.Contains(got, "@keyframes marquee-shift") {
		t.Fatalf("expected @keyframes, got:\n%s", got)
	}
	if !strings.Contains(got, "0% {") || !strings.Contains(got, "100% {") {
		t.Fatalf("expected stops, got:\n%s", got)
	}
}

func TestSheet_FontFace(t *testing.T) {
	s := NewSheet()
	s.AddFontFace(FontFaceDecl{
		Family:  "Fraunces",
		Src:     `url("/fonts/fraunces.woff2") format("woff2")`,
		Weight:  "100 900",
		Style:   "normal",
		Display: "swap",
	})
	got := s.RenderPretty()
	if !strings.Contains(got, "@font-face") {
		t.Fatalf("expected @font-face, got:\n%s", got)
	}
	if !strings.Contains(got, `font-family: "Fraunces"`) {
		t.Fatalf("expected font-family, got:\n%s", got)
	}
	if !strings.Contains(got, "font-display: swap") {
		t.Fatalf("expected font-display, got:\n%s", got)
	}
}

func TestSheet_Layer(t *testing.T) {
	s := NewSheet()
	s.Layer("tokens", func(l *Sheet) {
		l.AddRule(Rule{Selector: MustSelector(":root"), Decls: []Declaration{Decl("--x", Literal("1"))}})
	})
	got := s.RenderPretty()
	if !strings.Contains(got, "@layer tokens") {
		t.Fatalf("expected @layer, got:\n%s", got)
	}
}

func TestSheet_Supports(t *testing.T) {
	s := NewSheet()
	s.Supports("(display: grid)", func(b *Sheet) {
		b.AddRule(Rule{Selector: MustSelector(".grid"), Decls: []Declaration{Decl("display", Literal("grid"))}})
	})
	got := s.RenderPretty()
	if !strings.Contains(got, "@supports (display: grid)") {
		t.Fatalf("expected @supports, got:\n%s", got)
	}
}
