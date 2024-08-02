package parser

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
	"github.com/t14raptor/go-fast/unistring"
)

// parser ...
type parser struct {
	str    string
	length int

	chr       rune // The current character
	chrOffset int  // The offset of current character
	offset    int  // The offset after current character (may be greater than 1)

	idx           ast.Idx     // The index of token
	token         token.Token // The token
	literal       string      // The literal of the token, if any
	parsedLiteral unistring.String

	scope             *scope
	insertSemicolon   bool // If we see a newline, then insert an implicit semicolon
	implicitSemicolon bool // An implicit semicolon exists

	errors ErrorList

	recover struct {
		// Scratch when trying to seek to the next statement, etc.
		idx   ast.Idx
		count int
	}
}

// newParser ...
func newParser(src string) *parser {
	return &parser{
		chr:    ' ',
		str:    src,
		length: len(src),
	}
}

// ParseFile parses the source code of a single JavaScript/ECMAScript source file and returns
// the corresponding ast.Program node.
func ParseFile(src string) (*ast.Program, error) {
	p := newParser(src)
	return p.parse()
}

// slice ...
func (p *parser) slice(idx0, idx1 ast.Idx) string {
	from := int(idx0) - 1
	to := int(idx1) - 1
	if from >= 0 && to <= len(p.str) {
		return p.str[from:to]
	}
	return ""
}

// parse ...
func (p *parser) parse() (*ast.Program, error) {
	p.openScope()
	defer p.closeScope()
	p.next()
	program := p.parseProgram()
	return program, p.errors.Err()
}

// next ...
func (p *parser) next() {
	p.token, p.literal, p.parsedLiteral, p.idx = p.scan()
}

func (p *parser) optionalSemicolon() {
	if p.token == token.Semicolon {
		p.next()
		return
	}

	if p.implicitSemicolon {
		p.implicitSemicolon = false
		return
	}

	if p.token != token.Eof && p.token != token.RightBrace {
		p.expect(token.Semicolon)
	}
}

func (p *parser) semicolon() {
	if p.token != token.RightParenthesis && p.token != token.RightBrace {
		if p.implicitSemicolon {
			p.implicitSemicolon = false
			return
		}

		p.expect(token.Semicolon)
	}
}

func (p *parser) idxOf(offset int) ast.Idx {
	return ast.Idx(1 + offset)
}

func (p *parser) expect(value token.Token) ast.Idx {
	idx := p.idx
	if p.token != value {
		p.errorUnexpectedToken(p.token)
	}
	p.next()
	return idx
}
