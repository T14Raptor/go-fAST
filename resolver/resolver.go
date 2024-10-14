package resolver

import (
	"fmt"
	"github.com/t14raptor/go-fast/ast"
)

type IdentType int

const (
	IdentTypeRef      IdentType = iota // Reference (read)
	IdentTypeBinding                   // Binding (declaration)
	IdentTypeRefWrite                  // Reference (write)
)

type DeclKind int

const (
	DeclKindVar DeclKind = iota
	DeclKindFunction
	DeclKindClass
)

type ScopeKind int

const (
	ScopeKindBlock ScopeKind = iota
	ScopeKindFunction
	ScopeKindForStatement
)

type VarInfo struct {
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
		nextCtxt:  TopLevelMark + 1,
	}
	r.V = r

	r.pushScope(ScopeKindBlock) // Global scope
	return r
}

func (r *Resolver) VisitBlockStatement(n *ast.BlockStatement) {
	r.pushScope(ScopeKindBlock)
	r.hoistScope(n.List)
	n.VisitChildrenWith(r)
	r.popScope()
}

func (r *Resolver) VisitForOfStatement(n *ast.ForOfStatement) {
	r.pushScope(ScopeKindBlock) // Using Block scope for ForOfStatement

	oldIdentType := r.identType

	// Handle the 'Into' part (left-hand side of for...of)
	if n.Into != nil {
		r.identType = IdentTypeBinding
		switch into := n.Into.Into.(type) {
		case *ast.VariableDeclaration:
			// Handle ForIntoVar (likely a simple variable)
			for _, x := range into.List {
				r.VisitBindingTarget(x.Target)
			}

		//case *ast.ForDeclaration:
		// Handle ForDeclaration (likely a 'let' or 'const' declaration)

		//for _, decl := range into.Target. {
		//	r.VisitVariableDeclarator(decl)
		//}
		default:
			// Log unexpected type or handle error
			fmt.Printf("Unexpected ForInto type: %T\n", into)
		}
	}

	// Handle the 'Source' part (right-hand side of for...of)

	if n.Source != nil {
		r.VisitExpression(n.Source)
	}

	// Handle body
	r.identType = oldIdentType
	if n.Body != nil {
		//if blockStmt, ok := n.Body.Stmt.(*ast.BlockStatement); ok {
		//	//r.markBlock(&blockStmt.ScopeContext)
		//}
		n.Body.VisitWith(r)
	}
	r.identType = IdentTypeRef
	r.popScope()
}

//func (r *Resolver) VisitCallExpression(n *ast.CallExpression) {
//	for _, x := range
//}

func (r *Resolver) VisitForStatement(n *ast.ForStatement) {
	r.pushScope(ScopeKindBlock) // Using Block scope as ForStatement is not defined

	oldIdentType := r.identType

	// Handle initializer
	if n.Initializer != nil {
		r.identType = IdentTypeBinding
		if s, ok := n.Initializer.Initializer.(*ast.VariableDeclaration); ok {
			for _, x := range s.List {
				r.VisitBindingTarget(x.Target)
				r.identType = IdentTypeRef
				r.VisitExpression(x.Initializer)
			}
		} else if expr, ok := n.Initializer.Initializer.(*ast.Expression); ok {
			r.identType = IdentTypeRef
			r.VisitExpression(expr)
		}
	}

	// Handle test expression
	r.identType = IdentTypeRef
	if n.Test != nil {
		r.VisitExpression(n.Test)
	}

	// Handle update expression
	if n.Update != nil {
		r.VisitExpression(n.Update)
	}

	// Handle body
	r.identType = oldIdentType

	n.Body.VisitWith(r)

	r.popScope()
}

func (r *Resolver) VisitFunctionDeclaration(n *ast.FunctionDeclaration) {
	r.modify(n.Function.Name, DeclKindFunction)
	r.VisitFunctionLiteral(n.Function)
}

func (r *Resolver) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	r.pushScope(ScopeKindFunction)

	oldIdentType := r.identType
	r.identType = IdentTypeBinding
	for _, param := range n.ParameterList.List {
		r.VisitBindingTarget(param.Target)
	}

	if n.ParameterList.Rest != nil {
		switch rest := n.ParameterList.Rest.(type) {
		case *ast.Identifier:
			r.VisitIdentifier(rest)
		default:
			panic(fmt.Sprintf("Unexpected rest type: %T\n", n.ParameterList.Rest))
		}
	}

	r.identType = oldIdentType

	r.hoistScope(n.Body.List)
	n.Body.VisitChildrenWith(r)

	r.popScope()
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

func (r *Resolver) markForRef(sym string) (ast.ScopeContext, *Scope) {
	for scope := r.current; scope != nil; scope = scope.parent {
		if _, exists := scope.declaredSymbols[sym]; exists {
			return scope.mark, scope
		}
	}
	return UnresolvedMark, nil
}

func (r *Resolver) VisitProgram(n *ast.Program) {
	r.hoistScope(n.Body)
	n.VisitChildrenWith(r)
}

func (r *Resolver) hoistStatement(stmt ast.Statement, hoister *Hoister) {
	switch s := stmt.Stmt.(type) {
	case *ast.VariableDeclaration:
		hoister.VisitVariableStatement(s)
	case *ast.FunctionDeclaration:
		hoister.VisitFunctionDeclaration(s)
	case *ast.BlockStatement:
		oldInBlock := hoister.inBlock
		hoister.inBlock = true
		for _, blockStmt := range s.List {
			r.hoistStatement(blockStmt, hoister)
		}
		hoister.inBlock = oldInBlock
	}
}

func (r *Resolver) hoistScope(stmts []ast.Statement) {
	hoister := NewHoister(r)
	for _, stmt := range stmts {
		r.hoistStatement(stmt, hoister)
	}
}

func (r *Resolver) VisitVariableStatement(n *ast.VariableDeclaration) {
	oldDeclKind := r.declKind
	r.declKind = DeclKindVar

	for _, decl := range n.List {
		r.VisitVariableDeclarator(decl)
	}

	r.declKind = oldDeclKind
}

func (r *Resolver) VisitBindingTarget(target *ast.BindingTarget) {
	switch t := target.Target.(type) {
	case *ast.Identifier:
		r.VisitIdentifier(t)
	case *ast.ArrayPattern:
		for _, element := range t.Elements {
			if element.Expr != nil {
				r.VisitBindingTarget(element.Expr.(*ast.BindingTarget))
			}
		}
	case *ast.ObjectPattern:
	//fmt.Println(1)
	//for _, prop := range t.Properties {
	//	a := (prop).(*ast.Property)
	//	r.VisitBindingTarget(*prop.(*ast.Property)))
	//}
	default:
		fmt.Println(1)
	}
}

func (r *Resolver) VisitExpression(expr *ast.Expression) {
	if expr == nil || expr.Expr == nil {
		return
	}
	switch e := expr.Expr.(type) {
	case *ast.Identifier:
		r.VisitIdentifier(e)
	case *ast.AssignExpression:
		r.VisitAssignExpression(e)
	default:
		expr.VisitChildrenWith(r)
	}
}

func (r *Resolver) VisitAssignmentExpression(n *ast.AssignExpression) {
	oldIdentType := r.identType

	// lhs is a write reference
	r.identType = IdentTypeRefWrite
	r.VisitExpression(n.Left)

	// rhs is a read reference, so we need a reference type
	r.identType = IdentTypeRef
	r.VisitExpression(n.Right)

	r.identType = oldIdentType
}

func (r *Resolver) VisitIdentifier(n *ast.Identifier) {
	if n == nil {
		return
	}

	switch r.identType {
	case IdentTypeBinding:
		r.modify(n, r.declKind)
	case IdentTypeRef, IdentTypeRefWrite:
		if n.ScopeContext != UnresolvedMark {
			return
		}

		if mark, _ := r.markForRef(n.Name); mark != UnresolvedMark {
			n.ScopeContext = mark
		} else {
			n.ScopeContext = UnresolvedMark
		}

		if scope := r.findInOuterScopes(n.Name); scope != nil {
			n.ScopeContext = scope.mark
		} else {
			n.ScopeContext = UnresolvedMark
		}
	}
}

func (r *Resolver) findInOuterScopes(name string) *Scope {
	for scope := r.current; scope != nil; scope = scope.parent {
		if _, exists := scope.declaredSymbols[name]; exists {
			return scope
		}
	}
	return nil
}

/* Handle classes */
func (r *Resolver) VisitFieldDefinition(n *ast.FieldDefinition) {
	// Handle field definition
	if n.Key != nil {
		(*n.Key).VisitWith(r)
	}
	if n.Initializer != nil {
		r.VisitExpression(n.Initializer)
	}
}

func (r *Resolver) VisitMethodDefinition(n *ast.MethodDefinition) {
	// Handle method definition
	if n.Key != nil {
		(*n.Key).VisitWith(r)
	}
	r.VisitFunctionLiteral(n.Body)
}

func (r *Resolver) VisitClassStaticBlock(n *ast.ClassStaticBlock) {
	// Handle static block
	r.VisitBlockStatement(n.Block)
}

func Resolve(p *ast.Program) *Resolver {
	resolver := NewResolver()
	p.VisitWith(resolver)
	return resolver
}
