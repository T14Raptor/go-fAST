package resolver

import "github.com/t14raptor/go-fast/ast"

func (r *Resolver) bindDeclarator(n ast.VariableDeclarator) {
	r.declareBindingTarget(n.Target)
	r.resolveBindingTargetReferences(n.Target)
	if n.Initializer != nil {
		n.Initializer.VisitWith(r)
	}
}

func (r *Resolver) declareBindingTarget(n *ast.BindingTarget) {
	forEachBindingIdentInTarget(n, func(id *ast.Identifier) {
		r.modify(id, r.declKind)
	})
}

func (r *Resolver) declareBindingExpression(n *ast.Expression) {
	forEachBindingIdentInExpression(n, func(id *ast.Identifier) {
		r.modify(id, r.declKind)
	})
}

func (r *Resolver) resolveBindingTargetReferences(n *ast.BindingTarget) {
	if n == nil || n.IsNone() {
		return
	}

	if pattern, ok := n.ArrPat(); ok {
		r.resolveArrayPatternBindingReferences(pattern)
		return
	}

	if pattern, ok := n.ObjPat(); ok {
		r.resolveObjectPatternBindingReferences(pattern)
		return
	}

	if _, ok := n.Ident(); ok {
		return
	}

	n.VisitChildrenWith(r)
}

func (r *Resolver) resolveBindingExpressionReferences(n *ast.Expression) {
	if n == nil || n.IsNone() {
		return
	}

	if pattern, ok := n.ArrPat(); ok {
		r.resolveArrayPatternBindingReferences(pattern)
		return
	}

	if pattern, ok := n.ObjPat(); ok {
		r.resolveObjectPatternBindingReferences(pattern)
		return
	}

	if assign, ok := n.Assign(); ok {
		r.resolveBindingExpressionReferences(assign.Left)
		assign.Right.VisitWith(r)
		return
	}

	if keyed, ok := n.Keyed(); ok {
		if keyed.Computed {
			keyed.Key.VisitWith(r)
		}
		r.resolveBindingExpressionReferences(keyed.Value)
		return
	}

	if short, ok := n.Short(); ok {
		if short.Initializer != nil {
			short.Initializer.VisitWith(r)
		}
		return
	}

	if spread, ok := n.Spread(); ok {
		r.resolveBindingExpressionReferences(spread.Expression)
	}
}

func (r *Resolver) resolveArrayPatternBindingReferences(n *ast.ArrayPattern) {
	for i := range n.Elements {
		r.resolveBindingExpressionReferences(&n.Elements[i])
	}
	r.resolveBindingExpressionReferences(n.Rest)
}

func (r *Resolver) resolveObjectPatternBindingReferences(n *ast.ObjectPattern) {
	for i := range n.Properties {
		r.resolvePropertyBindingReferences(&n.Properties[i])
	}
	r.resolveBindingExpressionReferences(n.Rest)
}

func (r *Resolver) resolvePropertyBindingReferences(n *ast.Property) {
	if n == nil || n.IsNone() {
		return
	}

	if keyed, ok := n.Keyed(); ok {
		if keyed.Computed {
			keyed.Key.VisitWith(r)
		}
		r.resolveBindingExpressionReferences(keyed.Value)
		return
	}

	if short, ok := n.Short(); ok {
		if short.Initializer != nil {
			short.Initializer.VisitWith(r)
		}
		return
	}

	if spread, ok := n.Spread(); ok {
		r.resolveBindingExpressionReferences(spread.Expression)
	}
}

func forEachBindingIdentInTarget(n *ast.BindingTarget, visit func(*ast.Identifier)) {
	if n == nil || n.IsNone() {
		return
	}

	if ident, ok := n.Ident(); ok {
		visit(ident)
		return
	}

	if pattern, ok := n.ArrPat(); ok {
		forEachBindingIdentInArrayPattern(pattern, visit)
		return
	}

	if pattern, ok := n.ObjPat(); ok {
		forEachBindingIdentInObjectPattern(pattern, visit)
	}
}

func forEachBindingIdentInExpression(n *ast.Expression, visit func(*ast.Identifier)) {
	if n == nil || n.IsNone() {
		return
	}

	if ident, ok := n.Ident(); ok {
		visit(ident)
		return
	}

	if pattern, ok := n.ArrPat(); ok {
		forEachBindingIdentInArrayPattern(pattern, visit)
		return
	}

	if pattern, ok := n.ObjPat(); ok {
		forEachBindingIdentInObjectPattern(pattern, visit)
		return
	}

	if assign, ok := n.Assign(); ok {
		forEachBindingIdentInExpression(assign.Left, visit)
		return
	}

	if keyed, ok := n.Keyed(); ok {
		forEachBindingIdentInExpression(keyed.Value, visit)
		return
	}

	if short, ok := n.Short(); ok {
		visit(short.Name)
		return
	}

	if spread, ok := n.Spread(); ok {
		forEachBindingIdentInExpression(spread.Expression, visit)
	}
}

func forEachBindingIdentInArrayPattern(n *ast.ArrayPattern, visit func(*ast.Identifier)) {
	for i := range n.Elements {
		forEachBindingIdentInExpression(&n.Elements[i], visit)
	}
	forEachBindingIdentInExpression(n.Rest, visit)
}

func forEachBindingIdentInObjectPattern(n *ast.ObjectPattern, visit func(*ast.Identifier)) {
	for i := range n.Properties {
		forEachBindingIdentInProperty(&n.Properties[i], visit)
	}
	forEachBindingIdentInExpression(n.Rest, visit)
}

func forEachBindingIdentInProperty(n *ast.Property, visit func(*ast.Identifier)) {
	if n == nil || n.IsNone() {
		return
	}

	if keyed, ok := n.Keyed(); ok {
		forEachBindingIdentInExpression(keyed.Value, visit)
		return
	}

	if short, ok := n.Short(); ok {
		visit(short.Name)
		return
	}

	if spread, ok := n.Spread(); ok {
		forEachBindingIdentInExpression(spread.Expression, visit)
	}
}
