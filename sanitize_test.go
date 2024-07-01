// Implements: OSS-BB-001 (styleengine building block; tests).
// Per: ADR-0029 (every file declares its purpose); Library Excellence Standard (see README provenance) (Phase 0).
// Discipline: C-14.

package styleengine

import (
	"strings"
	"testing"
)

func TestSanitize_Allows_SimpleDeclarations(t *testing.T) {
	in := "color: red;\nfont-size: 16px;"
	out, err := Sanitize(in)
	if err != nil {
		t.Fatalf("Sanitize error: %v", err)
	}
	if !strings.Contains(out, "color: red") || !strings.Contains(out, "font-size: 16px") {
		t.Fatalf("expected both decls preserved, got:\n%s", out)
	}
}

func TestSanitize_Rejects_DangerousPatterns(t *testing.T) {
	cases := []string{
		"color: red; } body { display: none;",  // attempted brace breakout
		"background: url(javascript:alert(1))", // js url
		"@import 'evil.css'",                   // import
		"behavior: url(#hack)",                 // IE legacy
		"binding: url(#x)",                     // Mozilla legacy
		"color: expression(alert(1))",          // IE expressions
	}
	for _, in := range cases {
		if _, err := Sanitize(in); err == nil {
			t.Fatalf("expected rejection for %q", in)
		}
	}
}

func TestSanitize_PassesThroughVar(t *testing.T) {
	in := "color: var(--brand-primary);"
	out, err := Sanitize(in)
	if err != nil {
		t.Fatalf("Sanitize error: %v", err)
	}
	if !strings.Contains(out, "var(--brand-primary)") {
		t.Fatalf("expected var() preserved, got:\n%s", out)
	}
}

// TestSanitize_RejectsURLVariantsWithWhitespace guards a regression where
// `url(javascript:` was only matched as an exact substring. Whitespace and
// any url(...) form must be rejected for untrusted input.
func TestSanitize_RejectsURLVariantsWithWhitespace(t *testing.T) {
	cases := []string{
		"background: url ( javascript:alert(1))",
		"background: url(  javascript:alert(1))",
		"background: URL(javascript:alert(1))",
		"background: url('https://evil.example/img.png')",
		"background: url(/local/asset.png)",
		"color: var(--brand, javascript:alert(1))",
	}
	for _, in := range cases {
		if _, err := Sanitize(in); err == nil {
			t.Errorf("expected rejection for %q", in)
		}
	}
}

// TestSanitize_RejectsMarkupBreakout guards the </style> bypass where a
// untrusted value embedded in a <style> block would otherwise terminate the
// raw-text element and inject arbitrary HTML (with onerror handlers etc.).
func TestSanitize_RejectsMarkupBreakout(t *testing.T) {
	cases := []string{
		"color: red;</style><img src=x onerror=alert(1)>",
		"color: red;</STYLE><img src=x onerror=alert(1)>",
		"color: red; <style>alert(1)</style>",
		"color: red; </script>",
		"color: red; <script>",
	}
	for _, in := range cases {
		if _, err := Sanitize(in); err == nil {
			t.Errorf("expected rejection for %q", in)
		}
	}
}

// TestSanitize_RejectsImageFunctions guards a regression where CSS image
// functions (image-set, cross-fade, paint, element, image) could bypass
// the url() guard by accepting URL or string arguments directly.
func TestSanitize_RejectsImageFunctions(t *testing.T) {
	cases := []string{
		`background-image: image-set("https://attacker/x.png" 1x)`,
		`background: image("path.png")`,
		`background: cross-fade(image("a.png"), 50%)`,
		`background: element(#id)`,
		`background-image: paint(workerName)`,
		`background: IMAGE-SET("x.png" 1x)`, // case-insensitive
	}
	for _, in := range cases {
		if _, err := Sanitize(in); err == nil {
			t.Errorf("expected rejection for %q", in)
		}
	}
}

// TestSanitize_RejectsCSSEscapeSequences guards a regression where untrusted
// could encode dangerous tokens with CSS escapes — `\75 rl(...)` decodes to
// `url(...)` in the browser but would not match the literal substring or
// regex checks.
func TestSanitize_RejectsCSSEscapeSequences(t *testing.T) {
	cases := []string{
		`background: \75 rl(/x.png)`,                 // \75 == 'u'
		`background: \75 rl(\6a avascript:alert(1))`, // doubled escape
		`color: \72 ed`,                              // 'red' via escape — still rejected
		`color: red\;` + " color: blue",              // trailing backslash
	}
	for _, in := range cases {
		if _, err := Sanitize(in); err == nil {
			t.Errorf("expected rejection for %q", in)
		}
	}
}
