package simplifier

import (
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/t14raptor/go-fast/ast"
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
		if s, ok := asPureString(memExpr.Property); ok {
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
				return mayHaveSideEffects(&e)
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
				return mayHaveSideEffects(&e)
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
				extractSideEffectsTo(&exprs, &elem)
			}

			// Element value.
			val := v

			// Add after side effects.
			for _, elem := range after {
				extractSideEffectsTo(&exprs, &elem)
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
			if op != "yield" && isLiteral(&obj.Value) {
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
		expr.Expr = preserveEffects(ast.Expression{Expr: v}, []ast.Expression{{Expr: obj}}).Expr
	}
}

func nthChar(s string, idx int) (string, bool) {
	for _, c := range s {
		if len(utf16.Encode([]rune{c})) > 1 {
			return "", false
		}
	}

	if !strings.Contains(s, "\\ud") && !strings.Contains(s, "\\uD") {
		if idx < len([]rune(s)) {
			return string([]rune(s)[idx]), true
		}
		return "", false
	}

	iter := []rune(s)
	for i := 0; i < len(iter); i++ {
		c := iter[i]
		if c == '\\' && i+1 < len(iter) && iter[i+1] == 'u' {
			if idx == 0 {
				if i+5 < len(iter) {
					return string(iter[i : i+6]), true
				}
				return "", false
			}
			i += 5
		} else {
			if idx == 0 {
				return string(c), true
			}
		}
		idx--
	}

	return "", false
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
			if key != "__proto__" && propNameEq(prop.Key, "__proto__") {
				// If __proto__ is defined, we need to check the contents of it,
				// as well as any nested __proto__ objects
				if obj, ok := prop.Value.Expr.(*ast.ObjectLiteral); ok {
					if v := getKeyValue(obj.Value, key); v != nil {
						return v
					}
				}
				return nil
			} else if propNameEq(prop.Key, key) {
				return prop.Value
			}
		}
	}

	return nil
}

func Simplify(p *ast.Program) {
	resolver.Resolve(p)

	visitor := &Simplifier{}
	visitor.V = visitor
	p.VisitWith(visitor)
}
