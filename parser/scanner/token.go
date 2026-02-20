package scanner

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

type Token struct {
	Kind token.Token

	OnNewLine bool
	HasEscape bool

	Idx0, Idx1 ast.Idx // 8 bytes
}

func (t Token) String(s Scanner) string {
	if t.HasEscape {
		return s.EscapedStr
	}

	raw := s.src.Slice(t.Idx0, t.Idx1)
	switch t.Kind {
	case token.String:
		return raw[1 : len(raw)-1]
	case token.PrivateIdentifier:
		return raw[1:]
	case token.NoSubstitutionTemplate, token.TemplateTail:
		// Strip opening (` or }) and closing (`)
		return raw[1 : len(raw)-1]
	case token.TemplateHead, token.TemplateMiddle:
		// Strip opening (` or }) and closing (${)
		return raw[1 : len(raw)-2]
	}
	return raw
}

func (t Token) Raw(s Scanner) string {
	return s.src.Slice(t.Idx0, t.Idx1)
}

// TemplateLiteral returns the raw source text of a template literal element
// (without the surrounding delimiters).
func (t Token) TemplateLiteral(s Scanner) string {
	raw := s.src.Slice(t.Idx0, t.Idx1)
	switch t.Kind {
	case token.NoSubstitutionTemplate, token.TemplateTail:
		// ` ... ` or } ... `
		return raw[1 : len(raw)-1]
	case token.TemplateHead, token.TemplateMiddle:
		// ` ... ${ or } ... ${
		return raw[1 : len(raw)-2]
	}
	return raw
}

// TemplateParsed returns the escape-processed content of a template literal element.
// If no escapes were present, returns the same as TemplateLiteral.
func (t Token) TemplateParsed(s Scanner) string {
	if t.HasEscape {
		return s.EscapedStr
	}
	return t.TemplateLiteral(s)
}
