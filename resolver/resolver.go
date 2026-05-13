package resolver

import (
	"reflect"

	"github.com/t14raptor/go-fast/ast"
)

type IdentType int

const (
	IdentTypeRef IdentType = iota
	IdentTypeBinding
)

const (
	UnresolvedMark ast.ScopeContext = 0
	TopLevelMark   ast.ScopeContext = 1
)

type Resolver struct {
	ast.NoopVisitor

	current *Scope

	identType IdentType
	nextCtxt  ast.ScopeContext

	// chainView maps a name to its innermost ScopeContext in the active
	// scope chain. Maintained via per-scope shadow entries so lookupContext
	// is one map probe instead of a parent walk.
	chainView map[string]ast.ScopeContext
}

// Resolve assigns a ScopeContext to every identifier in p so that two
// identifiers share a context iff they bind to the same declaration.
func Resolve(p ast.VisitableNode) *Resolver {
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
	r.current = getScope(r.current, kind, ctx)
}

func (r *Resolver) popScope() {
	if r.current == nil || r.current.parent == nil {
		return
	}
	cur := r.current
	for i := len(cur.shadows) - 1; i >= 0; i-- {
		sh := cur.shadows[i]
		if sh.hadPrev {
			r.chainView[sh.name] = sh.prev
		} else {
			delete(r.chainView, sh.name)
		}
	}
	r.current = cur.parent
	putScope(cur)
}

// modify declares id in the current scope and stamps it with the
// scope's ScopeContext. A no-op if id already has a context.
func (r *Resolver) modify(id *ast.Identifier, kind DeclKind) {
	if id.ScopeContext != UnresolvedMark {
		return
	}

	cur := r.current
	if cur.declaredSymbols == nil {
		cur.declaredSymbols = make(map[string]DeclKind, 4)
	}
	if _, already := cur.declaredSymbols[id.Name]; !already {
		if r.chainView == nil {
			r.chainView = make(map[string]ast.ScopeContext, 64)
		}
		prev, had := r.chainView[id.Name]
		cur.shadows = append(cur.shadows, shadowEntry{name: id.Name, prev: prev, hadPrev: had})
		r.chainView[id.Name] = cur.ctx
	}
	cur.declaredSymbols[id.Name] = kind

	id.ScopeContext = cur.ctx
}

func (r *Resolver) lookupContext(sym string) (ast.ScopeContext, *Scope) {
	if ctx, ok := r.chainView[sym]; ok {
		return ctx, nil
	}
	return UnresolvedMark, nil
}

func (r *Resolver) withIdentType(kind IdentType, fn func()) {
	old := r.identType
	r.identType = kind
	fn()
	r.identType = old
}

func (r *Resolver) visitRef(n ast.VisitableNode) {
	if n == nil {
		return
	}
	rv := reflect.ValueOf(n)
	if rv.Kind() == reflect.Pointer && rv.IsNil() {
		return
	}
	r.withIdentType(IdentTypeRef, func() {
		n.VisitWith(r)
	})
}

func (r *Resolver) declareBindingTarget(target *ast.BindingTarget) {
	if target == nil {
		return
	}
	r.declareBindingExpr(target.Target)
}

func (r *Resolver) declareBindingExpr(expr ast.Expr) {
	switch expr := expr.(type) {
	case nil:
		return
	case *ast.Identifier:
		r.modify(expr, DeclKindVar)
	case *ast.ArrayPattern:
		for i := range expr.Elements {
			r.declareBindingExpr(expr.Elements[i].Expr)
		}
		if expr.Rest != nil {
			r.declareBindingExpr(expr.Rest.Expr)
		}
	case *ast.ObjectPattern:
		for i := range expr.Properties {
			switch prop := expr.Properties[i].Prop.(type) {
			case *ast.PropertyShort:
				r.modify(prop.Name, DeclKindVar)
			case *ast.PropertyKeyed:
				if prop.Value != nil {
					r.declareBindingExpr(prop.Value.Expr)
				}
			}
		}
		r.declareBindingExpr(expr.Rest)
	case *ast.AssignExpression:
		r.declareBindingExpr(expr.Left.Expr)
	}
}

func (r *Resolver) resolveBindingTargetRefs(target *ast.BindingTarget) {
	if target == nil {
		return
	}
	r.resolveBindingExprRefs(target.Target)
}

func (r *Resolver) resolveBindingExprRefs(expr ast.Expr) {
	switch expr := expr.(type) {
	case nil, *ast.Identifier, *ast.InvalidExpression:
		return
	case *ast.ArrayPattern:
		for i := range expr.Elements {
			r.resolveBindingExprRefs(expr.Elements[i].Expr)
		}
		if expr.Rest != nil {
			r.resolveBindingExprRefs(expr.Rest.Expr)
		}
	case *ast.ObjectPattern:
		for i := range expr.Properties {
			switch prop := expr.Properties[i].Prop.(type) {
			case *ast.PropertyShort:
				r.visitRef(prop.Initializer)
			case *ast.PropertyKeyed:
				if prop.Computed {
					r.visitRef(prop.Key)
				}
				if prop.Value != nil {
					r.resolveBindingExprRefs(prop.Value.Expr)
				}
			}
		}
		r.resolveBindingExprRefs(expr.Rest)
	case *ast.AssignExpression:
		r.resolveBindingExprRefs(expr.Left.Expr)
		r.visitRef(expr.Right)
	default:
		r.visitRef(expr)
	}
}

func (r *Resolver) visitParameterList(n *ast.ParameterList) {
	if n == nil {
		return
	}

	for i := range n.List {
		r.declareBindingTarget(n.List[i].Target)
	}
	r.declareBindingExpr(n.Rest)

	for i := range n.List {
		r.resolveBindingTargetRefs(n.List[i].Target)
		r.visitRef(n.List[i].Initializer)
	}
	r.resolveBindingExprRefs(n.Rest)
}

func (r *Resolver) VisitProgram(n *ast.Program) {
	r.pushScope(ScopeKindBlock)
	n.VisitChildrenWith(r)
	r.popScope()
}

func (r *Resolver) VisitStatements(n *ast.Statements) {
	h := getHoister(r)
	n.VisitWith(h)
	putHoister(h)

	n.VisitChildrenWith(r)
}

func (r *Resolver) VisitBlockStatement(n *ast.BlockStatement) {
	r.pushScope(ScopeKindBlock)
	n.ScopeContext = r.current.ctx
	n.VisitChildrenWith(r)
	r.popScope()
}

func (r *Resolver) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	r.pushScope(ScopeKindFunction)
	if n.Name != nil && n.Name.ScopeContext == UnresolvedMark {
		r.modify(n.Name, DeclKindFunction)
	}

	n.ScopeContext = r.current.ctx

	r.visitParameterList(n.ParameterList)
	// Prevent creating a new scope for the body.
	n.Body.ScopeContext = r.current.ctx
	n.Body.VisitChildrenWith(r)

	r.popScope()
}

func (r *Resolver) VisitArrowFunctionLiteral(n *ast.ArrowFunctionLiteral) {
	r.pushScope(ScopeKindFunction)

	n.ScopeContext = r.current.ctx

	r.visitParameterList(n.ParameterList)

	switch body := n.Body.Body.(type) {
	case *ast.BlockStatement:
		body.ScopeContext = r.current.ctx
		// Prevent creating a new scope for the body.
		body.VisitChildrenWith(r)
	case *ast.Expression:
		r.visitRef(body)
	}

	r.popScope()
}

func (r *Resolver) VisitForStatement(n *ast.ForStatement) {
	r.pushScope(ScopeKindBlock)

	oldIdentType := r.identType
	r.identType = IdentTypeBinding
	if n.Initializer != nil {
		n.Initializer.VisitWith(r)
	}

	r.identType = IdentTypeRef
	n.Test.VisitWith(r)
	n.Update.VisitWith(r)

	r.identType = oldIdentType
	n.Body.VisitWith(r)

	r.popScope()
}

func (r *Resolver) VisitForInStatement(n *ast.ForInStatement) {
	r.pushScope(ScopeKindBlock)

	oldIdentType := r.identType
	r.identType = IdentTypeRef
	n.Into.VisitWith(r)
	n.Source.VisitWith(r)
	n.Body.VisitWith(r)
	r.identType = oldIdentType

	r.popScope()
}

func (r *Resolver) VisitForOfStatement(n *ast.ForOfStatement) {
	r.pushScope(ScopeKindBlock)

	oldIdentType := r.identType
	r.identType = IdentTypeRef
	n.Into.VisitWith(r)
	n.Source.VisitWith(r)
	n.Body.VisitWith(r)
	r.identType = oldIdentType

	r.popScope()
}

func (r *Resolver) VisitCatchStatement(n *ast.CatchStatement) {
	r.pushScope(ScopeKindBlock)

	if n.Parameter != nil {
		r.declareBindingTarget(n.Parameter)
		r.resolveBindingTargetRefs(n.Parameter)
	}

	if n.Body != nil {
		n.Body.VisitWith(r)
	}

	r.popScope()
}

func (r *Resolver) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	for i := range n.List {
		r.declareBindingTarget(n.List[i].Target)
	}

	for i := range n.List {
		r.resolveBindingTargetRefs(n.List[i].Target)
		if n.List[i].Initializer != nil {
			r.visitRef(n.List[i].Initializer)
		}
	}
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
		r.modify(n, DeclKindVar)
	case IdentTypeRef:
		if mark, _ := r.lookupContext(n.Name); mark != UnresolvedMark {
			n.ScopeContext = mark
		}
	}
}

func (r *Resolver) VisitAssignExpression(n *ast.AssignExpression) {
	if n == nil {
		return
	}
	if r.identType == IdentTypeBinding && n.Operator == ast.AssignmentAssign {
		r.withIdentType(IdentTypeBinding, func() {
			n.Left.VisitWith(r)
		})
		r.visitRef(n.Right)
		return
	}

	r.visitRef(n.Left)
	r.visitRef(n.Right)
}

func (r *Resolver) VisitPropertyShort(n *ast.PropertyShort) {
	if n == nil {
		return
	}
	if n.Name != nil {
		n.Name.VisitWith(r)
	}
	r.visitRef(n.Initializer)
}

func (r *Resolver) VisitPropertyKeyed(n *ast.PropertyKeyed) {
	if n == nil {
		return
	}
	if n.Computed {
		r.visitRef(n.Key)
	}
	if n.Value != nil {
		n.Value.VisitWith(r)
	}
}

func (r *Resolver) VisitClassDeclaration(n *ast.ClassDeclaration) {
	if n == nil || n.Class == nil {
		return
	}
	if n.Class.Name != nil && n.Class.Name.Name != "" {
		r.modify(n.Class.Name, DeclKindClass)
	}
	n.Class.VisitWith(r)
}

func (r *Resolver) VisitClassLiteral(n *ast.ClassLiteral) {
	if n == nil {
		return
	}

	needsInnerNameScope := n.Name != nil && n.Name.Name != "" && n.Name.ScopeContext == UnresolvedMark
	if needsInnerNameScope {
		r.pushScope(ScopeKindBlock)
		r.modify(n.Name, DeclKindClass)
		defer r.popScope()
	}

	r.visitRef(n.SuperClass)
	n.Body.VisitWith(r)
}

func (r *Resolver) VisitMemberProperty(n *ast.MemberProperty) {
	if computed, ok := n.Prop.(*ast.ComputedProperty); ok {
		computed.VisitWith(r)
	}
}
