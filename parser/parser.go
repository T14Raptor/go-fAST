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

	chr       rune // The current character
	chrOffset int  // The offset of current character

	scope             *scope
	implicitSemicolon bool // An implicit semicolon exists

	errors ErrorList

	recover struct {
		// Scratch when trying to seek to the next statement, etc.
		idx   ast.Idx
		count int
	}

	exprArena *miniArena[ast.Expression]
	stmtArena *miniArena[ast.Statement]
}

// newParser ...
func newParser(src string) *parser {
	return &parser{
		str: src,

		scanner: scanner.NewScanner(src),

		exprArena: newArena[ast.Expression](1024),
		stmtArena: newArena[ast.Statement](1024),
	}
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
	return program, p.errors.Err()
}

// next ...
func (p *parser) next() {
	p.token = p.scanner.Next()
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

func (p *parser) makeExpr(expr ast.Expr) *ast.Expression {
	expression := p.exprArena.make()
	expression.Expr = expr
	return expression
}

func (p *parser) makeStmt(stmt ast.Stmt) *ast.Statement {
	statement := p.stmtArena.make()
	statement.Stmt = stmt
	return statement
}
