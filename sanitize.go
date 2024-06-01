// Implements: OSS-BB-001 (styleengine building block).
// Per: ADR-0029 (every file declares its purpose); Library Excellence Standard (see README provenance) (Phase 0 exemplar extraction).
// Discipline: C-14.

package styleengine

import (
	"fmt"
	"regexp"
	"strings"
)

// dangerousSubstrings flag patterns whose mere presence can break structural
// or security assumptions when CSS is accepted from untrusted sources.
// Matched case-insensitively after lowercasing. Categories:
//
//   - "}" — prevents escaping the surrounding rule block
//   - "@import" — would let untrusted input pull arbitrary remote stylesheets
//   - "expression(", "behavior:", "binding:" — legacy client-side script vectors
//   - "<script", "</style", "</script", "<style" — markup injection vectors.
//     The </style guard is especially important because browsers terminate
//     raw-text <style> elements at the first </style>.
//
// "<" alone is intentionally not listed — it can appear in attribute selectors.
// The specific tag prefixes are the real practical attack surface for
// declaration-block fragments.
var dangerousSubstrings = []string{
	"}",
	"@import",
	"expression(",
	"behavior:",
	"binding:",
	"<script",
	"<style",
	"</script",
	"</style",
}

// dangerousPatterns flag patterns whose syntactic shape requires whitespace
// tolerance (e.g. `url(  javascript:`). The regex set is case-insensitive.
var dangerousPatterns = []*regexp.Regexp{
	// url(...) of any shape: untrusted CSS should not pull external resources
	// unless you have a dedicated allowlist layer on top.
	regexp.MustCompile(`(?i)url\s*\(`),
	// CSS image / paint functions that accept URL or string arguments.
	regexp.MustCompile(`(?i)\b(image-set|cross-fade|paint|element|image)\s*\(`),
	// javascript: scheme even outside url().
	regexp.MustCompile(`(?i)javascript\s*:`),
}

// Sanitize trims and validates a CSS declaration-block fragment. It rejects
// tokens that could escape the containing rule, import remote resources, or
// introduce script execution.
//
// Call Sanitize on any CSS that comes from outside your trusted source tree
// (user input, customer theme overrides, plugin-provided fragments, etc.)
// before passing it to [Literal] or embedding it.
//
// Trusted, in-tree CSS (your design system tokens, official themes) should
// NOT go through Sanitize — it intentionally rejects url(), @font-face, and
// other constructs that are valid and useful inside first-party stylesheets.
//
// CSS escape sequences (e.g. `\75 rl(`) are rejected outright. Untrusted
// input has no legitimate need for them; allowing them would create bypasses.
func Sanitize(css string) (string, error) {
	css = strings.TrimSpace(css)
	if css == "" {
		return "", nil
	}
	if strings.ContainsRune(css, '\\') {
		return "", fmt.Errorf("styleengine: rejected CSS contains escape sequence")
	}
	lower := strings.ToLower(css)
	for _, pat := range dangerousSubstrings {
		if strings.Contains(lower, pat) {
			return "", fmt.Errorf("styleengine: rejected CSS contains unsafe pattern %q", pat)
		}
	}
	for _, re := range dangerousPatterns {
		if loc := re.FindStringIndex(css); loc != nil {
			return "", fmt.Errorf("styleengine: rejected CSS contains unsafe pattern %q", css[loc[0]:loc[1]])
		}
	}
	return css, nil
}
