// benchmarks_test.go measures representative sheet rendering costs.
// Validates: REQ-011.
// Per: ADR-0029 (every file declares its purpose).
// Discipline: C-14.

package styleengine

import (
	"fmt"
	"testing"
)

func bigSheet(n int) *Sheet {
	b := New()
	for i := 0; i < n; i++ {
		b.Var(fmt.Sprintf("v%d", i), fmt.Sprintf("%dpx", i))
	}
	for i := 0; i < n; i++ {
		sel := fmt.Sprintf(".class-%d", i)
		b.Rule(sel).
			Decl("color", VarRef(fmt.Sprintf("v%d", i), "")).
			Decl("padding", Literal("1rem")).
			Done()
	}
	return b.Build()
}

func BenchmarkRender_Pretty(b *testing.B) {
	s := bigSheet(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.Render(RenderOptions{Pretty: true, MaxBytes: 1 << 22}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRender_Minified(b *testing.B) {
	s := bigSheet(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.Render(RenderOptions{Minify: true, MaxBytes: 1 << 22}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse(b *testing.B) {
	s := bigSheet(50)
	css, _ := s.Render(RenderOptions{Pretty: true, MaxBytes: 1 << 22})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Parse(css); err != nil {
			b.Fatal(err)
		}
	}
}
