// builder.go owns the fluent construction API for sheets and rules.
// Implements: REQ-011.
// Per: ADR-0029 (every file declares its purpose).
// Discipline: C-14.

package styleengine

// Builder is the fluent surface over Sheet.
type Builder struct {
	sheet *Sheet
}

// New returns a fresh Builder backed by an empty Sheet.
func New() *Builder {
	return &Builder{sheet: NewSheet()}
}

// Var delegates to Sheet.Var.
func (b *Builder) Var(name, value string) *Builder {
	b.sheet.Var(name, value)
	return b
}

// Rule starts a RuleBuilder for the given selector.
func (b *Builder) Rule(selector string) *RuleBuilder {
	return &RuleBuilder{parent: b, sel: MustSelector(selector)}
}

// Media wraps Sheet.Media.
func (b *Builder) Media(query string, fn func(*Builder)) *Builder {
	b.sheet.Media(query, func(inner *Sheet) {
		fn(&Builder{sheet: inner})
	})
	return b
}

// Layer wraps Sheet.Layer.
func (b *Builder) Layer(name string, fn func(*Builder)) *Builder {
	b.sheet.Layer(name, func(inner *Sheet) {
		fn(&Builder{sheet: inner})
	})
	return b
}

// Supports wraps Sheet.Supports.
func (b *Builder) Supports(condition string, fn func(*Builder)) *Builder {
	b.sheet.Supports(condition, func(inner *Sheet) {
		fn(&Builder{sheet: inner})
	})
	return b
}

// Keyframes wraps Sheet.Keyframes.
func (b *Builder) Keyframes(name string, fn func(*KeyframesBuilder)) *Builder {
	b.sheet.Keyframes(name, fn)
	return b
}

// FontFace appends a font-face declaration.
func (b *Builder) FontFace(ff FontFaceDecl) *Builder {
	b.sheet.AddFontFace(ff)
	return b
}

// Build returns the underlying Sheet. Subsequent builder calls continue to
// mutate the same Sheet.
func (b *Builder) Build() *Sheet { return b.sheet }

// RuleBuilder accumulates declarations for one selector.
type RuleBuilder struct {
	parent *Builder
	sel    Selector
	decls  []Declaration
}

// Decl adds a declaration.
func (r *RuleBuilder) Decl(property string, value Value) *RuleBuilder {
	r.decls = append(r.decls, Decl(property, value))
	return r
}

// ImportantDecl adds an !important declaration.
func (r *RuleBuilder) ImportantDecl(property string, value Value) *RuleBuilder {
	r.decls = append(r.decls, Decl(property, value).Important())
	return r
}

// Done commits the rule into the parent Builder's Sheet.
func (r *RuleBuilder) Done() *Builder {
	r.parent.sheet.AddRule(Rule{Selector: r.sel, Decls: r.decls})
	return r.parent
}
