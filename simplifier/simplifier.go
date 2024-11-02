package simplifier

import (
	"strconv"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/resolver"
)

type Simplifier struct {
	ast.NoopVisitor

	changed       bool
	isArgOfUpdate bool
	isModifying   bool
	inCallee      bool
}

func (s *Simplifier) optimizeMemberExpression(n *ast.MemberExpression) {
	if s.isModifying {
		return
	}

	// [a, b].length
	type Len struct{}
	// [a, b][0]
	//
	// {0.5: "bar"}[0.5]
	// Note: callers need to check `v.fract() == 0.0` in some cases.
	// ie non-integer indexes for arrays result in `undefined`
	// but not for objects (because indexing an object
	// returns the value of the key, ie `0.5` will not
	// return `undefined` if a key `0.5` exists
	// and its value is not `undefined`).
	type Index float64
	// ({}).foo
	type IndexStr string

	var op any
	switch prop := n.Property.Expr.(type) {
	case *ast.Identifier:
		if _, ok := n.Object.Expr.(*ast.ObjectLiteral); !ok && prop.Name == "length" {
			op = Len{}
		}
		if s.inCallee {
			return
		}
		op = IndexStr(prop.Name)
	case *ast.NumberLiteral:
		if s.inCallee {
			return
		}
		// x[5]
		op = Index(prop.Value)
	default:
		if s.inCallee {
			return
		}
		if s, ok := asPureString(n.Property); ok {
			if _, ok := n.Object.Expr.(*ast.ObjectLiteral); !ok && s == "length" {
				// Length of non-object type
				op = Len{}
			} else if n, err := strconv.ParseFloat(s, 64); err == nil {
				// x['0'] is treated as x[0]
				op = Index(n)
			} else {
				// x[''] or x[...] where ... is an expression like [], ie x[[]]
				op = IndexStr(s)
			}
		} else {
			return
		}
	}
	_ = op
}

func Simplify(p *ast.Program) {
	resolver.Resolve(p)

	visitor := &Simplifier{}
	visitor.V = visitor
	p.VisitWith(visitor)
}
