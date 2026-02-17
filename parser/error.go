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

// errorf ...
func (p *parser) errorf(msg string, msgValues ...any) error {
	err := fmt.Errorf(msg, msgValues...)
	p.errors = errors.Join(p.errors, err)
	return err
}

// errorUnexpected ...
func (p *parser) errorUnexpected(chr rune) error {
	if chr == -1 {
		return p.errorf(errUnexpectedEndOfInput)
	}
	return p.errorf(errUnexpectedToken, token.Illegal)
}

func (p *parser) errorUnexpectedToken(tkn token.Token) error {
	switch tkn {
	case token.Eof:
		return p.errorf(errUnexpectedEndOfInput)
	case token.Boolean, token.Null:
		//value = p.literal TODO
	case token.Identifier:
		return p.errorf("Unexpected identifier")
	case token.Keyword:
		// TODO Might be a future reserved word
		return p.errorf("Unexpected reserved word")
	case token.EscapedReservedWord:
		return p.errorf("Keyword must not contain escaped characters")
	case token.Number:
		return p.errorf("Unexpected number")
	case token.String:
		return p.errorf("Unexpected string")
	}
	return p.errorf(errUnexpectedToken, tkn.String())
}
