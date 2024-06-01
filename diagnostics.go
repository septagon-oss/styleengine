// Implements: OSS-BB-001 (styleengine building block).
// Per: ADR-0029 (every file declares its purpose); Library Excellence Standard (see README provenance) (Phase 0 exemplar extraction).
// Discipline: C-14.

package styleengine

import "sort"

// DiagnosticKind classifies the severity of a [Diagnostic].
type DiagnosticKind int

const (
	// DiagnosticError indicates a finding that must be fixed; emitting the
	// CSS as-is would produce broken or unsafe output.
	DiagnosticError DiagnosticKind = iota
	// DiagnosticWarning indicates a likely bug or stylistic issue that
	// should be reviewed but does not block emission.
	DiagnosticWarning
	// DiagnosticHint indicates an opportunity for improvement; safe to ignore.
	DiagnosticHint
)

func (k DiagnosticKind) String() string {
	switch k {
	case DiagnosticError:
		return "error"
	case DiagnosticWarning:
		return "warning"
	case DiagnosticHint:
		return "hint"
	}
	return "unknown"
}

// Diagnostic is a stable, typed finding produced by [Sheet.Diagnostics].
// Downstream linters, compilers, and validation pipelines can consume these
// directly instead of regex-scanning rendered CSS text.
type Diagnostic struct {
	Code    string         // stable code (e.g., "SE001")
	Kind    DiagnosticKind // severity
	Message string         // human-readable description
	Subject string         // optional reference: selector, var name, etc.
}

// Stable diagnostic codes. Adding a new code is additive; reusing or
// reassigning an existing one breaks downstream tooling.
const (
	// DiagUndefinedVar — a var(--x) reference has no matching Var
	// registration in the Sheet. Subject is the missing name.
	DiagUndefinedVar = "SE001"
)

// Diagnostics returns the current set of typed findings for the Sheet.
// The result is sorted by (Code, Subject) for deterministic output.
//
// v1 reports SE001 for undefined var refs. Future codes follow the
// "SE0xx" prefix; consumers should treat unknown codes as informational.
func (s *Sheet) Diagnostics() []Diagnostic {
	if s == nil {
		return nil
	}
	var out []Diagnostic
	for _, name := range s.UndefinedVars() {
		out = append(out, Diagnostic{
			Code:    DiagUndefinedVar,
			Kind:    DiagnosticError,
			Message: "var(--" + name + ") has no matching definition in the Sheet",
			Subject: name,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Code != out[j].Code {
			return out[i].Code < out[j].Code
		}
		return out[i].Subject < out[j].Subject
	})
	return out
}
