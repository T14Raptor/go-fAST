package parser

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

// parser ...
type parser struct {
	str    string
	length int

	chr       rune // The current character
	chrOffset int  // The offset of current character
	offset    int  // The offset after current character (could be greater than 1)

	idx           ast.Idx     // The index of token
	token         token.Token // The token
	literal       string      // The literal of the token, if any
	parsedLiteral string

	scope             *scope
	insertSemicolon   bool // If we see a newline, then insert an implicit semicolon
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
		chr:    ' ',
		str:    src,
		length: len(src),

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
