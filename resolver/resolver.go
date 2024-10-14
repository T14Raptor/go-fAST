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

type VarInfo struct {
	DeclKind   DeclKind
	Identifier *ast.Identifier
	Value      *ast.Expression
}

type Scope struct {
	parent *Scope

	kind ScopeKind

	mark ast.ScopeContext

	declaredSymbols map[string]DeclKind
	vars            map[string]*VarInfo
}

func (s *Scope) findVarInfo(id string) *VarInfo {
	if v, ok := s.vars[id]; ok {
		return v
	}
	if s.parent != nil {
		return s.parent.findVarInfo(id)
	}
	return nil
}

type Resolver struct {
	ast.NoopVisitor

	current *Scope

	identType IdentType
	declKind  DeclKind

	nextCtxt ast.ScopeContext
}

const (
	TopLevelMark   ast.ScopeContext = 1
	UnresolvedMark ast.ScopeContext = 0
)

func NewResolver() *Resolver {
	r := &Resolver{
		identType: IdentTypeRef,
		nextCtxt:  TopLevelMark,
	}
	r.V = r

	return r
}

func Resolve(p *ast.Program) *Resolver {
	resolver := NewResolver()
	p.VisitWith(resolver)
	return resolver
}

func (r *Resolver) newCtxt() ast.ScopeContext {
	ctxt := r.nextCtxt
	r.nextCtxt++
	return ctxt
}

func (r *Resolver) pushScope(kind ScopeKind) {
	r.current = &Scope{
		parent:          r.current,
		kind:            kind,
		declaredSymbols: make(map[string]DeclKind),
		mark:            r.newCtxt(),
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

	id.ScopeContext = r.current.mark
}

func (r *Resolver) lookupContext(sym string) (ast.ScopeContext, *Scope) {
	for scope := r.current; scope != nil; scope = scope.parent {
		if _, exists := scope.declaredSymbols[sym]; exists {
			return scope.mark, scope
		}
	}
	return UnresolvedMark, nil
}

func (r *Resolver) VisitBlockStatement(n *ast.BlockStatement) {
	r.pushScope(ScopeKindBlock)
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

	//if blockStmt, ok := n.Body.Stmt.(*ast.BlockStatement); ok {
	//	//r.markBlock(&blockStmt.ScopeContext)
	//}
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

func (r *Resolver) VisitFunctionDeclaration(n *ast.FunctionDeclaration) {
	r.modify(n.Function.Name, DeclKindFunction)

	n.Function.VisitWith(r)
}

func (r *Resolver) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	r.pushScope(ScopeKindFunction)

	oldIdentType := r.identType
	r.identType = IdentTypeBinding
	n.ParameterList.VisitWith(r)

	if rest, ok := n.ParameterList.Rest.(*ast.Identifier); ok {
		rest.VisitWith(r)
	} else if rest != nil {
		panic(fmt.Sprintf("Unexpected rest type: %T\n", n.ParameterList.Rest))
	}

	r.identType = oldIdentType

	n.Body.VisitChildrenWith(r)

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
			decl.Initializer.VisitChildrenWith(r)
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
