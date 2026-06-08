// Implements: OSS-BB (see per-lib README for ID).
// Per: ADR-0029 (every file declares its purpose).
// Discipline: C-14.

package styleengine_test

import (
	"fmt"

	"github.com/septagon-oss/styleengine"
)

// Example shows basic construction of a sheet with variables and rules, then pretty render.
func Example() {
	sheet, err := styleengine.New().
		Var("color-primary", "#0a84ff").
		Rule(":root").
		Decl("color", styleengine.VarRef("color-primary", "black")).
		Done().
		Build().
		Render(styleengine.RenderOptions{Pretty: true})
	if err != nil {
		panic(err)
	}
	fmt.Println(sheet)
	// Output:
	// :root {
	//   --color-primary: #0a84ff;
	//   color: var(--color-primary, black);
	// }
}

// Example_sanitize demonstrates rejecting dangerous untrusted CSS.
func Example_sanitize() {
	_, err := styleengine.Sanitize(`body { background: url("javascript:alert(1)"); }`)
	fmt.Println("sanitized err:", err != nil)
	// Output:
	// sanitized err: true
}

// Example_parseRoundtrip shows a simple parse + render cycle.
func Example_parseRoundtrip() {
	parsed, err := styleengine.Parse(`:root { --foo: red; } .bar { color: var(--foo); }`)
	if err != nil {
		panic(err)
	}
	out, _ := parsed.Render(styleengine.RenderOptions{Pretty: false})
	fmt.Println("roundtrip ok:", len(out) > 0)
	// Output:
	// roundtrip ok: true
}

// Example_atruleMedia demonstrates a nested @media at-rule (black-box coverage for at-rules claimed in README).
func Example_atruleMedia() {
	sheet, _ := styleengine.New().
		Media("(prefers-color-scheme: dark)", func(b *styleengine.Builder) {
			b.Rule("body").Decl("background", styleengine.Literal("#000")).Done()
		}).
		Build().
		Render(styleengine.RenderOptions{Pretty: true})
	fmt.Println(sheet)
	// Output:
	// @media (prefers-color-scheme: dark) {
	//   body {
	//     background: #000;
	//   }
	// }
}
