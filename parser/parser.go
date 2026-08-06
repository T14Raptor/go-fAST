package parser

import (
	"sync"
	"unsafe"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/parser/scanner"
	"github.com/t14raptor/go-fast/parser/scanner/token"
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

	options parseOptions

	// Scratch buffers used as a stack for building Expression/Statement
	// slices without per-call heap allocations. Each builder saves
	// len(buf) as a mark, appends elements, copies the subslice to the
	// arena, then restores buf to the saved mark.
	exprBuf    []ast.Expression
	stmtBuf    []ast.Statement
	propBuf    []ast.Property
	elemBuf    []ast.ClassElement
	declBuf    []ast.VariableDeclarator
	patBuf     []ast.Pattern
	patPropBuf []ast.PatternProperty
}

var parserPool = sync.Pool{
	New: func() any {
		return &parser{
			exprBuf:    make([]ast.Expression, 0, 64),
			stmtBuf:    make([]ast.Statement, 0, 64),
			propBuf:    make([]ast.Property, 0, 16),
			elemBuf:    make([]ast.ClassElement, 0, 16),
			declBuf:    make([]ast.VariableDeclarator, 0, 16),
			patBuf:     make([]ast.Pattern, 0, 16),
			patPropBuf: make([]ast.PatternProperty, 0, 16),
		}
	},
}

func getParser(src string) *parser {
	p := parserPool.Get().(*parser)
	p.str = src
	p.alloc = newNodeAllocator()
	p.scanner = scanner.NewScanner(src, &p.errors)
	return p
}

func putParser(p *parser) {
	p.str = ""
	p.alloc = nodeAllocator{}
	p.options = parseOptions{}
	p.scanner = scanner.Scanner{}
	p.scope = nil
	p.errors = nil
	p.recover.idx = 0
	p.recover.count = 0
	p.exprBuf = p.exprBuf[:0]
	p.stmtBuf = p.stmtBuf[:0]
	p.propBuf = p.propBuf[:0]
	p.elemBuf = p.elemBuf[:0]
	p.declBuf = p.declBuf[:0]
	p.patBuf = p.patBuf[:0]
	p.patPropBuf = p.patPropBuf[:0]
	parserPool.Put(p)
}

// Parse parses src as an ECMAScript script and returns the program AST.
// Errors are accumulated; on a non-nil error the returned [*ast.Program]
// may still be partially populated. Behavior can be tuned with [Option]
// values such as [PreserveParens]; with none, parsing uses its defaults.
//
// To recover byte positions from errors use [errors.As] against
// [*Error] or [scanner.Error], or the shared [ast.Positioned] interface.
func Parse(src string, opts ...Option) (*ast.Program, error) {
	p := getParser(src)
	for _, opt := range opts {
		opt(&p.options)
	}
	program, err := p.parse()
	putParser(p)
	return program, err
}

// ParseBytes is identical to [Parse] but accepts a byte slice. The slice's
// contents must remain valid for the lifetime of the returned AST: the
// parser stores no-copy references into src for identifier and literal
// names.
func ParseBytes(src []byte, opts ...Option) (*ast.Program, error) {
	return Parse(unsafe.String(unsafe.SliceData(src), len(src)), opts...)
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

func (p *parser) requireSemicolon() {
	if !p.semicolon() {
		p.errorUnexpectedToken(p.currentKind())
	}
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

func (p *parser) finishPropBuf(mark int) ast.Properties {
	result := p.alloc.CopyProperties(p.propBuf[mark:])
	p.propBuf = p.propBuf[:mark]
	return result
}

func (p *parser) finishPatBuf(mark int) ast.Patterns {
	result := p.alloc.CopyPatterns(p.patBuf[mark:])
	p.patBuf = p.patBuf[:mark]
	return result
}

func (p *parser) finishPatPropBuf(mark int) ast.PatternProperties {
	result := p.alloc.CopyPatternProperties(p.patPropBuf[mark:])
	p.patPropBuf = p.patPropBuf[:mark]
	return result
}

func (p *parser) finishDeclBuf(mark int) ast.VariableDeclarators {
	result := p.alloc.CopyDeclarators(p.declBuf[mark:])
	p.declBuf = p.declBuf[:mark]
	return result
}

func (p *parser) finishElemBuf(mark int) ast.ClassElements {
	result := p.alloc.CopyClassElements(p.elemBuf[mark:])
	p.elemBuf = p.elemBuf[:mark]
	return result
}

// expect consumes the current token, which must be value, and returns its
// start offset. Node position fields populated from expect (brackets, braces,
// parentheses, and keyword indices) are token starts; the Idx1 methods add the
// token/keyword length to obtain the exclusive end.
func (p *parser) expect(value token.Token) ast.Idx {
	idx := p.scanner.Token.Idx0
	if p.scanner.Token.Kind != value {
		p.errorUnexpectedToken(p.scanner.Token.Kind)
	}
	p.next()
	return idx
}
