package resolver_test

import (
	"testing"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/parser"
	"github.com/t14raptor/go-fast/resolver"
)

type idVisitor struct {
	ast.NoopVisitor
	ids []ast.Id
}

func (v *idVisitor) VisitIdentifier(n *ast.Identifier) {
	if n.Name == "P8" {
		v.ids = append(v.ids, n.ToId())
	}
}

func TestFunctionRestParameterCreatesFunctionScopeBinding(t *testing.T) {
	program, err := parser.ParseFile(`function P8() { function inner(...P8) { P8.push(undefined); } P8(); }`)
	if err != nil {
		t.Fatal(err)
	}

	resolver.Resolve(program)

	visitor := &idVisitor{}
	visitor.V = visitor
	program.VisitWith(visitor)

	if len(visitor.ids) != 4 {
		t.Fatalf("expected 4 P8 identifiers, got %d", len(visitor.ids))
	}

	outer := visitor.ids[0]
	rest := visitor.ids[1]
	use := visitor.ids[2]
	outerUse := visitor.ids[3]

	if rest.ScopeContext == outer.ScopeContext {
		t.Fatalf("rest parameter resolved to outer binding: rest=%+v outer=%+v", rest, outer)
	}
	if use.ScopeContext != rest.ScopeContext {
		t.Fatalf("rest parameter use resolved to wrong binding: use=%+v rest=%+v", use, rest)
	}
	if outerUse.ScopeContext != outer.ScopeContext {
		t.Fatalf("outer use resolved to wrong binding: use=%+v outer=%+v", outerUse, outer)
	}
}
