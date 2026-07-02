package resolver

import "github.com/t14raptor/go-fast/ast"

type resolver struct {
	ast.NoopVisitor

	current *scope

	identType identType
	declKind  declKind

	nextCtxt ast.ScopeContext

	binder resolverBinder
}

func Resolve(p ast.VisitableNode) {
	r := &resolver{
		identType: identTypeRef,
		nextCtxt:  ast.TopLevelContext,
	}
	r.V = r
	r.binder.r = r
	r.binder.V = &r.binder

	p.VisitWith(r)
}

func (r *resolver) pushScope(kind scopeKind) {
	ctx := r.nextCtxt
	r.nextCtxt++

	r.current = &scope{
		parent:          r.current,
		kind:            kind,
		declaredSymbols: make(map[string]declKind),
		ctx:             ctx,
	}
}

func (r *resolver) popScope() {
	if r.current.parent != nil {
		r.current = r.current.parent
	}
}

func (r *resolver) modify(id *ast.Identifier, kind declKind) {
	if id.ScopeContext != ast.UnresolvedContext {
		return
	}

	r.current.declaredSymbols[id.Name] = kind

	id.ScopeContext = r.current.ctx
}

func (r *resolver) lookupContext(sym string) (ast.ScopeContext, *scope) {
	for s := r.current; s != nil; s = s.parent {
		if _, exists := s.declaredSymbols[sym]; exists {
			return s.ctx, s
		}
	}
	return ast.UnresolvedContext, nil
}

func (r *resolver) VisitArrowFunctionLiteral(n *ast.ArrowFunctionLiteral) {
	r.pushScope(scopeKindFunction)

	n.ScopeContext = r.current.ctx

	oldIdentType := r.identType
	r.identType = identTypeBinding
	n.ParameterList.VisitWith(r)

	r.identType = identTypeRef
	switch n.Body.Kind() {
	case ast.ConciseBodyBlock:
		body := n.Body.MustBlock()
		body.ScopeContext = r.current.ctx
		// Prevent creating a new scope.
		body.VisitChildrenWith(r)
	case ast.ConciseBodyExpr:
		n.Body.MustExpr().VisitWith(r)
	}
	r.identType = oldIdentType

	r.popScope()
}

func (r *resolver) VisitBlockStatement(n *ast.BlockStatement) {
	r.pushScope(scopeKindBlock)
	n.ScopeContext = r.current.ctx
	n.VisitChildrenWith(r)
	r.popScope()
}

func (r *resolver) VisitForOfStatement(n *ast.ForOfStatement) {
	r.pushScope(scopeKindBlock)

	oldIdentType := r.identType
	r.identType = identTypeRef

	n.Into.VisitWith(r)
	n.Source.VisitWith(r)

	if block, ok := n.Body.Block(); ok {
		block.ScopeContext = r.current.ctx
	}
	n.Body.VisitWith(r)

	r.identType = oldIdentType
	r.popScope()
}

func (r *resolver) VisitForInStatement(n *ast.ForInStatement) {
	r.pushScope(scopeKindBlock)

	oldIdentType := r.identType
	r.identType = identTypeRef

	n.Into.VisitWith(r)
	n.Source.VisitWith(r)

	if block, ok := n.Body.Block(); ok {
		block.ScopeContext = r.current.ctx
	}
	n.Body.VisitWith(r)

	r.identType = oldIdentType
	r.popScope()
}

func (r *resolver) VisitForStatement(n *ast.ForStatement) {
	r.pushScope(scopeKindBlock)

	oldIdentType := r.identType
	r.identType = identTypeBinding

	// Handle initializer
	if n.Initializer != nil {
		n.Initializer.VisitWith(r)
	}

	// Handle test expression
	r.identType = identTypeRef
	if n.Test != nil {
		n.Test.VisitWith(r)
	}

	// Handle update expression
	if n.Update != nil {
		n.Update.VisitWith(r)
	}

	// Handle body
	r.identType = oldIdentType
	n.Body.VisitWith(r)

	r.popScope()
}

func (r *resolver) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	r.pushScope(scopeKindFunction)
	if n.Name != nil {
		r.modify(n.Name, declKindFunction)
	}

	n.ScopeContext = r.current.ctx

	oldIdentType := r.identType
	r.identType = identTypeBinding
	n.ParameterList.VisitWith(r)

	r.identType = identTypeRef
	// Prevent creating new scope.
	n.Body.ScopeContext = r.current.ctx
	n.Body.VisitChildrenWith(r)

	r.identType = oldIdentType

	r.popScope()
}

func (r *resolver) VisitClassDeclaration(n *ast.ClassDeclaration) {
	if n.Class.Name != nil {
		r.modify(n.Class.Name, declKindClass)
	}
	n.Class.VisitWith(r)
}

func (r *resolver) VisitClassLiteral(n *ast.ClassLiteral) {
	needsInnerNameScope := n.Name != nil && n.Name.ScopeContext == ast.UnresolvedContext
	if needsInnerNameScope {
		r.pushScope(scopeKindBlock)
		r.modify(n.Name, declKindClass)
	}

	if n.SuperClass != nil {
		n.SuperClass.VisitWith(r)
	}
	n.Body.VisitWith(r)

	if needsInnerNameScope {
		r.popScope()
	}
}

func (r *resolver) VisitParameterList(n *ast.ParameterList) {
	// Phase 1: pre-declare every parameter binding so a default can forward-
	// reference a later parameter (e.g. function f(a = b, b) {}). The binder
	// declares binding identifiers and skips defaults/computed keys.
	n.VisitChildrenWith(&r.binder)

	// Phase 2: resolve defaults and computed keys as references. The binding
	// identifiers already have a scope context, so VisitIdentifier skips them.
	old := r.identType
	r.identType = identTypeRef
	n.VisitChildrenWith(r)
	r.identType = old
}

func (r *resolver) VisitCatchClause(n *ast.CatchClause) {
	r.pushScope(scopeKindBlock)
	if n.Parameter != nil {
		old := r.identType
		r.identType = identTypeBinding
		n.Parameter.VisitWith(r)
		r.identType = old
	}
	n.Body.ScopeContext = r.current.ctx
	n.Body.VisitChildrenWith(r)
	r.popScope()
}

func (r *resolver) VisitProgram(n *ast.Program) {
	r.pushScope(scopeKindBlock)
	n.VisitChildrenWith(r)
	r.popScope()
}

func (r *resolver) VisitStatements(n *ast.Statements) {
	// Handle hoisting
	h := newHoister(r)
	h.V = h
	n.VisitWith(h)

	// Resolve
	n.VisitChildrenWith(r)
}

func (r *resolver) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	oldDeclKind := r.declKind
	r.declKind = declKindVar

	for _, decl := range n.List {
		old := r.identType
		r.identType = identTypeBinding
		decl.Target.VisitWith(r)
		r.identType = old

		if decl.Initializer != nil {
			decl.Initializer.VisitWith(r)
		}
	}

	r.declKind = oldDeclKind
}

func (r *resolver) VisitExpression(expr *ast.Expression) {
	oldIdentType := r.identType
	r.identType = identTypeRef
	expr.VisitChildrenWith(r)
	r.identType = oldIdentType
}

func (r *resolver) VisitIdentifier(n *ast.Identifier) {
	if n == nil || n.ScopeContext != ast.UnresolvedContext {
		return
	}

	switch r.identType {
	case identTypeBinding:
		r.modify(n, r.declKind)
	case identTypeRef:
		if mark, _ := r.lookupContext(n.Name); mark != ast.UnresolvedContext {
			n.ScopeContext = mark
		} else {
			r.modify(n, r.declKind)
		}
	}
}

func (r *resolver) VisitMemberProperty(n *ast.MemberProperty) {
	if computed, ok := n.Computed(); ok {
		computed.VisitWith(r)
	}
}
