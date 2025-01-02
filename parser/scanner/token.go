package scanner

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

type Token struct {
	Kind token.Token

	Idx0, Idx1 ast.Idx

	OnNewLine bool

	escaped *string
}

func (t Token) String(s *Scanner) string {
	if t.escaped != nil {
		return *t.escaped
	}

	raw := s.src.Slice(t.Idx0, t.Idx1)
	switch t.Kind {
	case token.String:
		return raw[1 : len(raw)-1]
	case token.PrivateIdentifier:
		return raw[1:]
	}
	return raw
}

func (t Token) Raw(s *Scanner) string {
	return s.src.Slice(t.Idx0, t.Idx1)
}
