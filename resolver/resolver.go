package resolver

import "github.com/t14raptor/go-fast/ast"

type resolver struct {
	ast.NoopVisitor

	current *scope

	identType identType
	declKind  declKind

	nextCtxt ast.ScopeContext

	binder resolverBinder

	// hoister is shared across statement lists; VisitStatements resets it per run.
	hoister hoister

	// scopePool recycles popped scopes.
	scopePool []*scope

	// symbols holds each name's innermost visible binding; trail is the undo log
	// popScope rewinds to restore shadowed bindings.
	symbols map[string]symEntry
	trail   []trailEntry
}

func Resolve(p ast.VisitableNode) {
	r := &resolver{
		identType: identTypeRef,
		nextCtxt:  ast.TopLevelContext,
		symbols:   make(map[string]symEntry),
	}
	r.V = r
	r.binder.r = r
	r.binder.V = &r.binder

	r.hoister.resolver = r
	r.hoister.kind = declKindVar
	r.hoister.V = &r.hoister
	r.hoister.binder.h = &r.hoister
	r.hoister.binder.V = &r.hoister.binder

	p.VisitWith(r)
}

func (r *resolver) pushScope() {
	ctx := r.nextCtxt
	r.nextCtxt++

	var s *scope
	if n := len(r.scopePool); n > 0 {
		s = r.scopePool[n-1]
		r.scopePool = r.scopePool[:n-1]
	} else {
		s = new(scope)
	}
	s.parent = r.current
	s.ctx = ctx
	s.trailBase = len(r.trail)
	r.current = s
}

func (r *resolver) popScope() {
	s := r.current
	if s.parent == nil {
		return
	}

	// Restore the bindings this scope shadowed.
	for i := len(r.trail) - 1; i >= s.trailBase; i-- {
		u := &r.trail[i]
		if u.had {
			r.symbols[u.name] = u.prev
		} else {
			delete(r.symbols, u.name)
		}
	}
	r.trail = r.trail[:s.trailBase]

	r.current = s.parent
	s.parent = nil
	r.scopePool = append(r.scopePool, s)
}

// declare binds name in the current scope (last writer wins) and returns its
// context.
func (r *resolver) declare(name string, kind declKind) ast.ScopeContext {
	ctx := r.current.ctx
	prev, had := r.symbols[name]
	if had && prev.ctx == ctx {
		// Redeclaration in the same scope: only the kind can change.
		r.symbols[name] = symEntry{ctx: ctx, kind: kind}
		return ctx
	}
	// The root scope never pops, so its bindings need no undo record.
	if r.current.parent != nil {
		r.trail = append(r.trail, trailEntry{name: name, prev: prev, had: had})
	}
	r.symbols[name] = symEntry{ctx: ctx, kind: kind}
	return ctx
}

// lookup returns the kind and context of name's innermost visible binding.
func (r *resolver) lookup(name string) (declKind, ast.ScopeContext, bool) {
	e, ok := r.symbols[name]
	return e.kind, e.ctx, ok
}

func (r *resolver) modify(id *ast.Identifier, kind declKind) {
	if id.ScopeContext != ast.UnresolvedContext {
		return
	}
	id.ScopeContext = r.declare(id.Name, kind)
}

func (r *resolver) lookupContext(sym string) ast.ScopeContext {
	_, ctx, _ := r.lookup(sym)
	return ctx
}

func (r *resolver) VisitArrowFunctionLiteral(n *ast.ArrowFunctionLiteral) {
	r.pushScope()

	n.ScopeContext = r.current.ctx

	oldIdentType := r.identType
	r.identType = identTypeBinding
	n.ParameterList.VisitWith(r)

	r.identType = identTypeRef
	switch n.Body.Kind() {
	case ast.ConciseBodyBlock:
		body := n.Body.MustBlock()
		body.ScopeContext = r.current.ctx
		// Reuse the arrow scope for the body; don't open a new one.
		body.VisitChildrenWith(r)
	case ast.ConciseBodyExpr:
		n.Body.MustExpr().VisitWith(r)
	}
	r.identType = oldIdentType

	r.popScope()
}

func (r *resolver) VisitBlockStatement(n *ast.BlockStatement) {
	r.pushScope()
	n.ScopeContext = r.current.ctx
	n.VisitChildrenWith(r)
	r.popScope()
}

func (r *resolver) VisitForInStatement(n *ast.ForInStatement) {
	r.visitForInOf(n.Into, n.Source, n.Body)
}

func (r *resolver) VisitForOfStatement(n *ast.ForOfStatement) {
	r.visitForInOf(n.Into, n.Source, n.Body)
}

func (r *resolver) visitForInOf(into *ast.ForInto, source *ast.Expression, body *ast.Statement) {
	r.pushScope()

	old := r.identType
	r.identType = identTypeRef
	into.VisitWith(r)
	source.VisitWith(r)
	if block, ok := body.Block(); ok {
		block.ScopeContext = r.current.ctx
	}
	body.VisitWith(r)
	r.identType = old

	r.popScope()
}

func (r *resolver) VisitForStatement(n *ast.ForStatement) {
	r.pushScope()

	oldIdentType := r.identType
	r.identType = identTypeBinding
	if n.Initializer != nil {
		n.Initializer.VisitWith(r)
	}

	r.identType = identTypeRef
	if n.Test != nil {
		n.Test.VisitWith(r)
	}
	if n.Update != nil {
		n.Update.VisitWith(r)
	}

	r.identType = oldIdentType
	n.Body.VisitWith(r)

	r.popScope()
}

func (r *resolver) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	r.pushScope()
	if n.Name != nil {
		r.modify(n.Name, declKindFunction)
	}

	n.ScopeContext = r.current.ctx

	oldIdentType := r.identType
	r.identType = identTypeBinding
	n.ParameterList.VisitWith(r)

	r.identType = identTypeRef
	// Reuse the function scope for the body; don't open a new one.
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
		r.pushScope()
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
	// Pre-declare all parameter bindings first, so a default can forward-reference
	// a later parameter (e.g. function f(a = b, b) {}); the binder skips defaults
	// and computed keys.
	n.VisitChildrenWith(&r.binder)

	// Then resolve those defaults and computed keys as references; the bindings
	// already have a context, so VisitIdentifier leaves them untouched.
	old := r.identType
	r.identType = identTypeRef
	n.VisitChildrenWith(r)
	r.identType = old
}

func (r *resolver) VisitCatchClause(n *ast.CatchClause) {
	r.pushScope()
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
	r.pushScope()
	n.VisitChildrenWith(r)
	r.popScope()
}

func (r *resolver) VisitStatements(n *ast.Statements) {
	// Hoisting pre-pass. The hoister is shared, so reset it first; this is safe
	// because a hoister run never re-enters the resolver, so runs never overlap.
	h := &r.hoister
	h.kind = declKindVar
	h.inBlock = false
	h.inCatchBody = false
	h.excludedFromCatch = nil
	clear(h.catchParamDecls)
	n.VisitWith(h)

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
	if n.ScopeContext != ast.UnresolvedContext {
		return
	}

	switch r.identType {
	case identTypeBinding:
		r.modify(n, r.declKind)
	case identTypeRef:
		if mark := r.lookupContext(n.Name); mark != ast.UnresolvedContext {
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
