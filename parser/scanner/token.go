package scanner

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

type Token struct {
	kind token.Token

	idx0, idx1 ast.Idx

	onNewLine bool

	escaped *string
}

func (t Token) Kind() token.Token {
	return t.kind
}

func (t Token) WithKind(k token.Token) Token {
	t.kind = k
	return t
}

func (t Token) Idx0() ast.Idx {
	return t.idx0
}

func (t Token) Idx1() ast.Idx {
	return t.idx1
}

func (t Token) OnNewLine() bool {
	return t.onNewLine
}

func (t Token) String(s *Scanner) string {
	if t.escaped != nil {
		return *t.escaped
	}

	raw := s.src.Slice(t.idx0, t.idx1)
	switch t.kind {
	case token.String:
		return raw[1 : len(raw)-1]
	case token.PrivateIdentifier:
		return raw[1:]
	}
	return raw
}

func (t Token) Raw(s *Scanner) string {
	return s.src.Slice(t.idx0, t.idx1)
}
