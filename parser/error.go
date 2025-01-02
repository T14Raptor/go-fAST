package parser

import (
	"fmt"
	"github.com/t14raptor/go-fast/token"
)

const (
	errUnexpectedToken      = "Unexpected token %v"
	errUnexpectedEndOfInput = "Unexpected end of input"
)

// error ...
func (p *parser) error(msg string, msgValues ...any) error {
	msg = fmt.Sprintf(msg, msgValues...)
	p.errors.Add(msg)
	return p.errors[len(p.errors)-1]
}

// errorUnexpected ...
func (p *parser) errorUnexpected(chr rune) error {
	if chr == -1 {
		return p.error(errUnexpectedEndOfInput)
	}
	return p.error(errUnexpectedToken, token.Illegal)
}

func (p *parser) errorUnexpectedToken(tkn token.Token) error {
	//debug.PrintStack()
	//fmt.Println("unexpected", tkn.String())
	switch tkn {
	case token.Eof:
		return p.error(errUnexpectedEndOfInput)
	}
	value := tkn.String()
	switch tkn {
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
	return p.error(errUnexpectedToken, value)
}

// ErrorList is a list of *Errors.
type ErrorList []error

// Add adds an Error with given position and message to an ErrorList.
func (e *ErrorList) Add(msg string) {
	*e = append(*e, fmt.Errorf(msg))
}

// Error implements the Error interface.
func (e *ErrorList) Error() string {
	switch len(*e) {
	case 0:
		return "no errors"
	case 1:
		return (*e)[0].Error()
	}
	return fmt.Sprintf("%s (and %d more errors)", (*e)[0].Error(), len(*e)-1)
}

// Err returns an error equivalent to this ErrorList. If the list is empty, Err returns nil.
func (e *ErrorList) Err() error {
	if len(*e) == 0 {
		return nil
	}
	return e
}
