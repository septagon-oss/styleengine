// render.go owns pretty, minified, and size-bounded CSS emission.
// Implements: REQ-011.
// Per: ADR-0029 (every file declares its purpose).
// Discipline: C-14.

package styleengine

import (
	"bytes"
	"fmt"

	"github.com/tdewolff/minify/v2"
	mincss "github.com/tdewolff/minify/v2/css"
)

// RenderOptions configures [Sheet.Render].
//
// Pretty and Minify are mutually exclusive: setting both returns an error.
// MaxBytes caps the final output size; 0 means [DefaultMaxBytes].
type RenderOptions struct {
	Pretty   bool
	Minify   bool
	MaxBytes int64
}

// DefaultMaxBytes is the default Render byte ceiling (1 MiB).
const DefaultMaxBytes int64 = 1 << 20

// Render emits CSS per options. Returns an error if MaxBytes is exceeded
// or if both Pretty and Minify are set.
func (s *Sheet) Render(opts RenderOptions) (string, error) {
	if opts.Pretty && opts.Minify {
		return "", fmt.Errorf("styleengine: RenderOptions Pretty and Minify are mutually exclusive")
	}
	if s == nil {
		return "", nil
	}
	pretty := s.RenderPretty()
	var out string
	if opts.Minify {
		m := minify.New()
		m.AddFunc("text/css", mincss.Minify)
		var buf bytes.Buffer
		if err := m.Minify("text/css", &buf, bytes.NewReader([]byte(pretty))); err != nil {
			return "", fmt.Errorf("styleengine: minify: %w", err)
		}
		out = buf.String()
	} else {
		out = pretty
	}
	if err := enforceMaxBytes(out, opts.MaxBytes); err != nil {
		return "", err
	}
	return out, nil
}

// enforceMaxBytes checks the output against the requested ceiling. A zero
// limit uses [DefaultMaxBytes]; a negative limit disables the check
// (caller opt-in for one-shot batch jobs that know they exceed the
// runtime-safe default).
func enforceMaxBytes(out string, limit int64) error {
	if limit < 0 {
		return nil
	}
	if limit == 0 {
		limit = DefaultMaxBytes
	}
	if int64(len(out)) > limit {
		return fmt.Errorf("styleengine: rendered output (%d bytes) exceeds MaxBytes (%d)", len(out), limit)
	}
	return nil
}
