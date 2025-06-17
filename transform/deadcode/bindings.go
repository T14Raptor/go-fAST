package deadcode

import "github.com/t14raptor/go-fast/ast"

type bindingCollector struct {
	ast.NoopVisitor
	only     *ast.ScopeContext
	bindings map[ast.Id]struct{}
}

func (v *bindingCollector) add(id ast.Id) {
	if v.only != nil && *v.only != id.ScopeContext {
		return
	}
	v.bindings[id] = struct{}{}
}

func (v *bindingCollector) VisitClassLiteral(n *ast.ClassLiteral) {
	n.VisitChildrenWith(v)
	v.add(n.Name.ToId())
}

func (v *bindingCollector) VisitFunctionDeclaration(n *ast.FunctionDeclaration) {
	n.VisitChildrenWith(v)
	v.add(n.Function.Name.ToId())
}

func (v *bindingCollector) VisitBindingTarget(n *ast.BindingTarget) {
	n.VisitChildrenWith(v)
	if ident, ok := n.Target.(*ast.Identifier); ok {
		v.add(ident.ToId())
	}
}

// collectDeclarations collects binding identifiers.
func collectDeclarations(n ast.VisitableNode) map[ast.Id]struct{} {
	visitor := &bindingCollector{
		bindings: make(map[ast.Id]struct{}),
	}
	visitor.V = visitor
	n.VisitWith(visitor)
	return visitor.bindings
}

// collectDeclarationsInScope collects binding if they are in the given scope.
func collectDeclarationsInScope(n ast.VisitableNode, scope ast.ScopeContext) map[ast.Id]struct{} {
	visitor := &bindingCollector{
		only:     &scope,
		bindings: make(map[ast.Id]struct{}),
	}
	visitor.V = visitor
	n.VisitWith(visitor)
	return visitor.bindings
}

type destructingCollector struct {
	ast.NoopVisitor
	bindings map[ast.Id]struct{}
}

func (v *destructingCollector) VisitExpression(n *ast.Expression) {}
func (v *destructingCollector) VisitIdentifier(n *ast.Identifier) {
	v.bindings[n.ToId()] = struct{}{}
}

func collectIdentifiers(n ast.VisitableNode) map[ast.Id]struct{} {
	visitor := &destructingCollector{
		bindings: make(map[ast.Id]struct{}),
	}
	visitor.V = visitor
	n.VisitWith(visitor)
	return visitor.bindings
}
