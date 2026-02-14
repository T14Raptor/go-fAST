package parser

import (
	"errors"
	"fmt"
	"github.com/t14raptor/go-fast/token"
)

const (
	errUnexpectedToken      = "Unexpected token %v"
	errUnexpectedEndOfInput = "Unexpected end of input"
)

// error ...
func (p *parser) error(msg string, msgValues ...any) error {
	err := fmt.Errorf(msg, msgValues...)
	p.errors = errors.Join(p.errors, err)
	return err
}

// errorUnexpected ...
func (p *parser) errorUnexpected(chr rune) error {
	if chr == -1 {
		return p.error(errUnexpectedEndOfInput)
	}
	return p.error(errUnexpectedToken, token.Illegal)
}

func (p *parser) errorUnexpectedToken(tkn token.Token) error {
	switch tkn {
	case token.Eof:
		return p.error(errUnexpectedEndOfInput)
	case token.Boolean, token.Null:
		//value = p.literal TODO
	case token.Identifier:
		return p.error("Unexpected identifier")
	case token.Keyword:
		// TODO Might be a future reserved word
		return p.error("Unexpected reserved word")
	case token.EscapedReservedWord:
		return p.error("Keyword must not contain escaped characters")
	case token.Number:
		return p.error("Unexpected number")
	case token.String:
		return p.error("Unexpected string")
	}
	return p.error(errUnexpectedToken, tkn.String())
}
