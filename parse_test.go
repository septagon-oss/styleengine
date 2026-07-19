// parse_test.go validates CSS parsing and supported round trips.
// Validates: REQ-011.
// Per: ADR-0029 (every file declares its purpose).
// Discipline: C-14.

package styleengine

import (
	"os"
	"strings"
	"testing"
)

func TestParse_SimpleRule(t *testing.T) {
	in := ".foo { color: red; font-size: 16px; }"
	s, err := Parse(in)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	got := s.RenderPretty()
	if !strings.Contains(got, ".foo") {
		t.Fatalf("expected .foo, got:\n%s", got)
	}
	if !strings.Contains(got, "color: red") {
		t.Fatalf("expected color: red, got:\n%s", got)
	}
}

func TestParse_VarReference(t *testing.T) {
	s, err := Parse(".foo { color: var(--brand); }")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if !strings.Contains(s.RenderPretty(), "var(--brand)") {
		t.Fatalf("expected var() preserved")
	}
}

func TestParse_AtRule_DoesNotLeakInnerRules(t *testing.T) {
	in := `.top { color: red; }
@media (min-width: 800px) {
  .desktop { color: blue; }
  :root { --x: 1; }
}
.bottom { color: green; }`
	s, err := Parse(in)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	got := s.RenderPretty()
	if !strings.Contains(got, ".top") {
		t.Fatalf("expected .top preserved, got:\n%s", got)
	}
	if !strings.Contains(got, ".bottom") {
		t.Fatalf("expected .bottom preserved, got:\n%s", got)
	}
	if strings.Contains(got, ".desktop") {
		t.Fatalf("inner @media rule leaked to top level:\n%s", got)
	}
	if strings.Contains(got, "--x") {
		t.Fatalf("inner @media custom property leaked to top level:\n%s", got)
	}
}

func TestParse_Roundtrip(t *testing.T) {
	data, err := os.ReadFile("testdata/parse_roundtrip_simple.css")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}
	s1, err := Parse(string(data))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	rendered := s1.RenderPretty()
	s2, err := Parse(rendered)
	if err != nil {
		t.Fatalf("Re-parse error: %v\n--- rendered ---\n%s", err, rendered)
	}
	if s1.RenderPretty() != s2.RenderPretty() {
		t.Fatalf("Roundtrip not idempotent:\n--- s1 ---\n%s\n--- s2 ---\n%s", s1.RenderPretty(), s2.RenderPretty())
	}
}
