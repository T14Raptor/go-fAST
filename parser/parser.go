package parser

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/parser/scanner"
	"github.com/t14raptor/go-fast/token"
)

// parser ...
type parser struct {
	scanner scanner.Scanner

	str string

	scope *scope

	errors  error
	recover struct {
		// Scratch when trying to seek to the next statement, etc.
		idx   ast.Idx
		count int
	}

	alloc nodeAllocator

	// Scratch buffers used as a stack for building Expression/Statement
	// slices without per-call heap allocations. Each builder saves
	// len(buf) as a mark, appends elements, copies the subslice to the
	// arena, then restores buf to the saved mark.
	exprBuf []ast.Expression
	stmtBuf []ast.Statement
	propBuf []ast.Property
	elemBuf []ast.ClassElement
	declBuf []ast.VariableDeclarator
}

// newParser ...
func newParser(src string) *parser {
	p := &parser{
		str: src,

		alloc:   newNodeAllocator(),
		exprBuf: make([]ast.Expression, 0, 64),
		stmtBuf: make([]ast.Statement, 0, 64),
		propBuf: make([]ast.Property, 0, 16),
		elemBuf: make([]ast.ClassElement, 0, 16),
		declBuf: make([]ast.VariableDeclarator, 0, 16),
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
}

type parserState struct {
	c scanner.Checkpoint

	errors error
}

func (p *parser) mark() parserState {
	return parserState{
		c:      p.scanner.Checkpoint(),
		errors: p.errors,
	}
}

func (p *parser) restore(state parserState) {
	p.scanner.Rewind(state.c)
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
	return p.scanner.Token.String(p.scanner)
}

func (p *parser) currentKind() token.Token {
	return p.scanner.Token.Kind
}

func (p *parser) currentOffset() ast.Idx {
	return p.scanner.Token.Idx0
}

func (p *parser) canInsertSemicolon() bool {
	kind := p.currentKind()
	return kind == token.Semicolon || kind == token.RightBrace || kind == token.Eof || p.scanner.Token.OnNewLine
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

// finishExprBuf copies exprBuf[mark:] into an arena-backed Expressions slice
// and restores the scratch buffer to the saved mark.
func (p *parser) finishExprBuf(mark int) ast.Expressions {
	result := p.alloc.CopyExpressions(p.exprBuf[mark:])
	p.exprBuf = p.exprBuf[:mark]
	return result
}

// finishStmtBuf copies stmtBuf[mark:] into an arena-backed Statements slice
// and restores the scratch buffer to the saved mark.
func (p *parser) finishStmtBuf(mark int) ast.Statements {
	result := p.alloc.CopyStatements(p.stmtBuf[mark:])
	p.stmtBuf = p.stmtBuf[:mark]
	return result
}

// finishPropBuf copies propBuf[mark:] into a heap-allocated Properties slice
// and restores the scratch buffer to the saved mark.
func (p *parser) finishPropBuf(mark int) ast.Properties {
	src := p.propBuf[mark:]
	if len(src) == 0 {
		p.propBuf = p.propBuf[:mark]
		return nil
	}
	dst := make(ast.Properties, len(src))
	copy(dst, src)
	p.propBuf = p.propBuf[:mark]
	return dst
}

// finishDeclBuf copies declBuf[mark:] into a heap-allocated VariableDeclarators slice
// and restores the scratch buffer to the saved mark.
func (p *parser) finishDeclBuf(mark int) ast.VariableDeclarators {
	src := p.declBuf[mark:]
	if len(src) == 0 {
		p.declBuf = p.declBuf[:mark]
		return nil
	}
	dst := make(ast.VariableDeclarators, len(src))
	copy(dst, src)
	p.declBuf = p.declBuf[:mark]
	return dst
}

// finishElemBuf copies elemBuf[mark:] into a heap-allocated ClassElements slice
// and restores the scratch buffer to the saved mark.
func (p *parser) finishElemBuf(mark int) ast.ClassElements {
	src := p.elemBuf[mark:]
	if len(src) == 0 {
		p.elemBuf = p.elemBuf[:mark]
		return nil
	}
	dst := make(ast.ClassElements, len(src))
	copy(dst, src)
	p.elemBuf = p.elemBuf[:mark]
	return dst
}

func (p *parser) expect(value token.Token) ast.Idx {
	idx := p.scanner.Offset()
	if p.scanner.Token.Kind != value {
		p.errorUnexpectedToken(p.scanner.Token.Kind)
	}
	p.next()
	return idx
}
