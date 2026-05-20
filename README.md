# styleengine

Typed, composable, fast Go-native CSS construction and emission.

`styleengine` replaces ad-hoc `text/template` or string concatenation with a
structured intermediate representation (IR) that validates inputs at construction
time, deduplicates rules, enforces a byte-size safety ceiling, and emits either
pretty-printed or minified CSS through a single `Render` call.

It is a great building block for design system compilers, theme engines,
plugin and customization systems, and server-side rendering pipelines.

---

## Who calls it

- **Design-system token compilers** — emit `:root` custom-property sheets from
  design-token sources.
- **Theme / experience composers** — assemble layered overrides (brand,
  customer, user) into a single merged `Sheet`.
- **Untrusted CSS validators** — call [Sanitize] on user-, plugin-, or
  customer-supplied CSS fragments before accepting them.
- **Server renderers and email generators** — build and emit CSS safely
  without template string hacks.
- Custom tooling and CLIs that need reproducible, validated CSS output.

---

## Hello world

```go
package main

import (
	"fmt"
	"github.com/septagon-oss/styleengine"
)

func main() {
	sheet, err := styleengine.New().
		Var("font-display", "swap").
		Var("motion-duration-marquee", "8s").
		Rule(":root").
			Decl("font-display", styleengine.VarRef("font-display", "swap")).
			Done().
		Rule(".marquee").
			Decl("animation-duration", styleengine.VarRef("motion-duration-marquee", "6s")).
			Decl("overflow", styleengine.Literal("hidden")).
			Done().
		Build().
		Render(styleengine.RenderOptions{Pretty: true})
	if err != nil {
		panic(err)
	}
	fmt.Println(sheet)
}
```

`Var` registers a `--font-display` custom property in `:root`. `VarRef` emits
`var(--font-display, swap)`. Both enforce the `[a-z][a-z0-9-]*` naming rule at
call time, so mis-spellings are caught immediately rather than silently emitted.

---

## Public API surface

| Symbol | Kind | Role |
|---|---|---|
| `Builder` | struct | Fluent entry point; wraps a `Sheet` |
| `Sheet` | struct | Root IR — ordered, deduped rule + at-rule collection |
| `Rule` | struct | `(Selector, []Declaration)` pair |
| `Selector` | struct | Normalized CSS selector |
| `Declaration` | struct | `property: value [!important]` |
| `Value` | interface | Typed CSS RHS; self-renders via `.CSS()` |
| `Literal` | func | Wraps a trusted string as a `Value` |
| `VarRef` | func | Emits `var(--name[, fallback])` as a `Value` |
| `Decl` | func | Constructor for `Declaration` |
| `RuleBuilder` | struct | Accumulates declarations; `.Done()` commits to parent |
| `KeyframesBuilder` | struct | Accumulates `@keyframes` stops |
| `FontFaceDecl` | struct | Structured `@font-face` descriptor |
| `RenderOptions` | struct | Controls pretty/minify/MaxBytes for `Render` |
| `DefaultMaxBytes` | const | Default 1 MiB safety ceiling |

---

## RenderOptions

```go
type RenderOptions struct {
	Pretty   bool  // emit indented multi-line output
	Minify   bool  // run tdewolff/minify (overrides Pretty)
	MaxBytes int64 // 0 → DefaultMaxBytes (1 MiB)
}
```

`Render` always enforces a byte ceiling. When `MaxBytes` is zero it falls back
to `DefaultMaxBytes` (1 MiB = `1 << 20`). If the rendered output exceeds the
limit `Render` returns an error rather than a truncated string, so callers
cannot silently serve corrupt CSS.

---

## Variable safety

`Var` and `VarRef` both enforce `[a-z][a-z0-9-]*` on the custom-property name
and panic immediately if the constraint is violated. The leading `--` is added
automatically — callers never write it.

`UndefinedVars() []string` scans the sheet for `var(--x)` references that have
no matching `Var` registration and returns the sorted list of undefined names.
Use this in compiler pipelines to surface missing token definitions before
emitting CSS to a browser.

---

## Sanitize

Before accepting CSS text from untrusted sources (user input, customer
overrides, third-party plugins, etc.), pass it through `Sanitize`:

```go
clean, err := styleengine.Sanitize(userInput)
if err != nil {
	// reject the request
}
```

`Sanitize` trims whitespace and rejects (case-insensitively) any fragment
containing dangerous patterns that could break out of the expected stylesheet
context or introduce security issues.

---

## Parse roundtrip

`Parse(css string) (*Sheet, error)` ingests a CSS string via
`github.com/tdewolff/parse/v2/css` and reconstructs a `Sheet` IR.

For the rule + custom-property subset supported in v1, the roundtrip is
idempotent. This property is verified in the test suite and is useful for
overlay / customization pipelines.

---

## At-rules supported

| At-rule | Status |
|---|---|
| `@media` | Supported |
| `@keyframes` | Supported |
| `@font-face` | Supported |
| `@layer` | Supported |
| `@supports` | Supported |
| `@scope` | Deferred |
| `@container` | Deferred |

Nested at-rules are composed via closures that receive a child `Sheet`.

---

## Performance

Both pretty and minified render are well under the 500 µs target for a
full-page token sheet. See benchmarks in the package for current numbers on
your hardware.

---

## Dependencies

| Module | Role |
|---|---|
| `github.com/tdewolff/minify/v2` | CSS minification (direct) |
| `github.com/tdewolff/parse/v2` | CSS tokenizer used by `Parse` (direct) |

Both are pure Go with no cgo requirements.

---

## Status

The core library (IR, builder, render, parse, sanitize, variables, at-rules)
is stable and suitable for general use.

The package has comprehensive tests, fuzzing, and benchmarks. It has
minimal dependencies.

See the godoc for the full API and the `Example` functions (including roundtrips
and at-rule usage) that render in documentation and pass `go test -run Example`.

## Provenance

Extracted from PlatformKit (septagon-dev) as a reusable building block.
Original lived in platformkit-shared/styleengine. Made fully generic and
zero-PlatformKit for the OSS community.

See the case study in the PlatformKit docs for extraction details and the
Library Excellence Standard applied (rich godoc, black-box Examples, C-14
discipline on every file, perfect incremental git history).




## 2026 Polish for platformkit-courses

This library (with its history) is intended as a citable example of modern Go
library craft: real incremental development, rich godoc with black-box Examples, C-14 discipline, and exemplary git history.
