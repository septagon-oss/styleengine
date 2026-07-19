// render_test.go validates pretty, minified, and size-bounded CSS rendering.
// Validates: REQ-011.
// Per: ADR-0029 (every file declares its purpose).
// Discipline: C-14.

package styleengine

import (
	"strings"
	"testing"
)

func TestSheet_Render_Pretty(t *testing.T) {
	s := New().Var("a", "1").Rule(".b").Decl("color", Literal("red")).Done().Build()
	out, err := s.Render(RenderOptions{Pretty: true})
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(out, "  --a: 1;") {
		t.Fatalf("expected indented var, got:\n%s", out)
	}
}

func TestSheet_Render_Minified(t *testing.T) {
	s := New().Var("a", "1").Rule(".b").Decl("color", Literal("red")).Done().Build()
	out, err := s.Render(RenderOptions{Minify: true})
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if strings.Contains(out, "\n  ") {
		t.Fatalf("expected no indentation, got:\n%s", out)
	}
	if !strings.Contains(out, "--a:1") && !strings.Contains(out, "--a: 1") {
		t.Fatalf("expected var present, got:\n%s", out)
	}
}

func TestSheet_Render_MaxBytes(t *testing.T) {
	s := New().Var("a", strings.Repeat("x", 2000)).Build()
	_, err := s.Render(RenderOptions{Pretty: true, MaxBytes: 100})
	if err == nil {
		t.Fatalf("expected MaxBytes error")
	}
}

// TestSheet_Render_RejectsPrettyAndMinify guards the API tightening that
// makes the two output modes mutually exclusive instead of silently
// preferring Minify.
func TestSheet_Render_RejectsPrettyAndMinify(t *testing.T) {
	s := New().Var("a", "1").Build()
	_, err := s.Render(RenderOptions{Pretty: true, Minify: true})
	if err == nil {
		t.Fatalf("expected error when both Pretty and Minify are set")
	}
}

// TestSheet_Render_MaxBytesDisabledByNegative covers the opt-out for
// one-shot batch jobs that know they exceed the runtime-safe default.
func TestSheet_Render_MaxBytesDisabledByNegative(t *testing.T) {
	s := New().Var("a", strings.Repeat("x", 2_000_000)).Build()
	out, err := s.Render(RenderOptions{Pretty: true, MaxBytes: -1})
	if err != nil {
		t.Fatalf("MaxBytes=-1 should disable the check, got: %v", err)
	}
	if int64(len(out)) <= DefaultMaxBytes {
		t.Fatalf("expected output > %d bytes, got %d", DefaultMaxBytes, len(out))
	}
}
