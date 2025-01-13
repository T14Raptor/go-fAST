package utils

import "github.com/t14raptor/go-fast/ast"

type BindingCollector struct {
	ast.NoopVisitor
	only     *ast.ScopeContext
	bindings map[ast.Id]struct{}
}

func (v *BindingCollector) add(id ast.Id) {
	if v.only != nil && *v.only != id.ScopeContext {
		return
	}
	v.bindings[id] = struct{}{}
}

func (v *BindingCollector) VisitClassLiteral(n *ast.ClassLiteral) {
	n.VisitChildrenWith(v)
	v.add(n.Name.ToId())
}

func (v *BindingCollector) VisitFunctionDeclaration(n *ast.FunctionDeclaration) {
	n.VisitChildrenWith(v)
	v.add(n.Function.Name.ToId())
}

func (v *BindingCollector) VisitBindingTarget(n *ast.BindingTarget) {
	n.VisitChildrenWith(v)
	if ident, ok := n.Target.(*ast.Identifier); ok {
		v.add(ident.ToId())
	}
}

// CollectDeclarations collects binding identifiers.
func CollectDeclarations(n ast.VisitableNode) map[ast.Id]struct{} {
	visitor := &BindingCollector{
		bindings: make(map[ast.Id]struct{}),
	}
	visitor.V = visitor
	n.VisitWith(visitor)
	return visitor.bindings
}

// CollectDeclarationsInScope collects binding if they are in the given scope.
func CollectDeclarationsInScope(n ast.VisitableNode, scope ast.ScopeContext) map[ast.Id]struct{} {
	visitor := &BindingCollector{
		only:     &scope,
		bindings: make(map[ast.Id]struct{}),
	}
	visitor.V = visitor
	n.VisitWith(visitor)
	return visitor.bindings
}

type DestructingCollector struct {
	ast.NoopVisitor
	bindings map[ast.Id]struct{}
}

func (v *DestructingCollector) VisitExpression(n *ast.Expression) {}
func (v *DestructingCollector) VisitIdentifier(n *ast.Identifier) {
	v.bindings[n.ToId()] = struct{}{}
}

func CollectIdentifiers(n ast.VisitableNode) map[ast.Id]struct{} {
	visitor := &DestructingCollector{
		bindings: make(map[ast.Id]struct{}),
	}
	visitor.V = visitor
	n.VisitWith(visitor)
	return visitor.bindings
}
