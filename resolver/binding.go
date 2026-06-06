package resolver

import "github.com/t14raptor/go-fast/ast"

// resolverBinder declares each identifier in binding position within a pattern.
type resolverBinder struct {
	ast.NoopVisitor

	r *Resolver
}

func (b *resolverBinder) VisitExpression(*ast.Expression) {}

func (b *resolverBinder) VisitIdentifier(n *ast.Identifier) {
	b.r.modify(n, b.r.declKind)
}
