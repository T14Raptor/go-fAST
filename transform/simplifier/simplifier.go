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
	switch prop := memExpr.Property.Prop.(type) {
	case *ast.Identifier:
		if _, ok := memExpr.Object.Expr.(*ast.ObjectLiteral); !ok && prop.Name == "length" {
			op = Len{}
		} else if s.inCallee {
			return
		} else {
			op = IndexStr(prop.Name)
		}
	case *ast.ComputedProperty:
		if s.inCallee {
			return
		}
		if numLit, ok := prop.Expr.Expr.(*ast.NumberLiteral); ok {
			// x[5]
			op = Index(numLit.Value)
		} else if s := ext.AsPureString(prop.Expr); s.Known() {
			if _, ok := memExpr.Object.Expr.(*ast.ObjectLiteral); !ok && s.Val() == "length" {
				// Length of non-object type
				op = Len{}
			} else if n, err := strconv.ParseFloat(s.Val(), 64); err == nil {
				// x['0'] is treated as x[0]
				op = Index(n)
			} else {
				// x[''] or x[...] where ... is an expression like [], ie x[[]]
				op = IndexStr(s.Val())
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
		expr.Expr = makeBoolExpr(v, []ast.Expression{{Expr: left.Expr}, {Expr: right.Expr}}).Expr
	}
	tryReplaceNum := func(v float64, left, right *ast.Expression) {
		s.changed = true
		var value ast.Expr
		if !math.IsNaN(v) {
			value = &ast.NumberLiteral{Value: v}
		} else {
			value = &ast.Identifier{Name: "NaN"}
		}
		expr.Expr = ext.PreserveEffects(ast.Expression{Expr: value}, []ast.Expression{{Expr: left.Expr}, {Expr: right.Expr}}).Expr
	}

	switch binExpr.Operator {
	case token.Plus:
		// It's string concatenation if either left or right is string.
		if ext.IsString(binExpr.Left) || ext.IsArrayLiteral(binExpr.Left) || ext.IsString(binExpr.Right) || ext.IsArrayLiteral(binExpr.Right) {
			l := ext.AsPureString(binExpr.Left)
			r := ext.AsPureString(binExpr.Right)
			if l.Known() && r.Known() {
				s.changed = true
				expr.Expr = &ast.StringLiteral{Value: l.Val() + r.Val()}
			}
		}

		typ := ext.GetType(expr)
		if typ.Unknown() {
			return
		}
		switch typ.Val().(type) {
		// String concatenation
		case ext.StringType:
			if !ext.MayHaveSideEffects(binExpr.Left) && !ext.MayHaveSideEffects(binExpr.Right) {
				l := ext.AsPureString(binExpr.Left)
				r := ext.AsPureString(binExpr.Right)
				if l.Known() && r.Known() {
					s.changed = true
					expr.Expr = &ast.StringLiteral{Value: l.Val() + r.Val()}
				}
			}
		// Numerical calculation
		case ext.BoolType, ext.NullType, ext.NumberType, ext.UndefinedType:
			if v := s.performArithmeticOp(token.Plus, binExpr.Left, binExpr.Right); v.Known() {
				tryReplaceNum(v.Val(), binExpr.Left, binExpr.Right)
			}
		}

		//TODO: try string concat

	case token.LogicalAnd, token.LogicalOr:
		val, _ := ext.CastToBool(binExpr.Left)
		if val.Known() {
			var node ast.Expression
			if binExpr.Operator == token.LogicalAnd {
				if val.Val() {
					// 1 && $right
					node = *binExpr.Right
				} else {
					s.changed = true
					// 0 && $right
					expr.Expr = binExpr.Left.Expr
				}
			} else {
				if val.Val() {
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
			expr.Expr = makeBoolExpr(false, []ast.Expression{{Expr: binExpr.Right.Expr}}).Expr
			return
		}
		if isObj(binExpr.Left) && ext.IsGlobalRefTo(binExpr.Right, "Object") {
			s.changed = true
			expr.Expr = makeBoolExpr(true, []ast.Expression{{Expr: binExpr.Left.Expr}}).Expr
		}
	case token.Minus, token.Slash, token.Remainder, token.Exponent:
		if v := s.performArithmeticOp(binExpr.Operator, binExpr.Left, binExpr.Right); v.Known() {
			tryReplaceNum(v.Val(), binExpr.Left, binExpr.Right)
		}
	case token.ShiftLeft, token.ShiftRight, token.UnsignedShiftRight:
		tryFoldShift := func(op token.Token, left, right *ast.Expression) (float64, bool) {
			if _, ok := left.Expr.(*ast.NumberLiteral); !ok {
				return 0, false
			}
			if _, ok := right.Expr.(*ast.NumberLiteral); !ok {
				return 0, false
			}

			lv := ext.AsPureNumber(left)
			rv := ext.AsPureNumber(right)
			if lv.Unknown() || rv.Unknown() {
				return 0, false
			}
			// Shift
			// (Masking of 0x1f is to restrict the shift to a maximum of 31 places)
			switch op {
			case token.ShiftLeft:
				return float64(int32(int32(lv.Val()) << uint32(rv.Val()) & 0x1f)), true
			case token.ShiftRight:
				return float64(int32(int32(lv.Val())>>uint32(rv.Val())) & 0x1f), true
			case token.UnsignedShiftRight:
				return float64(uint32(uint32(lv.Val()) >> uint32(rv.Val()) & 0x1f)), true
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
		if v := s.performArithmeticOp(binExpr.Operator, binExpr.Left, binExpr.Right); v.Known() {
			tryReplaceNum(v.Val(), binExpr.Left, binExpr.Right)
		}

		// Try left.rhs * right
		if binExpr2, ok := binExpr.Left.Expr.(*ast.BinaryExpression); ok && binExpr2.Operator == binExpr.Operator {
			if v := s.performArithmeticOp(binExpr.Operator, binExpr2.Left, binExpr.Right); v.Known() {
				var valExpr ast.Expr
				if !math.IsNaN(v.Val()) {
					valExpr = &ast.NumberLiteral{Value: v.Val()}
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
		if v := s.performAbstractRelCmp(binExpr.Left, binExpr.Right, false); v.Known() {
			tryReplaceBool(v.Val(), binExpr.Left, binExpr.Right)
		}
	case token.Greater:
		if v := s.performAbstractRelCmp(binExpr.Right, binExpr.Left, false); v.Known() {
			tryReplaceBool(v.Val(), binExpr.Right, binExpr.Left)
		}
	case token.LessOrEqual:
		if v := s.performAbstractRelCmp(binExpr.Right, binExpr.Left, true); v.Known() {
			tryReplaceBool(v.Val(), binExpr.Right, binExpr.Left)
		}
	case token.GreaterOrEqual:
		if v := s.performAbstractRelCmp(binExpr.Left, binExpr.Right, true); v.Known() {
			tryReplaceBool(v.Val(), binExpr.Left, binExpr.Right)
		}
	case token.Equal:
		if v := s.performAbstractEqCmp(binExpr.Left, binExpr.Right); v.Known() {
			tryReplaceBool(v.Val(), binExpr.Left, binExpr.Right)
		}
	case token.NotEqual:
		if v := s.performAbstractEqCmp(binExpr.Left, binExpr.Right); v.Known() {
			tryReplaceBool(!v.Val(), binExpr.Left, binExpr.Right)
		}
	case token.StrictEqual:
		if v := s.performStrictEqCmp(binExpr.Left, binExpr.Right); v.Known() {
			tryReplaceBool(v.Val(), binExpr.Left, binExpr.Right)
		}
	case token.StrictNotEqual:
		if v := s.performStrictEqCmp(binExpr.Left, binExpr.Right); v.Known() {
			tryReplaceBool(!v.Val(), binExpr.Left, binExpr.Right)
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
		if val, _ := ext.CastToBool(unaryExpr.Operand); val.Known() {
			s.changed = true
			expr.Expr = makeBoolExpr(val.Not().Val(), []ast.Expression{{Expr: unaryExpr.Operand.Expr}}).Expr
		}
	case token.Plus:
		if val := ext.AsPureNumber(unaryExpr.Operand); val.Known() {
			s.changed = true
			if math.IsNaN(val.Val()) {
				expr.Expr = ext.PreserveEffects(ast.Expression{Expr: &ast.Identifier{Name: "NaN"}}, []ast.Expression{{Expr: unaryExpr.Operand.Expr}}).Expr
				return
			}
			expr.Expr = ext.PreserveEffects(ast.Expression{Expr: &ast.NumberLiteral{Value: val.Val()}}, []ast.Expression{{Expr: unaryExpr.Operand.Expr}}).Expr
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
		if val := ext.AsPureNumber(unaryExpr.Operand); !val.Unknown() {
			if _, frac := math.Modf(val.Val()); frac == 0.0 {
				s.changed = true
				var result float64
				if val.Val() < 0.0 {
					result = float64(^uint32(int32(val.Val())))
				} else {
					result = float64(^uint32(val.Val()))
				}
				expr.Expr = &ast.NumberLiteral{Value: result}
			}
			// TODO: Report error
		}
	}
}

func (s *Simplifier) performArithmeticOp(op token.Token, left, right *ast.Expression) ext.Value[float64] {
	lv := ext.AsPureNumber(left)
	rv := ext.AsPureNumber(right)

	if (lv.Unknown() && rv.Unknown()) || op == token.Plus &&
		(!ext.GetType(left).CastToNumberOnAdd() || !ext.GetType(right).CastToNumberOnAdd()) {
		return ext.Unknown[float64]()
	}

	switch op {
	case token.Plus:
		if lv.Known() && rv.Known() {
			return ext.Known(lv.Val() + rv.Val())
		}
		if lv == ext.Known(0.0) {
			return rv
		} else if rv == ext.Known(0.0) {
			return lv
		}
		return ext.Unknown[float64]()
	case token.Minus:
		if lv.Known() && rv.Known() {
			return ext.Known(lv.Val() - rv.Val())
		}

		// 0 - x => -x
		if lv == ext.Known(0.0) {
			return ext.Known(-rv.Val())
		}

		// x - 0 => x
		if rv == ext.Known(0.0) {
			return lv
		}
		return ext.Unknown[float64]()
	case token.Multiply:
		if lv.Known() && rv.Known() {
			return ext.Known(lv.Val() * rv.Val())
		}
		// NOTE: 0*x != 0 for all x, if x==0, then it is NaN.  So we can't take
		// advantage of that without some kind of non-NaN proof.  So the special cases
		// here only deal with 1*x
		if lv == ext.Known(1.0) {
			return rv
		}
		if rv == ext.Known(1.0) {
			return lv
		}
		return ext.Unknown[float64]()
	case token.Slash:
		if lv.Known() && rv.Known() {
			if rv.Val() == 0.0 {
				return ext.Unknown[float64]()
			}
			return ext.Known(lv.Val() / rv.Val())
		}
		// NOTE: 0/x != 0 for all x, if x==0, then it is NaN
		if rv == ext.Known(1.0) {
			// TODO: cloneTree
			// x/1->x
			return lv
		}
		return ext.Unknown[float64]()
	case token.Exponent:
		if rv == ext.Known(0.0) {
			return ext.Known(1.0)
		}
		if lv.Known() && rv.Known() {
			return ext.Known(math.Pow(lv.Val(), rv.Val()))
		}
		return ext.Unknown[float64]()
	}

	if lv.Unknown() || rv.Unknown() {
		return ext.Unknown[float64]()
	}

	switch op {
	case token.And:
		return ext.Known(float64(int32(lv.Val()) & int32(rv.Val())))
	case token.Or:
		return ext.Known(float64(int32(lv.Val()) | int32(rv.Val())))
	case token.ExclusiveOr:
		return ext.Known(float64(int32(lv.Val()) ^ int32(rv.Val())))
	case token.Remainder:
		if rv.Val() == 0.0 {
			return ext.Unknown[float64]()
		}
		return ext.Known(math.Mod(lv.Val(), rv.Val()))
	}
	return ext.Unknown[float64]()
}

func (s *Simplifier) performAbstractRelCmp(left, right *ast.Expression, willNegate bool) ext.BoolValue {
	// Special case: `x < x` is always false.
	if l, ok := left.Expr.(*ast.Identifier); ok {
		if r, ok := right.Expr.(*ast.Identifier); ok {
			if !willNegate && l.Name == r.Name && l.ScopeContext == r.ScopeContext {
				return ext.BoolValue{Value: ext.Known(false)}
			}
		}
	}
	// Special case: `typeof a < typeof a` is always false.
	if l, ok := left.Expr.(*ast.UnaryExpression); ok && l.Operator == token.Typeof {
		if r, ok := right.Expr.(*ast.UnaryExpression); ok && r.Operator == token.Typeof {
			if lid, lok := l.Operand.Expr.(*ast.Identifier); lok {
				if rid, rok := r.Operand.Expr.(*ast.Identifier); rok {
					if lid.ToId() == rid.ToId() {
						return ext.BoolValue{Value: ext.Known(false)}
					}
				}
			}
		}
	}

	// Try to evaluate based on the general type.
	lt := ext.GetType(left)
	rt := ext.GetType(right)
	if lt.Value == ext.Known[ext.Type](ext.StringType{}) && rt.Value == ext.Known[ext.Type](ext.StringType{}) {
		lv := ext.AsPureString(left)
		rv := ext.AsPureString(right)
		if lv.Known() && rv.Known() {
			// In JS, browsers parse \v differently. So do not compare strings if one
			// contains \v.
			if strings.ContainsRune(lv.Val(), '\u000B') || strings.ContainsRune(rv.Val(), '\u000B') {
				return ext.BoolValue{Value: ext.Unknown[bool]()}
			} else {
				return ext.BoolValue{Value: ext.Known(lv.Val() < rv.Val())}
			}
		}
	}

	// Then, try to evaluate based on the value of the node. Try comparing as
	// numbers.
	lv := ext.AsPureNumber(left)
	rv := ext.AsPureNumber(right)
	if lv.Known() && rv.Known() {
		if math.IsNaN(lv.Val()) || math.IsNaN(rv.Val()) {
			return ext.BoolValue{Value: ext.Known(willNegate)}
		}
		return ext.BoolValue{Value: ext.Known(lv.Val() < rv.Val())}
	}

	return ext.BoolValue{Value: ext.Unknown[bool]()}
}

func (s *Simplifier) performAbstractEqCmp(left, right *ast.Expression) ext.BoolValue {
	lt := ext.GetType(left)
	rt := ext.GetType(right)
	if lt.Unknown() || rt.Unknown() {
		return ext.BoolValue{Value: ext.Unknown[bool]()}
	}

	if lt.Val() == rt.Val() {
		return s.performStrictEqCmp(left, right)
	}

	if (lt.Val() == ext.NullType{} && rt.Val() == ext.UndefinedType{}) || (lt.Val() == ext.UndefinedType{} && rt.Val() == ext.NullType{}) {
		return ext.BoolValue{Value: ext.Known(true)}
	}
	if (lt.Val() == ext.NumberType{} && rt.Val() == ext.StringType{}) || (rt.Val() == ext.BoolType{}) {
		rv := ext.AsPureNumber(right)
		if rv.Unknown() {
			return ext.BoolValue{Value: ext.Unknown[bool]()}
		}
		return s.performAbstractEqCmp(left, &ast.Expression{Expr: &ast.NumberLiteral{Value: rv.Val()}})
	}
	if (lt.Val() == ext.StringType{} && rt.Val() == ext.NumberType{}) || lt.Val() == (ext.BoolType{}) {
		lv := ext.AsPureNumber(left)
		if lv.Unknown() {
			return ext.BoolValue{Value: ext.Unknown[bool]()}
		}
		return s.performAbstractEqCmp(&ast.Expression{Expr: &ast.NumberLiteral{Value: lv.Val()}}, right)
	}
	if (lt.Val() == ext.StringType{} && rt.Val() == ext.ObjectType{}) || (lt.Val() == ext.NumberType{} && rt.Val() == ext.ObjectType{}) ||
		(lt.Val() == ext.ObjectType{} && rt.Val() == ext.StringType{}) || (lt.Val() == ext.ObjectType{} && rt.Val() == ext.NumberType{}) {
		return ext.BoolValue{Value: ext.Unknown[bool]()}
	}

	return ext.BoolValue{Value: ext.Known(false)}
}

func (s *Simplifier) performStrictEqCmp(left, right *ast.Expression) ext.BoolValue {
	// Any strict equality comparison against NaN returns false.
	if ext.IsNaN(left) || ext.IsNaN(right) {
		return ext.BoolValue{Value: ext.Known(false)}
	}
	// Special case, typeof a == typeof a is always true.
	if l, ok := left.Expr.(*ast.UnaryExpression); ok && l.Operator == token.Typeof {
		if r, ok := right.Expr.(*ast.UnaryExpression); ok && r.Operator == token.Typeof {
			if lid, lok := l.Operand.Expr.(*ast.Identifier); lok {
				if rid, rok := r.Operand.Expr.(*ast.Identifier); rok {
					if lid.ToId() == rid.ToId() {
						return ext.BoolValue{Value: ext.Known(true)}
					}
				}
			}
		}
	}
	lt := ext.GetType(left)
	rt := ext.GetType(right)
	if lt.Unknown() || rt.Unknown() {
		return ext.BoolValue{Value: ext.Unknown[bool]()}
	}
	// Strict equality can only be true for values of the same type.
	if lt.Val() != rt.Val() {
		return ext.BoolValue{Value: ext.Known(false)}
	}
	switch lt.Val().(type) {
	case ext.UndefinedType, ext.NullType:
		return ext.BoolValue{Value: ext.Known(true)}
	case ext.NumberType:
		lv := ext.AsPureNumber(left)
		rv := ext.AsPureNumber(right)
		if lv.Unknown() || rv.Unknown() {
			return ext.BoolValue{Value: ext.Unknown[bool]()}
		}
		return ext.BoolValue{Value: ext.Known(lv.Val() == rv.Val())}
	case ext.StringType:
		lv := ext.AsPureString(left)
		rv := ext.AsPureString(right)
		if lv.Unknown() || rv.Unknown() {
			return ext.BoolValue{Value: ext.Unknown[bool]()}
		}
		// In JS, browsers parse \v differently. So do not consider strings
		// equal if one contains \v.
		if strings.ContainsRune(lv.Val(), '\u000B') || strings.ContainsRune(rv.Val(), '\u000B') {
			return ext.BoolValue{Value: ext.Unknown[bool]()}
		}
		return ext.BoolValue{Value: ext.Known(lv.Val() == rv.Val())}
	case ext.BoolType:
		lv := ext.AsPureBool(left)
		rv := ext.AsPureBool(right)
		// lv && rv || !lv && !rv
		return lv.And(rv).Or(lv.Not().And(rv.Not()))
	}

	return ext.BoolValue{Value: ext.Unknown[bool]()}
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
		if v, pure := ext.CastToBool(expr.Test); v.Known() {
			s.changed = true
			var val *ast.Expression
			if v.Val() {
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
					n.Expr = val.Expr
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
