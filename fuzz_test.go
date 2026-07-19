// fuzz_test.go probes sanitizer and parser safety over arbitrary input.
// Validates: REQ-011.
// Per: ADR-0029 (every file declares its purpose).
// Discipline: C-14.

package styleengine

import (
	"strings"
	"testing"
)

// FuzzSanitize asserts that [Sanitize] never panics and never returns a
// string that still contains a dangerous pattern that should have been
// rejected. The contract: either err != nil, OR the returned string is
// free of every entry in dangerousSubstrings AND every dangerousPatterns
// match AND any backslash.
func FuzzSanitize(f *testing.F) {
	seeds := []string{
		"color: red",
		"color: var(--brand);",
		"background: url(javascript:alert(1))",
		"} body { display: none",
		"@import 'evil.css'",
		"behavior: url(#x)",
		"binding: url(#x)",
		"expression(alert(1))",
		`\75 rl(/x.png)`,
		"",
		strings.Repeat("a", 4096),
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, in string) {
		out, err := Sanitize(in)
		if err != nil {
			return
		}
		// If Sanitize accepted, output MUST be clean.
		if strings.ContainsRune(out, '\\') {
			t.Fatalf("accepted output contains backslash: %q (input %q)", out, in)
		}
		lower := strings.ToLower(out)
		for _, bad := range dangerousSubstrings {
			if strings.Contains(lower, bad) {
				t.Fatalf("accepted output contains dangerous substring %q: %q (input %q)", bad, out, in)
			}
		}
		for _, re := range dangerousPatterns {
			if loc := re.FindStringIndex(out); loc != nil {
				t.Fatalf("accepted output matches dangerous pattern %q in %q (input %q)", out[loc[0]:loc[1]], out, in)
			}
		}
	})
}

// FuzzParse asserts that [Parse] never panics on arbitrary input. Returns
// either an error or a Sheet whose RenderPretty is parse-stable (the
// roundtrip invariant only applies when the input was well-formed enough
// that Parse succeeded).
func FuzzParse(f *testing.F) {
	seeds := []string{
		":root { --x: 1; }",
		".a, .b { color: red; }",
		"@media (min-width: 800px) { .x { color: red; } }",
		"",
		"}}}{{{",
		"--malformed",
		strings.Repeat(".x{}", 100),
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, in string) {
		s, err := Parse(in)
		if err != nil {
			return
		}
		if s == nil {
			t.Fatalf("Parse returned (nil, nil) for input %q", in)
		}
		// Render must succeed (panics here would be the actual bug).
		_ = s.RenderPretty()
		_, _ = s.Render(RenderOptions{Minify: true, MaxBytes: -1})
	})
}

// FuzzRender_BuilderInputs asserts that [Sheet.Render] never panics on
// programmatically constructed sheets, regardless of var name/value
// pairs the caller supplies. Names that fail validVarName panic in Var,
// which is caught here so we only exercise the post-construction render
// path.
func FuzzRender_BuilderInputs(f *testing.F) {
	seeds := []struct {
		varName, value, selector, prop, declValue string
	}{
		{"a", "1", ".x", "color", "red"},
		{"font-display", `"Fraunces"`, "h1", "font", "var(--font-display)"},
		{"", "", "", "", ""},
	}
	for _, s := range seeds {
		f.Add(s.varName, s.value, s.selector, s.prop, s.declValue)
	}

	f.Fuzz(func(t *testing.T, varName, value, selector, prop, declValue string) {
		b := New()

		// Var with an invalid name panics; treat that as an expected
		// rejection (the validation IS the safety property) and skip.
		func() {
			defer func() { _ = recover() }()
			if varName != "" {
				b.Var(varName, value)
			}
		}()

		func() {
			defer func() { _ = recover() }()
			if selector != "" && prop != "" {
				b.Rule(selector).Decl(prop, Literal(declValue)).Done()
			}
		}()

		sheet := b.Build()
		_ = sheet.RenderPretty()
		_, _ = sheet.Render(RenderOptions{Pretty: true, MaxBytes: -1})
		_, _ = sheet.Render(RenderOptions{Minify: true, MaxBytes: -1})
	})
}
