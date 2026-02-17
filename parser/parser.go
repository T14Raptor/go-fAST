package parser

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/parser/scanner"
	"github.com/t14raptor/go-fast/token"
)

// parser ...
type parser struct {
	token scanner.Token
	str   string

	scanner *scanner.Scanner

	scope             *scope
	implicitSemicolon bool // An implicit semicolon exists

	errors error

	recover struct {
		// Scratch when trying to seek to the next statement, etc.
		idx   ast.Idx
		count int
	}

	alloc nodeAllocator
}

// newParser ...
func newParser(src string) *parser {
	p := &parser{
		str: src,

		alloc: newNodeAllocator(),
	}
	p.scanner = scanner.NewScanner(src, &p.errors)
	return p
}

// ParseFile parses the source code of a single JavaScript/ECMAScript source file and returns
// the corresponding ast.Program node.
func ParseFile(src string) (*ast.Program, error) {
	return newParser(src).parse()
}

// parse ...
func (p *parser) parse() (*ast.Program, error) {
	p.openScope()
	p.next()
	program := p.parseProgram()
	p.closeScope()
	return program, p.errors
}

// next ...
func (p *parser) next() {
	p.scanner.Next()
	p.token = p.scanner.Token
}

type parserState struct {
	c scanner.Checkpoint

	tok scanner.Token

	errors error
}

func (p *parser) mark() parserState {
	return parserState{
		c:      p.scanner.Checkpoint(),
		tok:    p.token,
		errors: p.errors,
	}
}

func (p *parser) restore(state parserState) {
	p.scanner.Rewind(state.c)
	p.token = state.tok
	// Truncate parser errors back to checkpoint state
	p.errors = state.errors
}

func (p *parser) peek() scanner.Token {
	st := p.mark()
	p.scanner.Next()
	tok := p.scanner.Token
	p.restore(st)
	return tok
}

func (p *parser) currentString() string {
	return p.token.String(p.scanner)
}

func (p *parser) currentKind() token.Token {
	return p.token.Kind
}

func (p *parser) currentOffset() ast.Idx {
	return p.token.Idx0
}

func (p *parser) canInsertSemicolon() bool {
	kind := p.currentKind()
	return kind == token.Semicolon || kind == token.RightBrace /*|| p.scanner.EOF()*/ || p.token.OnNewLine
}

func (p *parser) semicolon() bool {
	if !p.canInsertSemicolon() {
		return false
	}

	if p.currentKind() == token.Semicolon {
		p.next()
	}
	return true
}

func (p *parser) idxOf(offset int) ast.Idx {
	return ast.Idx(1 + offset)
}

func (p *parser) expect(value token.Token) ast.Idx {
	idx := p.scanner.Offset()
	if p.token.Kind != value {
		p.errorUnexpectedToken(p.token.Kind)
	}
	p.next()
	return idx
}
