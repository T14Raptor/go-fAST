package simplifier

import (
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/ast/ext"
	"github.com/t14raptor/go-fast/resolver"
	"github.com/t14raptor/go-fast/token"
)

type Simplifier struct {
	ast.NoopVisitor

	changed       bool
	isArgOfUpdate bool
	isModifying   bool
	inCallee      bool
}

func (s *Simplifier) optimizeMemberExpression(expr *ast.Expression) {
	memExpr, ok := expr.Expr.(*ast.MemberExpression)
	if !ok {
		return
	}
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
	switch prop := memExpr.Property.Expr.(type) {
	case *ast.Identifier:
		if _, ok := memExpr.Object.Expr.(*ast.ObjectLiteral); !ok && prop.Name == "length" {
			op = Len{}
		} else if s.inCallee {
			return
		} else {
			op = IndexStr(prop.Name)
		}
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
		if s, ok := ext.AsPureString(memExpr.Property); ok {
			if _, ok := memExpr.Object.Expr.(*ast.ObjectLiteral); !ok && s == "length" {
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

	// Note: pristine_globals refers to the compress config option pristine_globals.
	// Any potential cases where globals are not pristine are handled in compress,
	// e.g. x[-1] is not changed as the object's prototype may be modified.
	// For example, Array.prototype[-1] = "foo" will result in [][-1] returning
	// "foo".

	switch obj := memExpr.Object.Expr.(type) {
	case *ast.StringLiteral:
		switch op := op.(type) {
		// 'foo'.length
		//
		// Prototype changes do not affect .length, so we don't need to worry
		// about pristine_globals here.
		case Len:
			s.changed = true
			expr.Expr = &ast.NumberLiteral{Value: float64(len(obj.Value))}

		// 'foo'[1]
		case Index:
			idx := float64(op)
			if _, frac := math.Modf(float64(idx)); frac != 0.0 || idx < 0.0 || int(idx) >= len(obj.Value) {
				// Prototype changes affect indexing if the index is out of bounds, so we
				// don't replace out-of-bound indexes.
				return
			}

			value, ok := nthChar(obj.Value, int(idx))
			if !ok {
				return
			}

			s.changed = true
			expr.Expr = &ast.StringLiteral{Value: value}

		// 'foo'['']
		//
		// Handled in compress
		case IndexStr:
		}

	// [1, 2, 3].length
	//
	// [1, 2, 3][0]
	case *ast.ArrayLiteral:
		// do nothing if spread exists
		if slices.ContainsFunc(obj.Value, func(e ast.Expression) bool {
			_, ok := e.Expr.(*ast.SpreadElement)
			return ok
		}) {
			return
		}

		switch op := op.(type) {
		case Len:
			// do nothing if replacement will have side effects
			if slices.ContainsFunc(obj.Value, func(e ast.Expression) bool {
				return ext.MayHaveSideEffects(&e)
			}) {
				return
			}

			s.changed = true
			expr.Expr = &ast.NumberLiteral{Value: float64(len(obj.Value))}
		case Index:
			idx := int(op)
			// If the fraction part is non-zero, or if the index is out of bounds,
			// then we handle this in compress as Array's prototype may be modified.
			if _, frac := math.Modf(float64(idx)); frac != 0.0 || idx < 0 || idx >= len(obj.Value) {
				return
			}

			// Don't change if after has side effects.
			if slices.ContainsFunc(obj.Value[idx+1:], func(e ast.Expression) bool {
				return ext.MayHaveSideEffects(&e)
			}) {
				return
			}

			s.changed = true

			// elements before target element
			before := obj.Value[:idx]
			// element at idx
			e := obj.Value[idx]
			// elements after target element
			after := obj.Value[idx+1:]

			// element value
			var v ast.Expr
			if e.Expr == nil {
				v = &ast.UnaryExpression{
					Operator: token.Void,
					Operand:  &ast.Expression{Expr: &ast.NumberLiteral{Value: 0.0}},
				}
			} else {
				v = e.Expr
			}

			// Replacement expressions.
			var exprs []ast.Expression

			// Add before side effects.
			for _, elem := range before {
				ext.ExtractSideEffectsTo(&exprs, &elem)
			}

			// Element value.
			val := v

			// Add after side effects.
			for _, elem := range after {
				ext.ExtractSideEffectsTo(&exprs, &elem)
			}

			// Note: we always replace with a SeqExpr so that
			// `this` remains undefined in strict mode.

			if exprs == nil {
				// No side effects exist, replace with:
				// (0, val)
				expr.Expr = &ast.SequenceExpression{
					Sequence: []ast.Expression{
						{Expr: &ast.NumberLiteral{Value: 0.0}},
						{Expr: val},
					},
				}
			}

			// Add value and replace with SeqExpr
			exprs = append(exprs, ast.Expression{Expr: val})
			expr.Expr = &ast.SequenceExpression{Sequence: exprs}

		// Handled in compress
		case IndexStr:
		}

	// { foo: true }['foo']
	//
	// { 0.5: true }[0.5]
	case *ast.ObjectLiteral:
		// get key
		var key string
		switch op := op.(type) {
		case Index:
			key = strconv.FormatFloat(float64(op), 'f', -1, 64)
		case IndexStr:
			if op != "yield" && ext.IsLiteral(&obj.Value) {
				key = string(op)
			}
		}

		// Get `key`s value. Non-existent keys are handled in compress.
		// This also checks if spread exists.
		v := getKeyValue(obj.Value, key)
		if v == nil {
			return
		}

		s.changed = true
		expr.Expr = ext.PreserveEffects(ast.Expression{Expr: v}, []ast.Expression{{Expr: obj}}).Expr
	}
}

func (s *Simplifier) optimizeBinaryExpression(expr *ast.Expression) {
	binExpr, ok := expr.Expr.(*ast.BinaryExpression)
	if !ok {
		return
	}

	tryReplaceBool := func(v bool, left, right *ast.Expression) {
		s.changed = true
		expr.Expr = makeBoolExpr(v, []ast.Expression{{Expr: left}, {Expr: right}}).Expr
	}
	tryReplaceNum := func(v float64, left, right *ast.Expression) {
		s.changed = true
		var value ast.Expr
		if !math.IsNaN(v) {
			value = &ast.NumberLiteral{Value: v}
		} else {
			value = &ast.Identifier{Name: "NaN"}
		}
		expr.Expr = ext.PreserveEffects(ast.Expression{Expr: value}, []ast.Expression{{Expr: left}, {Expr: right}}).Expr
	}

	switch binExpr.Operator {
	case token.Plus:
		// It's string concatenation if either left or right is string.
		if ext.IsStr(binExpr.Left) || ext.IsArrayLiteral(binExpr.Left) || ext.IsStr(binExpr.Right) || ext.IsArrayLiteral(binExpr.Right) {
			l, lok := ext.AsPureString(binExpr.Left)
			r, rok := ext.AsPureString(binExpr.Right)
			if lok && rok {
				s.changed = true
				expr.Expr = &ast.StringLiteral{Value: l + r}
			}
		}

		typ, ok := ext.GetType(expr)
		if !ok {
			return
		}
		switch typ {
		// String concatenation
		case ext.StringType:
			if !ext.MayHaveSideEffects(binExpr.Left) && !ext.MayHaveSideEffects(binExpr.Right) {
				l, lok := ext.AsPureString(binExpr.Left)
				r, rok := ext.AsPureString(binExpr.Right)
				if lok && rok {
					s.changed = true
					expr.Expr = &ast.StringLiteral{Value: l + r}
				}
			}
		// Numerical calculation
		case ext.BoolType, ext.NullType, ext.NumberType, ext.UndefinedType:
			v, ok := s.performArithmeticOp(token.Plus, binExpr.Left, binExpr.Right)
			if ok {
				tryReplaceNum(v, binExpr.Left, binExpr.Right)
			}
		}

		//TODO: try string concat

	case token.LogicalAnd, token.LogicalOr:
		val, ok, _ := ext.CastToBool(binExpr.Left)
		if ok {
			var node ast.Expression
			if binExpr.Operator == token.LogicalAnd {
				if val {
					// 1 && $right
					node = *binExpr.Right
				} else {
					s.changed = true
					// 0 && $right
					expr.Expr = binExpr.Left.Expr
				}
			} else {
				if val {
					s.changed = true
					// 1 || $right
					expr.Expr = binExpr.Left.Expr
				} else {
					// 0 || $right
					node = *binExpr.Right
				}
			}

			if !ext.MayHaveSideEffects(binExpr.Left) {
				s.changed = true
				if directnessMaters(&node) {
					expr.Expr = &ast.SequenceExpression{
						Sequence: []ast.Expression{
							{Expr: &ast.NumberLiteral{Value: 0.0}},
							{Expr: node.Expr},
						},
					}
				} else {
					expr.Expr = node.Expr
				}
			} else {
				s.changed = true
				seq := &ast.SequenceExpression{
					Sequence: []ast.Expression{
						{Expr: binExpr.Left.Expr},
						{Expr: node.Expr},
					},
				}
				seq.VisitWith(s)
				expr.Expr = seq
			}
		}
	case token.InstanceOf:
		// Non-object types are never instances.
		if isNonObj(binExpr.Left) {
			s.changed = true
			expr.Expr = makeBoolExpr(false, []ast.Expression{{Expr: binExpr.Right}}).Expr
			return
		}
		if isObj(binExpr.Left) && ext.IsGlobalRefTo(binExpr.Right, "Object") {
			s.changed = true
			expr.Expr = makeBoolExpr(true, []ast.Expression{{Expr: binExpr.Left}}).Expr
		}
	case token.Minus, token.Slash, token.Remainder, token.Exponent:
		if v, ok := s.performArithmeticOp(binExpr.Operator, binExpr.Left, binExpr.Right); ok {
			tryReplaceNum(v, binExpr.Left, binExpr.Right)
		}
	case token.ShiftLeft, token.ShiftRight, token.UnsignedShiftRight:
		tryFoldShift := func(op token.Token, left, right *ast.Expression) (float64, bool) {
			if _, ok := left.Expr.(*ast.NumberLiteral); !ok {
				return 0, false
			}
			if _, ok := right.Expr.(*ast.NumberLiteral); !ok {
				return 0, false
			}

			lv, lok := ext.AsPureNumber(left)
			rv, rok := ext.AsPureNumber(right)
			if !lok || !rok {
				return 0, false
			}
			// Shift
			// (Masking of 0x1f is to restrict the shift to a maximum of 31 places)
			switch op {
			case token.ShiftLeft:
				return float64(int32(int32(lv) << uint32(rv) & 0x1f)), true
			case token.ShiftRight:
				return float64(int32(int32(lv)>>uint32(rv)) & 0x1f), true
			case token.UnsignedShiftRight:
				return float64(uint32(uint32(lv) >> uint32(rv) & 0x1f)), true
			}
			return 0, false
		}
		if v, ok := tryFoldShift(binExpr.Operator, binExpr.Left, binExpr.Right); ok {
			tryReplaceNum(v, binExpr.Left, binExpr.Right)
		}

	// These needs one more check.
	//
	// (a * 1) * 2 --> a * (1 * 2) --> a * 2
	case token.Multiply:
		if v, ok := s.performArithmeticOp(binExpr.Operator, binExpr.Left, binExpr.Right); ok {
			tryReplaceNum(v, binExpr.Left, binExpr.Right)
		}

		// Try left.rhs * right
		if binExpr2, ok := binExpr.Left.Expr.(*ast.BinaryExpression); ok && binExpr2.Operator == binExpr.Operator {
			if v, ok := s.performArithmeticOp(binExpr.Operator, binExpr2.Left, binExpr.Right); ok {
				var valExpr ast.Expr
				if !math.IsNaN(v) {
					valExpr = &ast.NumberLiteral{Value: v}
				} else {
					valExpr = &ast.Identifier{Name: "NaN"}
				}
				s.changed = true
				binExpr.Left.Expr = binExpr2.Left.Expr
				binExpr.Right.Expr = valExpr
			}
		}

	// Comparisons
	case token.Less:
		if v, ok := s.performAbstractRelCmp(binExpr.Left, binExpr.Right, false); ok {
			tryReplaceBool(v, binExpr.Left, binExpr.Right)
		}
	case token.Greater:
		if v, ok := s.performAbstractRelCmp(binExpr.Right, binExpr.Left, false); ok {
			tryReplaceBool(v, binExpr.Right, binExpr.Left)
		}
	case token.LessOrEqual:
		if v, ok := s.performAbstractRelCmp(binExpr.Right, binExpr.Left, true); ok {
			tryReplaceBool(v, binExpr.Right, binExpr.Left)
		}
	case token.GreaterOrEqual:
		if v, ok := s.performAbstractRelCmp(binExpr.Left, binExpr.Right, true); ok {
			tryReplaceBool(v, binExpr.Left, binExpr.Right)
		}
	case token.Equal:
		if v, ok := s.performAbstractEqCmp(binExpr.Left, binExpr.Right); ok {
			tryReplaceBool(v, binExpr.Left, binExpr.Right)
		}
	case token.NotEqual:
		if v, ok := s.performAbstractEqCmp(binExpr.Left, binExpr.Right); ok {
			tryReplaceBool(!v, binExpr.Left, binExpr.Right)
		}
	case token.StrictEqual:
		if v, ok := s.performStrictEqCmp(binExpr.Left, binExpr.Right); ok {
			tryReplaceBool(v, binExpr.Left, binExpr.Right)
		}
	case token.StrictNotEqual:
		if v, ok := s.performStrictEqCmp(binExpr.Left, binExpr.Right); ok {
			tryReplaceBool(!v, binExpr.Left, binExpr.Right)
		}
	}
}

func (s *Simplifier) tryFoldTypeOf(expr *ast.Expression) {
	if unary, ok := expr.Expr.(*ast.UnaryExpression); ok && unary.Operator == token.Typeof {
		var val string
		switch operand := unary.Operand.Expr.(type) {
		case *ast.FunctionLiteral:
			val = "function"
		case *ast.StringLiteral:
			val = "string"
		case *ast.NumberLiteral:
			val = "number"
		case *ast.BooleanLiteral:
			val = "boolean"
		case *ast.NullLiteral, *ast.ObjectLiteral, *ast.ArrayLiteral:
			val = "object"
		case *ast.UnaryExpression:
			if operand.Operator == token.Void {
				val = "undefined"
			}
			return
		case *ast.Identifier:
			if operand.Name == "undefined" {
				val = "undefined"
			} else {
				return
			}
		default:
			return
		}
		s.changed = true
		expr.Expr = &ast.StringLiteral{Value: val}
	}
}

func (s *Simplifier) optimizeUnaryExpression(expr *ast.Expression) {
	unaryExpr, ok := expr.Expr.(*ast.UnaryExpression)
	if !ok {
		return
	}
	sideEffects := ext.MayHaveSideEffects(unaryExpr.Operand)

	switch unaryExpr.Operator {
	case token.Typeof:
		if !sideEffects {
			s.tryFoldTypeOf(expr)
		}
	case token.Not:
		switch operand := unaryExpr.Operand.Expr.(type) {
		// Don't expand booleans.
		case *ast.NumberLiteral:
			return
		// Don't remove ! from negated iifes.
		case *ast.CallExpression:
			if _, ok := operand.Callee.Expr.(*ast.FunctionLiteral); ok {
				return
			}
		}
		if val, ok, _ := ext.CastToBool(unaryExpr.Operand); ok {
			s.changed = true
			expr.Expr = makeBoolExpr(!val, []ast.Expression{{Expr: unaryExpr.Operand}}).Expr
		}
	case token.Plus:
		if val, ok := ext.AsPureNumber(unaryExpr.Operand); ok {
			s.changed = true
			if math.IsNaN(val) {
				expr.Expr = ext.PreserveEffects(ast.Expression{Expr: &ast.Identifier{Name: "NaN"}}, []ast.Expression{{Expr: unaryExpr.Operand}}).Expr
				return
			}
			expr.Expr = ext.PreserveEffects(ast.Expression{Expr: &ast.NumberLiteral{Value: val}}, []ast.Expression{{Expr: unaryExpr.Operand}}).Expr
		}
	case token.Minus:
		switch operand := unaryExpr.Operand.Expr.(type) {
		case *ast.Identifier:
			if operand.Name == "Infinity" {
			} else if operand.Name == "NaN" {
				// "-NaN" is "NaN"
				s.changed = true
				expr.Expr = operand
			}
		case *ast.NumberLiteral:
			s.changed = true
			expr.Expr = &ast.NumberLiteral{Value: -operand.Value}
		}
		// TODO: Report that user is something bad (negating
		// non-number value)

	case token.Void:
		if !sideEffects {
			if numLit, ok := unaryExpr.Operand.Expr.(*ast.NumberLiteral); ok && numLit.Value == 0 {
				return
			}
			s.changed = true
			expr.Expr = &ast.NumberLiteral{Value: 0.0}
		}
	case token.BitwiseNot:
		if val, ok := ext.AsPureNumber(unaryExpr.Operand); ok {
			if _, frac := math.Modf(val); frac == 0.0 {
				s.changed = true
				var result float64
				if val < 0.0 {
					result = float64(^uint32(int32(val)))
				} else {
					result = float64(^uint32(val))
				}
				expr.Expr = &ast.NumberLiteral{Value: result}
			}
			// TODO: Report error
		}
	}
}

func (s *Simplifier) performArithmeticOp(op token.Token, left, right *ast.Expression) (float64, bool) {
	lv, lok := ext.AsPureNumber(left)
	rv, rok := ext.AsPureNumber(right)

	typl, _ := ext.GetType(left)
	typr, _ := ext.GetType(right)
	if (!lok && !rok) || op == token.Plus && (!typl.CastToNumberOnAdd() || !typr.CastToNumberOnAdd()) {
		return 0, false
	}

	switch op {
	case token.Plus:
		if lok && rok {
			return lv + rv, true
		}
		if lv == 0.0 && lok {
			return rv, rok
		} else if rv == 0.0 && rok {
			return lv, lok
		}
		return 0, false
	case token.Minus:
		if lok && rok {
			return lv - rv, true
		}

		// 0 - x => -x
		if lv == 0.0 && lok {
			return -rv, rok
		}

		// x - 0 => x
		if rv == 0.0 && rok {
			return lv, lok
		}
		return 0, false
	case token.Multiply:
		if lok && rok {
			return lv * rv, true
		}
		// NOTE: 0*x != 0 for all x, if x==0, then it is NaN.  So we can't take
		// advantage of that without some kind of non-NaN proof.  So the special cases
		// here only deal with 1*x
		if lv == 1.0 && lok {
			return rv, rok
		}
		if rv == 1.0 && rok {
			return lv, lok
		}
		return 0, false
	case token.Slash:
		if lok && rok {
			if rv == 0.0 {
				return 0, false
			}
			return lv / rv, true
		}
		// NOTE: 0/x != 0 for all x, if x==0, then it is NaN
		if rv == 1.0 && rok {
			// TODO: cloneTree
			// x/1->x
			return lv, lok
		}
		return 0, false
	case token.Exponent:
		if rv == 0.0 && rok {
			return 1, true
		}
		if lok && rok {
			return math.Pow(lv, rv), true
		}
		return 0, false
	}

	if !lok || !rok {
		return 0, false
	}

	switch op {
	case token.And:
		return float64(int32(lv) & int32(rv)), true
	case token.Or:
		return float64(int32(lv) | int32(rv)), true
	case token.ExclusiveOr:
		return float64(int32(lv) ^ int32(rv)), true
	case token.Remainder:
		if rv == 0.0 {
			return 0, false
		}
		return float64(int(lv) % int(rv)), true
	}
	return 0, false
}

func (s *Simplifier) performAbstractRelCmp(left, right *ast.Expression, willNegate bool) (bool, bool) {
	// Special case: `x < x` is always false.
	if l, ok := left.Expr.(*ast.Identifier); ok {
		if r, ok := right.Expr.(*ast.Identifier); ok {
			if !willNegate && l.Name == r.Name && l.ScopeContext == r.ScopeContext {
				return false, true
			}
		}
	}
	// Special case: `typeof a < typeof a` is always false.
	if l, ok := left.Expr.(*ast.UnaryExpression); ok && l.Operator == token.Typeof {
		if r, ok := right.Expr.(*ast.UnaryExpression); ok && r.Operator == token.Typeof {
			if lid, lok := l.Operand.Expr.(*ast.Identifier); lok {
				if rid, rok := r.Operand.Expr.(*ast.Identifier); rok {
					if lid.ToId() == rid.ToId() {
						return false, true
					}
				}
			}
		}
	}

	// Try to evaluate based on the general type.
	lt, lok := ext.GetType(left)
	rt, rok := ext.GetType(right)
	if lt == ext.StringType && rt == ext.StringType && lok && rok {
		lv, lok := ext.AsPureString(left)
		rv, rok := ext.AsPureString(right)
		if lok && rok {
			// In JS, browsers parse \v differently. So do not compare strings if one
			// contains \v.
			if strings.ContainsRune(lv, '\u000B') || strings.ContainsRune(rv, '\u000B') {
				return false, false
			} else {
				return lv < rv, true
			}
		}
	}

	// Then, try to evaluate based on the value of the node. Try comparing as
	// numbers.
	lv, lok := ext.AsPureNumber(left)
	rv, rok := ext.AsPureNumber(right)
	if lok && rok {
		if math.IsNaN(lv) || math.IsNaN(rv) {
			return willNegate, true
		}
		return lv < rv, true
	}

	return false, false
}

func (s *Simplifier) performAbstractEqCmp(left, right *ast.Expression) (bool, bool) {
	lt, lok := ext.GetType(left)
	rt, rok := ext.GetType(right)
	if !lok || !rok {
		return false, false
	}

	if lt == rt {
		return s.performStrictEqCmp(left, right)
	}

	if (lt == ext.NullType && rt == ext.UndefinedType) || (lt == ext.UndefinedType && rt == ext.NullType) {
		return true, true
	}
	if (lt == ext.NumberType && rt == ext.StringType) || rt == ext.BoolType {
		rv, rok := ext.AsPureNumber(right)
		if !rok {
			return false, false
		}
		return s.performAbstractEqCmp(left, &ast.Expression{Expr: &ast.NumberLiteral{Value: rv}})
	}
	if (lt == ext.StringType && rt == ext.NumberType) || lt == ext.BoolType {
		lv, lok := ext.AsPureNumber(left)
		if !lok {
			return false, false
		}
		return s.performAbstractEqCmp(&ast.Expression{Expr: &ast.NumberLiteral{Value: lv}}, right)
	}
	if (lt == ext.StringType && rt == ext.ObjectType) || (lt == ext.NumberType && rt == ext.ObjectType) ||
		(lt == ext.ObjectType && rt == ext.StringType) || (lt == ext.ObjectType && rt == ext.NumberType) {
		return false, false
	}

	return false, true
}

func (s *Simplifier) performStrictEqCmp(left, right *ast.Expression) (bool, bool) {
	// Any strict equality comparison against NaN returns false.
	if ext.IsNaN(left) || ext.IsNaN(right) {
		return false, true
	}
	// Special case, typeof a == typeof a is always true.
	if l, ok := left.Expr.(*ast.UnaryExpression); ok && l.Operator == token.Typeof {
		if r, ok := right.Expr.(*ast.UnaryExpression); ok && r.Operator == token.Typeof {
			if lid, lok := l.Operand.Expr.(*ast.Identifier); lok {
				if rid, rok := r.Operand.Expr.(*ast.Identifier); rok {
					if lid.ToId() == rid.ToId() {
						return true, true
					}
				}
			}
		}
	}
	lt, lok := ext.GetType(left)
	rt, rok := ext.GetType(right)
	if !lok || !rok {
		return false, false
	}
	// Strict equality can only be true for values of the same type.
	if lt != rt {
		return false, true
	}
	switch lt {
	case ext.UndefinedType, ext.NullType:
		return true, true
	case ext.NumberType:
		lv, lok := ext.AsPureNumber(left)
		rv, rok := ext.AsPureNumber(right)
		if !lok || !rok {
			return false, false
		}
		return lv == rv, true
	case ext.StringType:
		lv, lok := ext.AsPureString(left)
		rv, rok := ext.AsPureString(right)
		if !lok || !rok {
			return false, false
		}
		// In JS, browsers parse \v differently. So do not consider strings
		// equal if one contains \v.
		if strings.ContainsRune(lv, '\u000B') || strings.ContainsRune(rv, '\u000B') {
			return false, false
		}
		return lv == rv, true
	case ext.BoolType:
		lv, lok := ext.AsPureBool(left)
		rv, rok := ext.AsPureBool(right)
		// lv && rv || !lv && !rv
		andVal, andOk := ext.And(lv, rv, lok, rok)
		notAndVal, notAndOk := ext.And(!lv, !rv, lok, rok)
		return ext.Or(andVal, notAndVal, andOk, notAndOk)
	}

	return false, false
}

func (s *Simplifier) VisitAssignExpression(n *ast.AssignExpression) {
	old := s.isModifying
	s.isModifying = true
	n.Left.VisitWith(s)
	s.isModifying = old

	s.isModifying = false
	n.Right.VisitWith(s)
	s.isModifying = old
}

// This is overriden to preserve `this`.
func (s *Simplifier) VisitCallExpression(n *ast.CallExpression) {
	oldInCallee := s.inCallee

	s.inCallee = true
	mayInjectZero := !needZeroForThis(n.Callee)

	switch e := n.Callee.Expr.(type) {
	case *ast.SequenceExpression:
		if len(e.Sequence) == 1 {
			expr := e.Sequence[0]
			expr.VisitWith(s)
			n.Callee.Expr = expr.Expr
		} else if len(e.Sequence) > 0 && directnessMaters(&e.Sequence[len(e.Sequence)-1]) {
			first := e.Sequence[0]
			switch first.Expr.(type) {
			case *ast.NumberLiteral, *ast.Identifier:
			default:
				e.Sequence = append([]ast.Expression{{Expr: &ast.NumberLiteral{Value: 0.0}}}, e.Sequence...)
			}
			e.VisitWith(s)
		}
	default:
		e.VisitWith(s)
	}

	if mayInjectZero && needZeroForThis(n.Callee) {
		switch e := n.Callee.Expr.(type) {
		case *ast.SequenceExpression:
			e.Sequence = append([]ast.Expression{{Expr: &ast.NumberLiteral{Value: 0.0}}}, e.Sequence...)
		default:
			n.Callee.Expr = &ast.SequenceExpression{
				Sequence: []ast.Expression{
					{Expr: &ast.NumberLiteral{Value: 0.0}},
					{Expr: e},
				},
			}
		}
	}

	s.inCallee = false
	n.ArgumentList.VisitWith(s)

	s.inCallee = oldInCallee
}

func (s *Simplifier) VisitExpression(n *ast.Expression) {
	if unaryExpr, ok := n.Expr.(*ast.UnaryExpression); ok && unaryExpr.Operator == token.Delete {
		return
	}
	// fold children before doing something more.
	n.VisitChildrenWith(s)

	switch expr := n.Expr.(type) {
	// Do nothing.
	case *ast.StringLiteral, *ast.BooleanLiteral, *ast.NullLiteral, *ast.NumberLiteral, *ast.RegExpLiteral, *ast.ThisExpression:
		return
	case *ast.SequenceExpression:
		if len(expr.Sequence) == 0 {
			return
		}
	case *ast.UnaryExpression, *ast.BinaryExpression, *ast.MemberExpression, *ast.ConditionalExpression, *ast.ArrayLiteral, *ast.ObjectLiteral, *ast.NewExpression:
	default:
		return
	}

	switch expr := n.Expr.(type) {
	case *ast.UnaryExpression:
		s.optimizeUnaryExpression(n)
	case *ast.BinaryExpression:
		s.optimizeBinaryExpression(n)
	case *ast.MemberExpression:
		s.optimizeMemberExpression(n)
	case *ast.ConditionalExpression:
		v, ok, pure := ext.CastToBool(expr.Test)
		if ok {
			s.changed = true
			var val *ast.Expression
			if v {
				val = expr.Consequent
			} else {
				val = expr.Alternate
			}
			if pure {
				if directnessMaters(val) {
					n.Expr = &ast.SequenceExpression{
						Sequence: []ast.Expression{
							{Expr: &ast.NumberLiteral{Value: 0.0}},
							{Expr: val.Expr},
						},
					}
				} else {
					n.Expr = val
				}
			} else {
				n.Expr = &ast.SequenceExpression{
					Sequence: []ast.Expression{
						{Expr: expr.Test.Expr},
						{Expr: val.Expr},
					},
				}
			}
		}

	// Simplify sequence expression.
	case *ast.SequenceExpression:
		if len(expr.Sequence) == 1 {
			n.Expr = expr.Sequence[0].Expr
		}

	case *ast.ArrayLiteral:
		var exprs []ast.Expression
		for _, elem := range expr.Value {
			if arrLit, ok := elem.Expr.(*ast.ArrayLiteral); ok {
				s.changed = true
				exprs = append(exprs, arrLit.Value...)
			} else {
				exprs = append(exprs, elem)
			}
		}
		expr.Value = exprs

	case *ast.ObjectLiteral:
		// If the object has a spread property, we can't simplify it.
		if slices.ContainsFunc(expr.Value, func(e ast.Property) bool {
			_, ok := e.Prop.(*ast.SpreadElement)
			return ok
		}) {
			return
		}

		var props []ast.Property
		for _, prop := range expr.Value {
			if spread, ok := prop.Prop.(*ast.SpreadElement); ok {
				if obj, ok := spread.Expression.Expr.(*ast.ObjectLiteral); ok {
					s.changed = true
					props = append(props, obj.Value...)
				} else {
					props = append(props, prop)
				}
			} else {
				props = append(props, prop)
			}
		}
		expr.Value = props
	}
}

// Currently noop
func (s *Simplifier) VisitOptionalChain(n *ast.OptionalChain) {
}

func (s *Simplifier) VisitVariableDeclarator(n *ast.VariableDeclarator) {
	if n.Initializer != nil {
		if seqExpr, ok := n.Initializer.Expr.(*ast.SequenceExpression); ok {
			if len(seqExpr.Sequence) == 0 {
				n = nil
			}
		}
	}

	n.VisitChildrenWith(s)
}

// Drops unused values
func (s *Simplifier) VisitSequenceExpression(n *ast.SequenceExpression) {
	if len(n.Sequence) == 0 {
		return
	}

	oldInCallee := s.inCallee
	length := len(n.Sequence)
	for i := range n.Sequence {
		if i == length-1 {
			s.inCallee = oldInCallee
		} else {
			s.inCallee = false
		}
		n.Sequence[i].VisitWith(s)
	}
	s.inCallee = oldInCallee

	length = len(n.Sequence)
	last := n.Sequence[length-1]

	// Expressions except last one
	var exprs []ast.Expression
	for _, expr := range n.Sequence[:length-1] {
		if e, ok := expr.Expr.(*ast.NumberLiteral); ok && s.inCallee && e.Value == 0.0 {
			if len(exprs) == 0 {
				exprs = append(exprs, ast.Expression{Expr: &ast.NumberLiteral{Value: 0.0}})
			}
			continue
		}
		if s.inCallee && !ext.MayHaveSideEffects(&expr) {
			switch expr.Expr.(type) {
			case *ast.StringLiteral, *ast.BooleanLiteral, *ast.NullLiteral, *ast.NumberLiteral, *ast.RegExpLiteral, *ast.Identifier:
				if len(exprs) == 0 {
					s.changed = true
					exprs = append(exprs, ast.Expression{Expr: &ast.NumberLiteral{Value: 0.0}})
				}
				continue
			}
		}
		// Drop side-effect free nodes.
		switch expr.Expr.(type) {
		case *ast.StringLiteral, *ast.BooleanLiteral, *ast.NullLiteral, *ast.NumberLiteral, *ast.RegExpLiteral, *ast.Identifier:
			continue
		}
		// Flatten array
		if arrLit, ok := expr.Expr.(*ast.ArrayLiteral); ok {
			isSimple := !slices.ContainsFunc(arrLit.Value, func(e ast.Expression) bool {
				_, ok := e.Expr.(*ast.SpreadElement)
				return ok
			})
			if isSimple {
				exprs = append(exprs, arrLit.Value...)
			} else {
				exprs = append(exprs, ast.Expression{Expr: &ast.ArrayLiteral{Value: arrLit.Value}})
			}
			continue
		}
		// Default case: preserve it
		exprs = append(exprs, expr)
	}

	exprs = append(exprs, last)
	s.changed = s.changed || len(exprs) != len(n.Sequence)

	n.Sequence = exprs
}

func (s *Simplifier) VisitStatement(n *ast.Statement) {
	oldIsModifying := s.isModifying
	s.isModifying = false
	oldIsArgOfUpdate := s.isArgOfUpdate
	s.isArgOfUpdate = false
	n.VisitChildrenWith(s)
	s.isArgOfUpdate = oldIsArgOfUpdate
	s.isModifying = oldIsModifying
}

func (s *Simplifier) VisitUpdateExpression(n *ast.UpdateExpression) {
	old := s.isModifying
	s.isModifying = true
	n.Operand.VisitWith(s)
	s.isModifying = old
}

func (s *Simplifier) VisitForInStatement(n *ast.ForInStatement) {
	old := s.isModifying
	s.isModifying = true
	n.VisitChildrenWith(s)
	s.isModifying = old
}

func (s *Simplifier) VisitForOfStatement(n *ast.ForOfStatement) {
	old := s.isModifying
	s.isModifying = true
	n.VisitChildrenWith(s)
	s.isModifying = old
}

func (s *Simplifier) VisitTemplateLiteral(n *ast.TemplateLiteral) {
	if n.Tag != nil {
		old := s.inCallee
		s.inCallee = true

		n.Tag.VisitWith(s)

		s.inCallee = false
		n.Expressions.VisitWith(s)

		s.inCallee = old
	}
}

func (s *Simplifier) VisitWithStatement(n *ast.WithStatement) {
	n.Object.VisitWith(s)
}

// Simplify simplifies the AST by optimizing expressions.
// By default, it is expected that the AST is already resolved.
func Simplify(p *ast.Program, resolve bool) {
	if resolve {
		resolver.Resolve(p)
	}

	visitor := &Simplifier{}
	visitor.V = visitor
	p.VisitWith(visitor)
}
