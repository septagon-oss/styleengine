// Implements: REQ-011.
// Per: ADR-0029 (every file declares its purpose).
// Discipline: C-14.

// Package styleengine provides a typed, safe, and fast Go-native CSS
// construction and emission engine.
//
// It is useful for design systems, theme compilers, server-rendered UIs,
// email templates, and any situation where you need to build or sanitize
// CSS programmatically without string concatenation or ad-hoc templates.
//
// A [Sheet] is the root IR: an ordered, deduped, mergeable collection of
// [Rule]s and [AtRule]s. The fluent [Builder] produces a Sheet from typed
// inputs; [Render] emits CSS (pretty or minified); [Parse] reverses the
// process (e.g. for validating or normalizing untrusted CSS fragments).
//
// styleengine is intentionally narrow: no CSS-in-JS, no preprocessor
// syntax, no third-party language extensions. Modern native CSS only.
//
// # Basic usage
//
// Create a sheet with variables and rules, then render:
//
//	sheet, err := styleengine.New().
//	    Var("color-primary", "#0a84ff").
//	    Rule(":root").
//	        Decl("color", styleengine.VarRef("color-primary", "black")).
//	        Done().
//	    Build().
//	    Render(styleengine.RenderOptions{Pretty: true})
//
// # Sanitizing untrusted input
//
// Always sanitize CSS that comes from users, plugins, or customer overrides:
//
//	clean, err := styleengine.Sanitize(userSuppliedCSS)
//
// # Parsing existing CSS
//
// Round-trip existing fragments for validation or normalization:
//
//	parsed, err := styleengine.Parse(someCSS)
//
// See the [examples] directory and the README for more complete programs.
package styleengine
