package parser

import (
	"errors"
	"fmt"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/parser/scanner/token"
)

const (
	errUnexpectedToken      = "Unexpected token %v"
	errUnexpectedEndOfInput = "Unexpected end of input"
)

// Error is a parser-level syntax error. It carries the byte range of the
// offending token in the original source.
//
// Errors are accumulated via [errors.Join]; callers can extract individual
// Errors with [errors.As], or use the shared [ast.Positioned] interface
// to handle parser and scanner errors uniformly.
type Error struct {
	Message string
	Start   ast.Idx
	End     ast.Idx
}

func (e *Error) Error() string {
	return e.Message
}

// Pos implements [ast.Positioned].
func (e *Error) Pos() (start, end ast.Idx) {
	return e.Start, e.End
}

// errorAt records a parser error at the given source range. Start and end
// are byte offsets into the original source.
func (p *parser) errorAt(start, end ast.Idx, msg string, args ...any) error {
	e := &Error{
		Message: fmt.Sprintf(msg, args...),
		Start:   start,
		End:     end,
	}
	p.errors = errors.Join(p.errors, e)
	return e
}

// errorf records a parser error at the current scanner offset. Use
// errorAt when the relevant position is not the current token.
func (p *parser) errorf(msg string, args ...any) error {
	idx := p.currentOffset()
	return p.errorAt(idx, idx, msg, args...)
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
