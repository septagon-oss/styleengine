// main.go demonstrates standalone use of the typed CSS intermediate
// representation.
// Implements: REQ-011.
// Per: ADR-0029 (every file declares its purpose).
// Discipline: C-14.

//go:build ignore

// Minimal example of using styleengine as a standalone, generic typed CSS IR.
// Build/run with: go run examples/minimal/main.go

package main

import (
	"fmt"

	"github.com/septagon-oss/styleengine"
)

func main() {
	sheet, err := styleengine.New().
		Var("color-primary", "#0a84ff").
		Var("space-2", "0.5rem").
		Rule(":root").
		Decl("color", styleengine.VarRef("color-primary", "#000")).
		Decl("padding", styleengine.VarRef("space-2", "0")).
		Done().
		Build().
		Render(styleengine.RenderOptions{Pretty: true})
	if err != nil {
		panic(err)
	}
	fmt.Println(sheet)
}
