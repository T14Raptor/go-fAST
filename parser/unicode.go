package parser

import (
	"github.com/nukilabs/unicodeid"
	"github.com/t14raptor/go-fast/token"
	"unicode"
)

func (p *parser) handleUnicodeChar() token.Token {
	switch c := p._peek(); {
	case unicodeid.IsIDStartUnicode(c):
		p.scanIdentifierTail(p.offset)
		return token.Identifier
	case unicode.IsSpace(c):
		p.read()
		return token.Skip
	case isLineTerminator(c):
		p.read()
		return token.Skip
	default:
		p.read()
		p.errorUnexpected(c)
		return token.Undetermined
	}
}
