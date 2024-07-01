// Implements: OSS-BB (library building block; see README).
// Per: ADR-0029 (every file declares its purpose).
// Discipline: C-14.

package styleengine

import "testing"

func TestLiteralValue_CSS(t *testing.T) {
	v := Literal("16px")
	if got := v.CSS(); got != "16px" {
		t.Fatalf("Literal.CSS = %q, want %q", got, "16px")
	}
}

func TestVarRefValue_CSS(t *testing.T) {
	cases := []struct {
		name string
		ref  Value
		want string
	}{
		{"no fallback", VarRef("font-display", ""), "var(--font-display)"},
		{"with fallback", VarRef("font-display", "Georgia, serif"), "var(--font-display, Georgia, serif)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.ref.CSS(); got != tc.want {
				t.Fatalf("VarRef.CSS = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestVarRefValue_RejectsBadName(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic on invalid var name")
		}
	}()
	_ = VarRef("Bad-Name", "")
}

// TestVarRefValue_RejectsBadFallback guards a regression where the fallback
// string was concatenated verbatim and could break out of the var()
// expression (and the surrounding rule) when given attacker-controlled or
// sloppy input.
func TestVarRefValue_RejectsBadFallback(t *testing.T) {
	cases := []string{
		"red); } body{display:none}/*",
		"red)",
		"red; color: blue",
		"red /* comment */",
		"red{}",
	}
	for _, bad := range cases {
		bad := bad
		t.Run(bad, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("expected panic for fallback %q", bad)
				}
			}()
			_ = VarRef("brand", bad)
		})
	}
}

func TestDeclaration_CSS(t *testing.T) {
	d := Decl("color", Literal("red"))
	if got := d.CSS(); got != "color: red" {
		t.Fatalf("Decl.CSS = %q, want %q", got, "color: red")
	}
	d2 := Decl("color", Literal("red")).Important()
	if got := d2.CSS(); got != "color: red !important" {
		t.Fatalf("Decl.Important.CSS = %q, want %q", got, "color: red !important")
	}
}

func TestSelector_Normalize(t *testing.T) {
	cases := []struct{ in, want string }{
		{".foo  .bar", ".foo .bar"},
		{".foo,.bar", ".foo, .bar"},
		{"  :root  ", ":root"},
	}
	for _, tc := range cases {
		if got := MustSelector(tc.in).String(); got != tc.want {
			t.Fatalf("Selector(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSelector_RejectsEmpty(t *testing.T) {
	if _, err := NewSelector(""); err == nil {
		t.Fatalf("expected error for empty selector")
	}
}

func TestRule_CSS_Pretty(t *testing.T) {
	r := Rule{
		Selector: MustSelector(".foo"),
		Decls: []Declaration{
			Decl("color", Literal("red")),
			Decl("font-size", VarRef("text-md", "")),
		},
	}
	want := ".foo {\n  color: red;\n  font-size: var(--text-md);\n}"
	if got := r.CSSPretty(); got != want {
		t.Fatalf("Rule.CSSPretty mismatch:\n got:  %q\n want: %q", got, want)
	}
}
