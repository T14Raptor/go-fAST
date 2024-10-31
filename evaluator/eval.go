package evaluator

import (
	"fmt"
	"math"
	"strings"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

func Eval(n *ast.Expression) (*ast.Expression, bool) {
	switch expr := n.Expr.(type) {
	case *ast.SequenceExpression:
		if len(expr.Sequence) == 0 {
			return nil, false
		}
		return Eval(&expr.Sequence[len(expr.Sequence)-1])
	case *ast.NumberLiteral, *ast.StringLiteral, *ast.BooleanLiteral, *ast.NullLiteral:
		return n.Clone(), true
	case *ast.ConditionalExpression:
		test, ok := Eval(expr.Test)
		if !ok {
			return nil, false
		}
		res, ok := isTruthy(test)
		if !ok {
			return nil, false
		}
		if res {
			return Eval(expr.Consequent)
		} else {
			return Eval(expr.Alternate)
		}
	case *ast.MemberExpression:
		if strLit1, ok := expr.Object.Expr.(*ast.StringLiteral); ok {
			if strLit2, ok := expr.Property.Expr.(*ast.StringLiteral); ok && strLit2.Value == "length" {
				return &ast.Expression{Expr: &ast.NumberLiteral{
					Literal: fmt.Sprintf("%d", len(strLit1.Value)),
					Value:   float64(len(strLit1.Value)),
				}}, true
			}
		}
	case *ast.UnaryExpression:
		switch expr.Operator {
		case token.Void:
			return &ast.Expression{Expr: &ast.Identifier{Name: "undefined"}}, true
		case token.Not:
			arg, ok := Eval(expr.Operand)
			if !ok {
				return nil, false
			}
			res, ok := isTruthy(arg)
			if !ok {
				return nil, false
			}
			if res {
				return &ast.Expression{Expr: &ast.BooleanLiteral{Value: false}}, true
			} else {
				return &ast.Expression{Expr: &ast.BooleanLiteral{Value: true}}, true
			}
		}
	case *ast.ArrayLiteral:
		exprClone := expr.Clone()
		for i, elem := range expr.Value {
			if e, ok := Eval(&elem); !ok {
				return nil, false
			} else {
				exprClone.Value[i] = *e
			}
		}
		return &ast.Expression{Expr: exprClone}, true
	case *ast.ObjectLiteral:
		exprClone := expr.Clone()
		for i, prop := range expr.Value {
			propKeyed, ok := prop.Prop.(*ast.PropertyKeyed)
			if !ok {
				return nil, false
			}
			key, ok := Eval(propKeyed.Key)
			if !ok {
				return nil, false
			}
			value, ok := Eval(propKeyed.Value)
			if !ok {
				return nil, false
			}
			exprClone.Value[i] = ast.Property{
				Prop: &ast.PropertyKeyed{
					Key:   key,
					Value: value,
				},
			}
		}
		return &ast.Expression{Expr: exprClone}, true
	case *ast.BinaryExpression:
		left, ok := Eval(expr.Left)
		if !ok {
			return nil, false
		}
		right, ok := Eval(expr.Right)
		if !ok {
			return nil, false
		}
		var leftValue, rightValue Value
		switch leftExpr := left.Expr.(type) {
		case *ast.NumberLiteral:
			leftValue = toValue(leftExpr.Value)
		case *ast.StringLiteral:
			leftValue = stringValue(leftExpr.Value)
		case *ast.BooleanLiteral:
			leftValue = boolValue(leftExpr.Value)
		case *ast.NullLiteral:
			leftValue = nullValue
		default:
			return nil, false
		}
		switch rightExpr := right.Expr.(type) {
		case *ast.NumberLiteral:
			rightValue = toValue(rightExpr.Value)
		case *ast.StringLiteral:
			rightValue = stringValue(rightExpr.Value)
		case *ast.BooleanLiteral:
			rightValue = boolValue(rightExpr.Value)
		case *ast.NullLiteral:
			rightValue = nullValue
		default:
			return nil, false
		}
		res, ok := calculateBinaryExpression(expr.Operator, leftValue, rightValue)
		if !ok {
			return nil, false
		}
		switch res.kind {
		case valueNumber:
			return &ast.Expression{Expr: &ast.NumberLiteral{
				Literal: fmt.Sprintf("%v", res.value),
				Value:   res.value,
			}}, true
		case valueString:
			return &ast.Expression{Expr: &ast.StringLiteral{
				Literal: fmt.Sprintf("%v", res.value),
				Value:   res.value.(string),
			}}, true
		case valueBoolean:
			return &ast.Expression{Expr: &ast.BooleanLiteral{
				Value: res.value.(bool),
			}}, true
		case valueNull:
			return &ast.Expression{Expr: &ast.NullLiteral{}}, true
		}
	}
	return nil, false
}

func isTruthy(n *ast.Expression) (truthy bool, ok bool) {
	if n == nil {
		return false, false
	}
	switch expr := n.Expr.(type) {
	case *ast.StringLiteral:
		return expr.Value != "", true
	case *ast.BooleanLiteral:
		return expr.Value, true
	case *ast.NullLiteral:
		return false, true
	case *ast.NumberLiteral:
		switch num := expr.Value.(type) {
		case int64:
			return num != 0, true
		case float64:
			return expr.Value != 0.0, true
		default:
			return false, false
		}
	case *ast.Identifier:
		if expr.Name == "undefined" {
			return false, true
		}
		return false, false
	case *ast.ArrayLiteral, *ast.ObjectLiteral:
		return true, true
	default:
		return false, false
	}
}

func evaluateDivide(left float64, right float64) Value {
	if math.IsNaN(left) || math.IsNaN(right) {
		return NaNValue()
	}
	if math.IsInf(left, 0) && math.IsInf(right, 0) {
		return NaNValue()
	}
	if left == 0 && right == 0 {
		return NaNValue()
	}
	if math.IsInf(left, 0) {
		if math.Signbit(left) == math.Signbit(right) {
			return positiveInfinityValue()
		}
		return negativeInfinityValue()
	}
	if math.IsInf(right, 0) {
		if math.Signbit(left) == math.Signbit(right) {
			return positiveZeroValue()
		}
		return negativeZeroValue()
	}
	if right == 0 {
		if math.Signbit(left) == math.Signbit(right) {
			return positiveInfinityValue()
		}
		return negativeInfinityValue()
	}
	return float64Value(left / right)
}

func calculateBinaryExpression(operator token.Token, left Value, right Value) (Value, bool) {
	switch operator {
	// Additive
	case token.Plus:
		if left.IsString() || right.IsString() {
			return stringValue(strings.Join([]string{left.string(), right.string()}, "")), true
		}
		return float64Value(left.float64() + right.float64()), true
	case token.Minus:
		return float64Value(left.float64() - right.float64()), true

	// Multiplicative
	case token.Multiply:
		return float64Value(left.float64() * right.float64()), true
	case token.Slash:
		return evaluateDivide(left.float64(), right.float64()), true
	case token.Remainder:
		return float64Value(math.Mod(left.float64(), right.float64())), true

	// Logical
	case token.LogicalAnd:
		left := left.bool()
		if !left {
			return falseValue, true
		}
		return boolValue(right.bool()), true
	case token.LogicalOr:
		left := left.bool()
		if left {
			return trueValue, true
		}
		return boolValue(right.bool()), true

	// Bitwise
	case token.And:
		return int32Value(toInt32(left) & toInt32(right)), true
	case token.Or:
		return int32Value(toInt32(left) | toInt32(right)), true
	case token.ExclusiveOr:
		return int32Value(toInt32(left) ^ toInt32(right)), true

	// Shift
	// (Masking of 0x1f is to restrict the shift to a maximum of 31 places)
	case token.ShiftLeft:
		return int32Value(toInt32(left) << (toUint32(right) & 0x1f)), true
	case token.ShiftRight:
		return int32Value(toInt32(left) >> (toUint32(right) & 0x1f)), true
	case token.UnsignedShiftRight:
		// Shifting an unsigned integer is a logical shift
		return uint32Value(toUint32(left) >> (toUint32(right) & 0x1f)), true
	}

	return Value{}, false
}
