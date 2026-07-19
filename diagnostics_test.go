// diagnostics_test.go validates deterministic sheet diagnostics.
// Validates: REQ-011.
// Per: ADR-0029 (every file declares its purpose).
// Discipline: C-14.

package styleengine

import "testing"

func TestDiagnostics_EmptySheet(t *testing.T) {
	if got := NewSheet().Diagnostics(); len(got) != 0 {
		t.Fatalf("expected no diagnostics on empty sheet, got %v", got)
	}
}

func TestDiagnostics_ReportsUndefinedVars(t *testing.T) {
	s := NewSheet()
	s.Var("registered", "1rem")
	s.AddRule(Rule{
		Selector: MustSelector(".x"),
		Decls: []Declaration{
			Decl("font", VarRef("missing-one", "")),
			Decl("color", VarRef("missing-two", "")),
			Decl("padding", VarRef("registered", "")),
		},
	})
	diags := s.Diagnostics()
	if len(diags) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d: %v", len(diags), diags)
	}
	for _, d := range diags {
		if d.Code != DiagUndefinedVar {
			t.Errorf("expected code %s, got %s", DiagUndefinedVar, d.Code)
		}
		if d.Kind != DiagnosticError {
			t.Errorf("expected error kind, got %s", d.Kind)
		}
	}
	if diags[0].Subject != "missing-one" || diags[1].Subject != "missing-two" {
		t.Errorf("expected sorted subjects [missing-one missing-two], got %s %s", diags[0].Subject, diags[1].Subject)
	}
}

func TestDiagnosticKind_String(t *testing.T) {
	cases := []struct {
		k    DiagnosticKind
		want string
	}{
		{DiagnosticError, "error"},
		{DiagnosticWarning, "warning"},
		{DiagnosticHint, "hint"},
		{DiagnosticKind(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.k.String(); got != tc.want {
			t.Errorf("DiagnosticKind(%d).String() = %q, want %q", tc.k, got, tc.want)
		}
	}
}
