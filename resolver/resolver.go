package resolver

import (
	"fmt"

	"github.com/t14raptor/go-fast/ast"
)

type IdentType int

const (
	IdentTypeRef     IdentType = iota // Reference (read)
	IdentTypeBinding                  // Binding (declaration)
)

type DeclKind int

const (
	DeclKindVar DeclKind = iota
	DeclKindFunction
)

type ScopeKind int

const (
	ScopeKindBlock ScopeKind = iota
	ScopeKindFunction
)

type Scope struct {
	parent *Scope

	kind ScopeKind

	ctx ast.ScopeContext

	declaredSymbols map[string]DeclKind
}

func (s *Scope) isDeclared(id string) (DeclKind, bool) {
	for scope := s; scope != nil; scope = scope.parent {
		if declKind, exists := scope.declaredSymbols[id]; exists {
			return declKind, true
		}
	}
	return 0, false
}

type Resolver struct {
	ast.NoopVisitor

	current *Scope

	identType IdentType
	declKind  DeclKind

	nextCtxt ast.ScopeContext
}

const (
	UnresolvedMark ast.ScopeContext = 0
	TopLevelMark   ast.ScopeContext = 1
)

func Resolve(p *ast.Program) *Resolver {
	r := &Resolver{
		identType: IdentTypeRef,
		nextCtxt:  TopLevelMark,
	}
	r.V = r

	p.VisitWith(r)
	return r
}

func (r *Resolver) pushScope(kind ScopeKind) {
	ctx := r.nextCtxt
	r.nextCtxt++

	r.current = &Scope{
		parent:          r.current,
		kind:            kind,
		declaredSymbols: make(map[string]DeclKind),
		ctx:             ctx,
	}
}

func (r *Resolver) popScope() {
	if r.current.parent != nil {
		r.current = r.current.parent
	}
}

func (r *Resolver) modify(id *ast.Identifier, kind DeclKind) {
	if id.ScopeContext != UnresolvedMark {
		return
	}

	r.current.declaredSymbols[id.Name] = kind

	id.ScopeContext = r.current.ctx
}

func (r *Resolver) lookupContext(sym string) (ast.ScopeContext, *Scope) {
	for scope := r.current; scope != nil; scope = scope.parent {
		if _, exists := scope.declaredSymbols[sym]; exists {
			return scope.ctx, scope
		}
	}
	return UnresolvedMark, nil
}

func (r *Resolver) VisitArrowFunctionLiteral(n *ast.ArrowFunctionLiteral) {
	r.pushScope(ScopeKindFunction)

	n.ScopeContext = r.current.ctx

	oldIdentType := r.identType
	r.identType = IdentTypeBinding
	n.ParameterList.VisitWith(r)

	r.identType = IdentTypeRef
	switch body := n.Body.Body.(type) {
	case *ast.BlockStatement:
		body.ScopeContext = r.current.ctx
		// Prevent creating a new scope.
		body.VisitChildrenWith(r)
	case *ast.Expression:
		body.VisitWith(r)
	}
	r.identType = oldIdentType

	r.popScope()
}

func (r *Resolver) VisitBlockStatement(n *ast.BlockStatement) {
	r.pushScope(ScopeKindBlock)
	n.ScopeContext = r.current.ctx
	n.VisitChildrenWith(r)
	r.popScope()
}

func (r *Resolver) VisitForOfStatement(n *ast.ForOfStatement) {
	r.pushScope(ScopeKindBlock) // Using Block scope for ForOfStatement

	oldIdentType := r.identType
	r.identType = IdentTypeRef

	// Handle the 'Into' part (left-hand side of for...of)
	n.Into.VisitWith(r)

	// Handle the 'Source' part (right-hand side of for...of)
	n.Source.VisitWith(r)

	if blockStmt, ok := n.Body.Stmt.(*ast.BlockStatement); ok {
		blockStmt.ScopeContext = r.current.ctx
	}
	n.Body.VisitWith(r)

	r.identType = oldIdentType
	r.popScope()
}

func (r *Resolver) VisitForInStatement(n *ast.ForInStatement) {
	r.pushScope(ScopeKindBlock) // Using Block scope for ForOfStatement

	oldIdentType := r.identType
	r.identType = IdentTypeRef

	n.Into.VisitWith(r)
	n.Source.VisitWith(r)

	if blockStmt, ok := n.Body.Stmt.(*ast.BlockStatement); ok {
		blockStmt.ScopeContext = r.current.ctx
	}
	n.Body.VisitWith(r)

	r.identType = oldIdentType
	r.popScope()
}

func (r *Resolver) VisitForStatement(n *ast.ForStatement) {
	r.pushScope(ScopeKindBlock) // Using Block scope as ForStatement is not defined

	oldIdentType := r.identType
	r.identType = IdentTypeBinding

	// Handle initializer
	if n.Initializer != nil {
		n.Initializer.VisitWith(r)
	}

	// Handle test expression
	r.identType = IdentTypeRef
	n.Test.VisitWith(r)

	// Handle update expression
	n.Update.VisitWith(r)

	// Handle body
	r.identType = oldIdentType
	n.Body.VisitWith(r)

	r.popScope()
}

func (r *Resolver) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	if n.Name != nil {
		r.modify(n.Name, DeclKindFunction)
	}

	r.pushScope(ScopeKindFunction)

	n.ScopeContext = r.current.ctx

	oldIdentType := r.identType
	r.identType = IdentTypeBinding
	n.ParameterList.VisitWith(r)

	if rest, ok := n.ParameterList.Rest.(*ast.Identifier); ok {
		rest.VisitWith(r)
	} else if n.ParameterList.Rest != nil {
		panic(fmt.Sprintf("Unexpected rest type: %T\n", n.ParameterList.Rest))
	}

	r.identType = IdentTypeRef
	// Prevent creating new scope.
	n.Body.ScopeContext = r.current.ctx
	n.Body.VisitChildrenWith(r)

	r.identType = oldIdentType

	r.popScope()
}

func (r *Resolver) VisitProgram(n *ast.Program) {
	r.pushScope(ScopeKindBlock)
	n.VisitChildrenWith(r)
	r.popScope()
}

func (r *Resolver) VisitStatements(n *ast.Statements) {
	// Handle hoisting
	h := NewHoister(r)
	h.V = h
	n.VisitWith(h)

	// Resolve
	n.VisitChildrenWith(r)
}

func (r *Resolver) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	oldDeclKind := r.declKind
	r.declKind = DeclKindVar

	for _, decl := range n.List {
		oldIdentType := r.identType
		r.identType = IdentTypeBinding
		decl.Target.VisitWith(r)
		r.identType = oldIdentType

		if decl.Initializer != nil {
			decl.Initializer.VisitWith(r)
		}
	}

	r.declKind = oldDeclKind
}

func (r *Resolver) VisitExpression(expr *ast.Expression) {
	if expr == nil || expr.Expr == nil {
		return
	}

	oldIdentType := r.identType
	r.identType = IdentTypeRef
	expr.VisitChildrenWith(r)
	r.identType = oldIdentType
}

func (r *Resolver) VisitIdentifier(n *ast.Identifier) {
	if n == nil || n.ScopeContext != UnresolvedMark {
		return
	}

	switch r.identType {
	case IdentTypeBinding:
		r.modify(n, r.declKind)
	case IdentTypeRef:
		if mark, _ := r.lookupContext(n.Name); mark != UnresolvedMark {
			n.ScopeContext = mark
		} else {
			r.modify(n, r.declKind)
		}
	}
}
