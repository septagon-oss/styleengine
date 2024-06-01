// Implements: OSS-BB-001 (styleengine building block).
// Per: ADR-0029 (every file declares its purpose); Library Excellence Standard (see README provenance) (Phase 0 exemplar extraction).
// Discipline: C-14.

package styleengine

import (
	"fmt"
	"io"
	"strings"

	"github.com/tdewolff/parse/v2"
	pcss "github.com/tdewolff/parse/v2/css"
)

// Parse converts a CSS string into a Sheet. Only top-level qualified rules
// (selector { ... }) and their declarations are recognized in v1. At-rules
// and other grammar constructs are skipped without error.
func Parse(src string) (*Sheet, error) {
	p := pcss.NewParser(parse.NewInput(strings.NewReader(src)), false)
	s := NewSheet()
	if err := parseInto(p, s); err != nil {
		return nil, err
	}
	return s, nil
}

func parseInto(p *pcss.Parser, s *Sheet) error {
	var currentSel string
	var currentDecls []Declaration
	// atDepth tracks how many BeginAtRuleGrammar wrappers we are nested
	// inside. While atDepth > 0 we discard inner rules and declarations to
	// avoid leaking them to the top-level Sheet. v1 does not preserve
	// at-rule wrappers from parsed input; consumers should construct them
	// programmatically via Sheet.Media/Layer/Supports/Keyframes instead.
	atDepth := 0

	flush := func() {
		if currentSel == "" {
			return
		}
		sel, err := NewSelector(currentSel)
		if err == nil {
			s.AddRule(Rule{Selector: sel, Decls: currentDecls})
		}
		currentSel = ""
		currentDecls = nil
	}

	for {
		gt, _, data := p.Next()

		if gt == pcss.ErrorGrammar {
			if err := p.Err(); err != nil && err != io.EOF {
				return fmt.Errorf("styleengine: parse: %w", err)
			}
			flush()
			return nil
		}

		switch gt {
		case pcss.BeginAtRuleGrammar:
			atDepth++

		case pcss.EndAtRuleGrammar:
			if atDepth > 0 {
				atDepth--
			}

		case pcss.BeginRulesetGrammar:
			if atDepth > 0 {
				continue
			}
			// Values() contains all selector tokens; data holds the last token
			// before '{'. Prefer Values() for the full selector text.
			sel := joinValueTokens(p.Values())
			if sel == "" {
				sel = strings.TrimSpace(string(data))
			}
			currentSel = strings.TrimSpace(sel)
			currentDecls = nil

		case pcss.QualifiedRuleGrammar:
			if atDepth > 0 {
				continue
			}
			// Intermediate selector part (comma-separated multi-selector).
			// Accumulate into currentSel.
			part := joinValueTokens(p.Values())
			if part == "" {
				part = strings.TrimSpace(string(data))
			}
			if currentSel == "" {
				currentSel = strings.TrimSpace(part)
			} else {
				currentSel += strings.TrimSpace(part)
			}

		case pcss.DeclarationGrammar:
			if atDepth > 0 {
				continue
			}
			prop := strings.TrimSpace(string(data))
			valueRaw := strings.TrimSpace(joinValueTokens(p.Values()))
			important := false
			lv := strings.ToLower(valueRaw)
			if strings.HasSuffix(lv, "!important") {
				important = true
				valueRaw = strings.TrimSpace(valueRaw[:len(valueRaw)-len("!important")])
			}
			d := Decl(prop, Literal(valueRaw))
			if important {
				d = d.Important()
			}
			currentDecls = append(currentDecls, d)

		case pcss.CustomPropertyGrammar:
			if atDepth > 0 {
				continue
			}
			prop := strings.TrimSpace(string(data))
			valueRaw := strings.TrimSpace(joinValueTokens(p.Values()))
			d := Decl(prop, Literal(valueRaw))
			currentDecls = append(currentDecls, d)

		case pcss.EndRulesetGrammar:
			if atDepth > 0 {
				continue
			}
			flush()

			// AtRuleGrammar (one-shot at-rules like @charset, @import) and
			// Comment/Token grammar types are intentionally skipped in v1.
		}
	}
}

// joinValueTokens concatenates the Data bytes of all tokens into a single
// string. The parser may include WhitespaceToken entries between values; these
// are included as-is to preserve round-trip fidelity.
func joinValueTokens(vs []pcss.Token) string {
	if len(vs) == 0 {
		return ""
	}
	var b strings.Builder
	for _, v := range vs {
		b.Write(v.Data)
	}
	return b.String()
}
