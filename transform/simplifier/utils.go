package simplifier

import (
	"slices"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/ast/ext"
	"github.com/t14raptor/go-fast/token"
)

func isNonObj(n *ast.Expression) bool {
	switch n := n.Expr.(type) {
	case *ast.StringLiteral, *ast.NumberLiteral, *ast.NullLiteral, *ast.BooleanLiteral:
		return true
	case *ast.Identifier:
		if n.Name == "undefined" || n.Name == "Infinity" || n.Name == "NaN" {
			return true
		}
	case *ast.UnaryExpression:
		if n.Operator == token.Not || n.Operator == token.Minus || n.Operator == token.Void {
			return isNonObj(n.Operand)
		}
	}
	return false
}

func isObj(n *ast.Expression) bool {
	switch n.Expr.(type) {
	case *ast.ArrayLiteral, *ast.ObjectLiteral, *ast.FunctionLiteral, *ast.NewExpression:
		return true
	default:
		return false
	}
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

func makeBoolExpr(value bool, orig ast.Expressions) ast.Expression {
	return ext.PreserveEffects(ast.Expression{Expr: &ast.BooleanLiteral{Value: value}}, orig)
}

func needZeroForThis(e *ast.Expression) bool {
	_, ok := e.Expr.(*ast.SequenceExpression)
	return directnessMaters(e) || ok
}

func getKeyValue(props []ast.Property, key string) ast.Expr {
	// It's impossible to know the value for certain if a spread property exists.
	if slices.ContainsFunc(props, func(p ast.Property) bool {
		_, ok := p.Prop.(*ast.SpreadElement)
		return ok
	}) {
		return nil
	}

	for _, prop := range slices.Backward(props) {
		switch prop := prop.Prop.(type) {
		case *ast.PropertyShort:
			if prop.Name.Name == key {
				return prop.Name
			}
		case *ast.PropertyKeyed:
			if key != "__proto__" && ext.PropNameEq(prop.Key, "__proto__") {
				// If __proto__ is defined, we need to check the contents of it,
				// as well as any nested __proto__ objects
				if obj, ok := prop.Value.Expr.(*ast.ObjectLiteral); ok {
					if v := getKeyValue(obj.Value, key); v != nil {
						return v
					}
				}
				return nil
			} else if ext.PropNameEq(prop.Key, key) {
				return prop.Value.Expr
			}
		}
	}

	return nil
}
