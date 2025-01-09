package scanner

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

type Token struct {
	Kind      token.Token
	OnNewLine bool

	// 6 bytes of padding so that the next field starts at offset 8, the compiler automatically
	// does this, but for some reason it results in performance worse than manual padding.
	_ [6]byte

	escaped    *string
	Idx0, Idx1 ast.Idx
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
