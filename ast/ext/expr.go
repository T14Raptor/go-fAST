package ext

import (
	"fmt"
	"github.com/t14raptor/go-fast/resolver"
	"math"
	"slices"
	"strings"

	"github.com/nukilabs/ftoa"
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

// IsString returns true if the expression is a potential string value.
func IsString(n *ast.Expression) bool {
	switch n.Kind() {
	case ast.ExprStrLit, ast.ExprTmplLit:
		return true
	case ast.ExprUnary:
		if n.MustUnary().Operator == token.Typeof {
			return true
		}
	case ast.ExprBinary:
		e := n.MustBinary()
		if e.Operator == token.Plus {
			return IsString(e.Left) || IsString(e.Right)
		}
	case ast.ExprAssign:
		e := n.MustAssign()
		if e.Operator == token.Assign || e.Operator == token.AddAssign {
			return IsString(e.Right)
		}
	case ast.ExprSequence:
		e := n.MustSequence()
		if len(e.Sequence) == 0 {
			return false
		}
		return IsString(&e.Sequence[len(e.Sequence)-1])
	case ast.ExprConditional:
		e := n.MustConditional()
		return IsString(e.Consequent) && IsString(e.Alternate)
	}
	return false
}

// IsArrayLiteral returns true if the expression is an array literal.
func IsArrayLiteral(n *ast.Expression) bool {
	return n.IsArrLit()
}

// IsNaN returns true if expr is a global reference to NaN.
func IsNaN(expr *ast.Expression) bool {
	return IsGlobalRefTo(expr, "NaN")
}

// IsUndefined returns true if expr is a global reference to undefined.
func IsUndefined(expr *ast.Expression) bool {
	return IsGlobalRefTo(expr, "undefined")
}

// IsVoid returns true if expr is a void operator.
func IsVoid(expr *ast.Expression) bool {
	if e, ok := expr.Unary(); ok {
		return e.Operator == token.Void
	}
	return false
}

// IsGlobalRefTo returns true if id references a global object.
func IsGlobalRefTo(expr *ast.Expression, id string) bool {
	if ident, ok := expr.Ident(); ok {
		return ident.Name == id && ident.ScopeContext == resolver.TopLevelMark
	}
	return false
}

// AsPureBool gets the boolean value if it does not have any side effects.
func AsPureBool(expr *ast.Expression) BoolValue {
	if v, pure := CastToBool(expr); pure {
		return v
	}
	return BoolValue{Unknown[bool]()}
}

// CastToBool emulates the Boolean() JavaScript cast function.
func CastToBool(expr *ast.Expression) (value BoolValue, pure bool) {
	if IsGlobalRefTo(expr, "undefined") || IsNaN(expr) {
		return BoolValue{}, true
	}

	switch expr.Kind() {
	case ast.ExprAssign:
		e := expr.MustAssign()
		if e.Operator == token.Assign {
			v, _ := CastToBool(e.Right)
			return v, false
		}
	case ast.ExprUnary:
		e := expr.MustUnary()
		switch e.Operator {
		case token.Minus:
			if n := AsPureNumber(e.Operand); n.Known() {
				value = BoolValue{Known(!(math.IsNaN(n.Val()) || n.Val() == 0))}
			} else {
				return BoolValue{Unknown[bool]()}, false
			}
		case token.Not:
			b, pure := CastToBool(e.Operand)
			return b.Not(), pure
		case token.Void:
			value = BoolValue{}
		}
	case ast.ExprSequence:
		e := expr.MustSequence()
		if len(e.Sequence) != 0 {
			value, _ = CastToBool(&e.Sequence[len(e.Sequence)-1])
		}
	case ast.ExprBinary:
		e := expr.MustBinary()
		switch e.Operator {
		case token.Minus:
			ln, lp := CastToNumber(e.Left)
			rn, rp := CastToNumber(e.Right)

			if ln.Known() && rn.Known() {
				return BoolValue{Known(ln.Val() != rn.Val())}, lp && rp
			}
			return BoolValue{Unknown[bool]()}, lp && rp
		case token.Slash:
			ln := AsPureNumber(e.Left)
			rn := AsPureNumber(e.Right)
			if ln.Known() && rn.Known() {
				// NaN is false
				if ln.Val() == 0.0 && rn.Val() == 0.0 {
					return BoolValue{Known(false)}, true
				}
				// Infinity is true
				if rn.Val() == 0.0 {
					return BoolValue{Known(true)}, true
				}
				v := ln.Val() / rn.Val()
				return BoolValue{Known(v != 0.0)}, true
			}
			value = BoolValue{Unknown[bool]()}
		case token.And, token.Or:
			if GetType(e.Left).Value != Known[Type](BoolType{}) || GetType(e.Right).Value != Known[Type](BoolType{}) {
				return BoolValue{Unknown[bool]()}, false
			}

			// TODO: Ignore purity if value cannot be reached.
			lv, lp := CastToBool(e.Left)
			rv, rp := CastToBool(e.Right)

			var v BoolValue
			if e.Operator == token.And {
				v = lv.And(rv)
			} else {
				v = lv.Or(rv)
			}
			if lp && rp {
				return v, true
			}
			value = v
		case token.LogicalOr:
			lv, lp := CastToBool(e.Left)
			if lv.Value == Known(true) {
				return lv, lp
			}
			rv, rp := CastToBool(e.Right)
			if rv.Value == Known(true) {
				return rv, lp && rp
			}
			value = BoolValue{Unknown[bool]()}
		case token.LogicalAnd:
			lv, lp := CastToBool(e.Left)
			if lv.Value == Known(false) {
				return lv, lp
			}
			rv, rp := CastToBool(e.Right)
			if rv.Value == Known(false) {
				return rv, lp && rp
			}
			value = BoolValue{Unknown[bool]()}
		case token.Plus:
			if s, ok := e.Left.StrLit(); ok && s.Value != "" {
				return BoolValue{Known(true)}, false
			}
			if s, ok := e.Right.StrLit(); ok && s.Value != "" {
				return BoolValue{Known(true)}, false
			}
			value = BoolValue{Unknown[bool]()}
		default:
			value = BoolValue{Unknown[bool]()}
		}
	case ast.ExprFuncLit, ast.ExprClassLit, ast.ExprNew, ast.ExprArrLit, ast.ExprObjLit:
		value = BoolValue{Known(true)}
	case ast.ExprNumLit:
		e := expr.MustNumLit()
		if e.Value == 0.0 || math.IsNaN(e.Value) {
			return BoolValue{}, true
		}
		return BoolValue{Known(true)}, true
	case ast.ExprBoolLit:
		return BoolValue{Known(expr.MustBoolLit().Value)}, true
	case ast.ExprStrLit:
		return BoolValue{Known(expr.MustStrLit().Value != "")}, true
	case ast.ExprNullLit:
		return BoolValue{}, true
	case ast.ExprRegExpLit:
		return BoolValue{Known(true)}, true
	default:
		value = BoolValue{Unknown[bool]()}
	}

	if MayHaveSideEffects(expr) {
		return value, false
	} else {
		return value, true
	}
}

// AsPureNumber gets the number value if it does not have any side effects.
func AsPureNumber(expr *ast.Expression) Value[float64] {
	if v, pure := CastToNumber(expr); pure {
		return v
	}
	return Unknown[float64]()
}

// CastToNumber emulates the Number() JavaScript cast function.
func CastToNumber(expr *ast.Expression) (value Value[float64], pure bool) {
	switch expr.Kind() {
	case ast.ExprBoolLit:
		e := expr.MustBoolLit()
		if e.Value {
			return Known(1.0), true
		}
		return Known(0.0), true
	case ast.ExprNumLit:
		return Known(expr.MustNumLit().Value), true
	case ast.ExprStrLit:
		return numFromStr(expr.MustStrLit().Value), true
	case ast.ExprNullLit:
		return Known(0.0), true
	case ast.ExprArrLit:
		s := AsPureString(expr)
		if s.Unknown() {
			return Unknown[float64](), false
		}
		return numFromStr(s.Val()), true
	case ast.ExprIdent:
		e := expr.MustIdent()
		if e.Name == "undefined" || e.Name == "NaN" && e.ScopeContext == resolver.TopLevelMark {
			return Known(math.NaN()), true
		}
		if e.Name == "Infinity" && e.ScopeContext == resolver.TopLevelMark {
			return Known(math.Inf(1)), true
		}
		return Unknown[float64](), true
	case ast.ExprUnary:
		e := expr.MustUnary()
		switch e.Operator {
		case token.Minus:
			if n, pure := CastToNumber(e.Operand); n.Known() && pure {
				return Known(-n.val), true
			}
			return Unknown[float64](), false
		case token.Not:
			if b, pure := CastToBool(e.Operand); b.Known() && pure {
				if b.val {
					return Value[float64]{}, true
				}
				return Known(1.0), true
			}
			return Unknown[float64](), false
		case token.Void:
			if MayHaveSideEffects(e.Operand) {
				return Known(math.NaN()), false
			} else {
				return Known(math.NaN()), true
			}
		}
	case ast.ExprTmplLit:
		if s := AsPureString(expr); s.Known() {
			return numFromStr(s.Val()), true
		}
	case ast.ExprSequence:
		e := expr.MustSequence()
		if len(e.Sequence) != 0 {
			v, _ := CastToNumber(&e.Sequence[len(e.Sequence)-1])
			return v, false
		}
	}
	return Unknown[float64](), false
}

// AsPureString gets the string value if it does not have any side effects.
func AsPureString(expr *ast.Expression) Value[string] {
	objectToStr := func(name string) string {
		return fmt.Sprintf("[object %s]", name)
	}
	funcToStr := func(name string) string {
		return fmt.Sprintf("function %s() { [native code] }", name)
	}

	switch expr.Kind() {
	case ast.ExprStrLit:
		return Known(expr.MustStrLit().Value)
	case ast.ExprNumLit:
		e := expr.MustNumLit()
		if e.Value == 0.0 {
			return Known("0")
		}
		return Known(ftoa.FormatFloat(e.Value, 'g', -1, 64))
	case ast.ExprBoolLit:
		return Known(fmt.Sprint(expr.MustBoolLit().Value))
	case ast.ExprNullLit:
		return Known("null")
	case ast.ExprTmplLit:
		// TODO:
		// Only convert a template literal if all its expressions can be
		// converted.
	case ast.ExprIdent:
		e := expr.MustIdent()
		switch e.Name {
		case "undefined", "Infinity", "NaN":
			return Known(e.Name)
		case "Math", "JSON":
			return Known(objectToStr(e.Name))
		case "Date":
			return Known(funcToStr(e.Name))
		}
	case ast.ExprUnary:
		e := expr.MustUnary()
		switch e.Operator {
		case token.Void:
			return Known("undefined")
		case token.Not:
			if b := AsPureBool(e.Operand); b.Known() {
				return Known(fmt.Sprint(!b.Val()))
			}
		}
	case ast.ExprArrLit:
		e := expr.MustArrLit()
		var sb strings.Builder
		// null, undefined is "" in array literal.
		for idx, elem := range e.Value {
			if idx > 0 {
				sb.WriteString(",")
			}
			switch elem.Kind() {
			case ast.ExprNullLit:
				sb.WriteString("")
			case ast.ExprUnary:
				ue := elem.MustUnary()
				if ue.Operator == token.Void {
					if MayHaveSideEffects(ue.Operand) {
						return Unknown[string]()
					}
					sb.WriteString("")
				}
			case ast.ExprIdent:
				if elem.MustIdent().Name == "undefined" {
					sb.WriteString("")
				}
			case ast.ExprNone:
				sb.WriteString("")
			default:
				if s := AsPureString(&elem); s.Known() {
					sb.WriteString(s.Val())
				} else {
					return Unknown[string]()
				}
			}
		}
		return Known(sb.String())
	case ast.ExprMember:
		e := expr.MustMember()
		var sym string
		switch e.Property.Kind() {
		case ast.MemPropIdent:
			sym = e.Property.MustIdent().Name
		case ast.MemPropComputed:
			if s, ok := e.Property.MustComputed().Expr.StrLit(); ok {
				sym = s.Value
			}
		default:
			return Unknown[string]()
		}
		// Convert some built-in funcs to string.
		switch e.Object.Kind() {
		case ast.ExprIdent:
			obj := e.Object.MustIdent()
			switch obj.Name {
			case "Math":
				if slices.Contains([]string{"abs", "acos", "acosh", "asin", "asinh", "atan", "atan2", "atanh", "cbrt", "ceil", "clz32", "cos", "cosh", "exp", "expm1", "floor", "fround", "hypot", "imul", "log", "log10", "log1p", "log2", "max", "min", "pow", "random", "round", "sign", "sin", "sinh", "sqrt", "tan", "tanh", "trunc"}, sym) {
					return Known(funcToStr(sym))
				}
			case "JSON":
				if slices.Contains([]string{"parse", "stringify"}, sym) {
					return Known(funcToStr(sym))
				}
			case "Date":
				if slices.Contains([]string{"now", "parse", "UTC"}, sym) {
					return Known(funcToStr(sym))
				}
			}
		case ast.ExprStrLit:
			if slices.Contains([]string{"anchor", "at", "big", "blink", "bold", "charAt", "charCodeAt", "codePointAt", "concat", "endsWith", "fixed", "fontcolor", "fontsize", "includes", "indexOf", "isWellFormed", "italics", "lastIndexOf", "link", "localeCompare", "match", "matchAll", "normalize", "padEnd", "padStart", "repeat", "replace", "replaceAll", "search", "slice", "small", "split", "startsWith", "strike", "sub", "substr", "substring", "sup", "toLocaleLowerCase", "toLocaleUpperCase", "toLowerCase", "toString", "toUpperCase", "toWellFormed", "trim", "trimEnd", "trimStart", "valueOf"}, sym) {
				return Known(funcToStr(sym))
			}
		case ast.ExprNumLit:
			if slices.Contains([]string{"toExponential", "toFixed", "toLocaleString", "toPrecision", "toString", "valueOf"}, sym) {
				return Known(funcToStr(sym))
			}
		case ast.ExprBoolLit:
			if slices.Contains([]string{"toString", "valueOf"}, sym) {
				return Known(funcToStr(sym))
			}
		case ast.ExprArrLit:
			if slices.Contains([]string{"at", "concat", "copyWithin", "entries", "every", "fill", "filter", "find", "findIndex", "findLast", "findLastIndex", "flat", "flatMap", "forEach", "includes", "indexOf", "join", "keys", "lastIndexOf", "map", "pop", "push", "reduce", "reduceRight", "reverse", "shift", "slice", "some", "sort", "splice", "toLocaleString", "toReversed", "toSorted", "toSpliced", "toString", "unshift", "values", "with"}, sym) {
				return Known(funcToStr(sym))
			}
		case ast.ExprObjLit:
			if slices.Contains([]string{"hasOwnProperty", "isPrototypeOf", "propertyIsEnumerable", "toLocaleString", "toString", "valueOf"}, sym) {
				return Known(funcToStr(sym))
			}
		}
	}
	return Unknown[string]()
}

// GetType returns the type of the expression.
func GetType(expr *ast.Expression) TypeValue {
	switch expr.Kind() {
	case ast.ExprAssign:
		e := expr.MustAssign()
		switch e.Operator {
		case token.Assign:
			return GetType(e.Right)
		case token.AddAssign:
			if rt := GetType(e.Right); !rt.Unknown() && (rt.Val() == StringType{}) {
				return TypeValue{Known[Type](StringType{})}
			}
		case token.AndAssign, token.ExclusiveOrAssign, token.OrAssign,
			token.ShiftLeftAssign, token.ShiftRightAssign, token.UnsignedShiftRightAssign,
			token.SubtractAssign, token.MultiplyAssign, token.ExponentAssign, token.QuotientAssign, token.RemainderAssign:
			return TypeValue{Known[Type](NumberType{})}
		}
	case ast.ExprMember:
		e := expr.MustMember()
		if ident, ok := e.Property.Ident(); ok {
			if ident.Name == "length" {
				switch e.Object.Kind() {
				case ast.ExprArrLit, ast.ExprStrLit:
					return TypeValue{Known[Type](NumberType{})}
				case ast.ExprIdent:
					if e.Object.MustIdent().Name == "arguments" {
						return TypeValue{Known[Type](NumberType{})}
					}
				}
			}
		}
	case ast.ExprSequence:
		e := expr.MustSequence()
		if len(e.Sequence) != 0 {
			return GetType(&e.Sequence[len(e.Sequence)-1])
		}
	case ast.ExprBinary:
		e := expr.MustBinary()
		switch e.Operator {
		case token.LogicalAnd, token.LogicalOr:
			if lt, rt := GetType(e.Left), GetType(e.Right); !lt.Unknown() && !rt.Unknown() && lt == rt {
				return lt
			}
		case token.Plus:
			rt := GetType(e.Right)
			if !rt.Unknown() && (rt.Val() == StringType{}) {
				return TypeValue{Known[Type](StringType{})}
			}

			lt := GetType(e.Left)
			if !lt.Unknown() && (lt.Val() == StringType{}) {
				return TypeValue{Known[Type](StringType{})}
			}
			// There are some pretty weird cases for object types:
			//   {} + [] === "0"
			//   [] + {} ==== "[object Object]"
			if !rt.Unknown() && (rt.Val() == ObjectType{}) {
				return TypeValue{Unknown[Type]()}
			}
			if !lt.Unknown() && (lt.Val() == ObjectType{}) {
				return TypeValue{Unknown[Type]()}
			}
			if !rt.Unknown() && !lt.Unknown() && !mayBeStr(lt.Val()) && !mayBeStr(rt.Val()) {
				return TypeValue{Known[Type](NumberType{})}
			}
		case token.Or, token.ExclusiveOr, token.And, token.ShiftLeft, token.ShiftRight, token.UnsignedShiftRight,
			token.Minus, token.Multiply, token.Remainder, token.Slash, token.Exponent:
			return TypeValue{Known[Type](NumberType{})}
		case token.Equal, token.NotEqual, token.StrictEqual, token.StrictNotEqual, token.Less, token.LessOrEqual,
			token.Greater, token.GreaterOrEqual, token.In, token.InstanceOf:
			return TypeValue{Known[Type](BoolType{})}
		}
	case ast.ExprConditional:
		e := expr.MustConditional()
		ct := GetType(e.Consequent)
		at := GetType(e.Alternate)
		if ct == at {
			return ct
		}
		return TypeValue{Unknown[Type]()}
	case ast.ExprIdent:
		e := expr.MustIdent()
		switch e.Name {
		case "undefined":
			return TypeValue{Known[Type](UndefinedType{})}
		case "Infinity", "NaN":
			return TypeValue{Known[Type](NumberType{})}
		default:
			return TypeValue{Unknown[Type]()}
		}
	case ast.ExprNumLit:
		return TypeValue{Known[Type](NumberType{})}
	case ast.ExprUnary:
		e := expr.MustUnary()
		switch e.Operator {
		case token.Minus, token.Plus, token.BitwiseNot:
			return TypeValue{Known[Type](NumberType{})}
		case token.Not, token.Delete:
			return TypeValue{Known[Type](BoolType{})}
		case token.Typeof:
			return TypeValue{Known[Type](StringType{})}
		case token.Void:
			return TypeValue{Known[Type](UndefinedType{})}
		}
	case ast.ExprUpdate:
		e := expr.MustUpdate()
		switch e.Operator {
		case token.Increment, token.Decrement:
			return TypeValue{Known[Type](NumberType{})}
		}
	case ast.ExprBoolLit:
		return TypeValue{Known[Type](BoolType{})}
	case ast.ExprStrLit, ast.ExprTmplLit:
		return TypeValue{Known[Type](StringType{})}
	case ast.ExprNullLit:
		return TypeValue{Known[Type](NullType{})}
	case ast.ExprFuncLit, ast.ExprNew, ast.ExprArrLit, ast.ExprObjLit, ast.ExprRegExpLit:
		return TypeValue{Known[Type](ObjectType{})}
	}
	return TypeValue{Unknown[Type]()}
}

// IsPureCallee returns true if the expression is a pure function.
func IsPureCallee(expr *ast.Expression) bool {
	if IsGlobalRefTo(expr, "Date") {
		return true
	}
	switch expr.Kind() {
	case ast.ExprMember:
		e := expr.MustMember()
		if IsGlobalRefTo(e.Object, "Math") {
			return true
		}
		// Some methods of string are pure
		if ident, ok := e.Property.Ident(); ok {
			if slices.Contains([]string{"charAt", "charCodeAt", "concat", "endsWith",
				"includes", "indexOf", "lastIndexOf", "localeCompare", "slice", "split",
				"startsWith", "substr", "substring", "toLocaleLowerCase", "toLocaleUpperCase",
				"toLowerCase", "toString", "toUpperCase", "trim", "trimEnd", "trimStart"}, ident.Name) {
				return true
			}
		}
	case ast.ExprFuncLit:
		e := expr.MustFuncLit()
		all := true
		for _, decl := range e.ParameterList.List {
			_, ok := decl.Target.Ident()
			if !ok {
				all = false
				break
			}
			if decl.Initializer != nil {
				all = false
				break
			}
		}
		if all && len(e.Body.List) == 0 {
			return true
		}
	}
	return false
}

// MayHaveSideEffects returns true if the expression may have side effects.
func MayHaveSideEffects(expr *ast.Expression) bool {
	if expr == nil || expr.IsNone() {
		return false
	}
	if IsPureCallee(expr) {
		return false
	}
	switch expr.Kind() {
	case ast.ExprIdent:
		e := expr.MustIdent()
		if e.ScopeContext == resolver.UnresolvedMark &&
			!slices.Contains([]string{"Infinity", "NaN", "Math", "undefined",
				"Object", "Array", "Promise", "Boolean", "Number", "String",
				"BigInt", "Error", "RegExp", "Function", "document"}, e.Name) {
			return true
		}
		return false
	case ast.ExprStrLit, ast.ExprNumLit, ast.ExprBoolLit, ast.ExprNullLit, ast.ExprRegExpLit:
		return false
	// Function expression does not have any side effect if it's not used.
	case ast.ExprFuncLit, ast.ExprArrowFuncLit:
		return false
	case ast.ExprClassLit:
		return classHasSideEffect(expr.MustClassLit())
	case ast.ExprArrLit:
		e := expr.MustArrLit()
		for _, elem := range e.Value {
			if MayHaveSideEffects(&elem) {
				return true
			}
		}
		return false
	case ast.ExprUnary:
		e := expr.MustUnary()
		if e.Operator == token.Delete {
			return true
		}
		return MayHaveSideEffects(e.Operand)
	case ast.ExprBinary:
		e := expr.MustBinary()
		return MayHaveSideEffects(e.Left) || MayHaveSideEffects(e.Right)
	case ast.ExprMember:
		e := expr.MustMember()
		switch e.Object.Kind() {
		case ast.ExprObjLit, ast.ExprFuncLit, ast.ExprArrowFuncLit, ast.ExprClassLit:
			if MayHaveSideEffects(e.Object) {
				return true
			}
			switch e.Object.Kind() {
			case ast.ExprClassLit:
				obj := e.Object.MustClassLit()
				for _, elem := range obj.Body {
					if method, ok := elem.Method(); ok && method.Static {
						if method.Kind == ast.PropertyKindGet || method.Kind == ast.PropertyKindSet {
							return true
						}
					}
				}
				return false
			case ast.ExprObjLit:
				obj := e.Object.MustObjLit()
				for _, prop := range obj.Value {
					switch prop.Kind() {
					case ast.PropSpread:
						return true
					case ast.PropShort:
						p := prop.MustShort()
						if p.Name.Name == "__proto__" {
							return true
						}
					case ast.PropKeyed:
						p := prop.MustKeyed()
						if s, ok := p.Key.StrLit(); ok && s.Value == "__proto__" {
							return true
						}
						if id, ok := p.Key.Ident(); ok && id.Name == "__proto__" {
							return true
						}
						if p.Computed {
							return true
						}
					}
				}
				return false
			}

			switch e.Property.Kind() {
			case ast.MemPropIdent:
				return false
			case ast.MemPropComputed:
				return MayHaveSideEffects(e.Property.MustComputed().Expr)
			}
		}

	case ast.ExprTmplLit:
	case ast.ExprMetaProp:
	case ast.ExprAwait, ast.ExprYield, ast.ExprSuper, ast.ExprUpdate, ast.ExprAssign:

	case ast.ExprNew:

	case ast.ExprOptChain:
		e := expr.MustOptChain()
		switch e.Base.Kind() {
		case ast.ExprMember:
		case ast.ExprCall:
			base := e.Base.MustCall()
			if IsPureCallee(base.Callee) {
				for _, arg := range base.ArgumentList {
					if MayHaveSideEffects(&arg) {
						return true
					}
				}
				return false
			}
		}
	case ast.ExprCall:
		e := expr.MustCall()
		if IsPureCallee(e.Callee) {
			for _, arg := range e.ArgumentList {
				if MayHaveSideEffects(&arg) {
					return true
				}
			}
			return false
		}
	case ast.ExprSequence:
		e := expr.MustSequence()
		for _, expr := range e.Sequence {
			if MayHaveSideEffects(&expr) {
				return true
			}
		}
		return false
	case ast.ExprConditional:
		e := expr.MustConditional()
		return MayHaveSideEffects(e.Test) || MayHaveSideEffects(e.Consequent) || MayHaveSideEffects(e.Alternate)
	case ast.ExprObjLit:
		e := expr.MustObjLit()
		for _, prop := range e.Value {
			switch prop.Kind() {
			case ast.PropSpread:
				return true
			case ast.PropShort:
			case ast.PropKeyed:
				p := prop.MustKeyed()
				if p.Computed && MayHaveSideEffects(p.Key) {
					return true
				}
				if MayHaveSideEffects(p.Value) {
					return true
				}
			}
		}
		return false
	case ast.ExprInvalid:
		return true
	}
	return true
}
