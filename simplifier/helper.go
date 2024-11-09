package simplifier

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

func isStr(n *ast.Expression) bool {
	switch n := n.Expr.(type) {
	case *ast.StringLiteral, *ast.TemplateLiteral:
		return true
	case *ast.UnaryExpression:
		if n.Operator == token.Typeof {
			return true
		}
	case *ast.BinaryExpression:
		if n.Operator == token.Plus {
			return isStr(n.Left) || isStr(n.Right)
		}
	case *ast.AssignExpression:
		if n.Operator == token.Assign || n.Operator == token.AddAssign {
			return isStr(n.Right)
		}
	case *ast.SequenceExpression:
		if len(n.Sequence) == 0 {
			return false
		}
		return isStr(&n.Sequence[len(n.Sequence)-1])
	case *ast.ConditionalExpression:
		return isStr(n.Consequent) && isStr(n.Alternate)
	}
	return false
}

func isArrayLiteral(n *ast.Expression) bool {
	_, ok := n.Expr.(*ast.ArrayLiteral)
	return ok
}

func directnessMaters(n *ast.Expression) bool {
	switch n := n.Expr.(type) {
	case *ast.Identifier:
		return n.Name == "eval"
	case *ast.MemberExpression:
		return true
	}
	return false
}
