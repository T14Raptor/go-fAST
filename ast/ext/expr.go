package ext

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/nukilabs/ftoa"
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
	"github.com/t14raptor/go-fast/transform/resolver"
)

// IsString returns true if the expression is a potential string value.
func IsString(n *ast.Expression) bool {
	switch n := n.Expr.(type) {
	case *ast.StringLiteral, *ast.TemplateLiteral:
		return true
	case *ast.UnaryExpression:
		if n.Operator == token.Typeof {
			return true
		}
	case *ast.BinaryExpression:
		if n.Operator == token.Plus {
			return IsString(n.Left) || IsString(n.Right)
		}
	case *ast.AssignExpression:
		if n.Operator == token.Assign || n.Operator == token.AddAssign {
			return IsString(n.Right)
		}
	case *ast.SequenceExpression:
		if len(n.Sequence) == 0 {
			return false
		}
		return IsString(&n.Sequence[len(n.Sequence)-1])
	case *ast.ConditionalExpression:
		return IsString(n.Consequent) && IsString(n.Alternate)
	}
	return false
}

// IsArrayLiteral returns true if the expression is an array literal.
func IsArrayLiteral(n *ast.Expression) bool {
	_, ok := n.Expr.(*ast.ArrayLiteral)
	return ok
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
	if unary, ok := expr.Expr.(*ast.UnaryExpression); ok {
		return unary.Operator == token.Void
	}
	return false
}

// IsGlobalRefTo returns true if id references a global object.
func IsGlobalRefTo(expr *ast.Expression, id string) bool {
	if ident, ok := expr.Expr.(*ast.Identifier); ok {
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

	switch e := expr.Expr.(type) {
	case *ast.AssignExpression:
		if e.Operator == token.Assign {
			v, _ := CastToBool(e.Right)
			return v, false
		}
	case *ast.UnaryExpression:
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
	case *ast.SequenceExpression:
		if len(e.Sequence) != 0 {
			value, _ = CastToBool(&e.Sequence[len(e.Sequence)-1])
		}
	case *ast.BinaryExpression:
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
			if strLit, ok := e.Left.Expr.(*ast.StringLiteral); ok && strLit.Value != "" {
				return BoolValue{Known(true)}, false
			}
			if strLit, ok := e.Right.Expr.(*ast.StringLiteral); ok && strLit.Value != "" {
				return BoolValue{Known(true)}, false
			}
			value = BoolValue{Unknown[bool]()}
		default:
			value = BoolValue{Unknown[bool]()}
		}
	case *ast.FunctionLiteral, *ast.ClassLiteral, *ast.NewExpression, *ast.ArrayLiteral, *ast.ObjectLiteral:
		value = BoolValue{Known(true)}
	case *ast.NumberLiteral:
		if e.Value == 0.0 || math.IsNaN(e.Value) {
			return BoolValue{}, true
		}
		return BoolValue{Known(true)}, true
	case *ast.BooleanLiteral:
		return BoolValue{Known(e.Value)}, true
	case *ast.StringLiteral:
		return BoolValue{Known(e.Value != "")}, true
	case *ast.NullLiteral:
		return BoolValue{}, true
	case *ast.RegExpLiteral:
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
	switch e := expr.Expr.(type) {
	case *ast.BooleanLiteral:
		if e.Value {
			return Known(1.0), true
		}
		return Known(0.0), true
	case *ast.NumberLiteral:
		return Known(e.Value), true
	case *ast.StringLiteral:
		return numFromStr(e.Value), true
	case *ast.NullLiteral:
		return Known(0.0), true
	case *ast.ArrayLiteral:
		s := AsPureString(expr)
		if s.Unknown() {
			return Unknown[float64](), false
		}
		return numFromStr(s.Val()), true
	case *ast.Identifier:
		if e.Name == "undefined" || e.Name == "NaN" && e.ScopeContext == resolver.TopLevelMark {
			return Known(math.NaN()), true
		}
		if e.Name == "Infinity" && e.ScopeContext == resolver.TopLevelMark {
			return Known(math.Inf(1)), true
		}
		return Unknown[float64](), true
	case *ast.UnaryExpression:
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
	case *ast.TemplateLiteral:
		if s := AsPureString(expr); s.Known() {
			return numFromStr(s.Val()), true
		}
	case *ast.SequenceExpression:
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

	switch e := expr.Expr.(type) {
	case *ast.StringLiteral:
		return Known(e.Value)
	case *ast.NumberLiteral:
		if e.Value == 0.0 {
			return Known("0")
		}
		return Known(ftoa.FormatFloat(e.Value, 'g', -1, 64))
	case *ast.BooleanLiteral:
		return Known(fmt.Sprint(e.Value))
	case *ast.NullLiteral:
		return Known("null")
	case *ast.TemplateLiteral:
		// TODO:
		// Only convert a template literal if all its expressions can be
		// converted.
	case *ast.Identifier:
		switch e.Name {
		case "undefined", "Infinity", "NaN":
			return Known(e.Name)
		case "Math", "JSON":
			return Known(objectToStr(e.Name))
		case "Date":
			return Known(funcToStr(e.Name))
		}
	case *ast.UnaryExpression:
		switch e.Operator {
		case token.Void:
			return Known("undefined")
		case token.Not:
			if b := AsPureBool(e.Operand); b.Known() {
				return Known(fmt.Sprint(!b.Val()))
			}
		}
	case *ast.ArrayLiteral:
		var sb strings.Builder
		// null, undefined is "" in array literal.
		for idx, elem := range e.Value {
			if idx > 0 {
				sb.WriteString(",")
			}
			switch e := elem.Expr.(type) {
			case *ast.NullLiteral:
				sb.WriteString("")
			case *ast.UnaryExpression:
				if e.Operator == token.Void {
					if MayHaveSideEffects(e.Operand) {
						return Unknown[string]()
					}
					sb.WriteString("")
				}
			case *ast.Identifier:
				if e.Name == "undefined" {
					sb.WriteString("")
				}
			case nil:
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
	case *ast.MemberExpression:
		var sym string
		switch prop := e.Property.Prop.(type) {
		case *ast.Identifier:
			sym = prop.Name
		case *ast.ComputedProperty:
			if strLit, ok := prop.Expr.Expr.(*ast.StringLiteral); ok {
				sym = strLit.Value
			}
		default:
			return Unknown[string]()
		}
		// Convert some built-in funcs to string.
		switch obj := e.Object.Expr.(type) {
		case *ast.Identifier:
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
		case *ast.StringLiteral:
			if slices.Contains([]string{"anchor", "at", "big", "blink", "bold", "charAt", "charCodeAt", "codePointAt", "concat", "endsWith", "fixed", "fontcolor", "fontsize", "includes", "indexOf", "isWellFormed", "italics", "lastIndexOf", "link", "localeCompare", "match", "matchAll", "normalize", "padEnd", "padStart", "repeat", "replace", "replaceAll", "search", "slice", "small", "split", "startsWith", "strike", "sub", "substr", "substring", "sup", "toLocaleLowerCase", "toLocaleUpperCase", "toLowerCase", "toString", "toUpperCase", "toWellFormed", "trim", "trimEnd", "trimStart", "valueOf"}, sym) {
				return Known(funcToStr(sym))
			}
		case *ast.NumberLiteral:
			if slices.Contains([]string{"toExponential", "toFixed", "toLocaleString", "toPrecision", "toString", "valueOf"}, sym) {
				return Known(funcToStr(sym))
			}
		case *ast.BooleanLiteral:
			if slices.Contains([]string{"toString", "valueOf"}, sym) {
				return Known(funcToStr(sym))
			}
		case *ast.ArrayLiteral:
			if slices.Contains([]string{"at", "concat", "copyWithin", "entries", "every", "fill", "filter", "find", "findIndex", "findLast", "findLastIndex", "flat", "flatMap", "forEach", "includes", "indexOf", "join", "keys", "lastIndexOf", "map", "pop", "push", "reduce", "reduceRight", "reverse", "shift", "slice", "some", "sort", "splice", "toLocaleString", "toReversed", "toSorted", "toSpliced", "toString", "unshift", "values", "with"}, sym) {
				return Known(funcToStr(sym))
			}
		case *ast.ObjectLiteral:
			if slices.Contains([]string{"hasOwnProperty", "isPrototypeOf", "propertyIsEnumerable", "toLocaleString", "toString", "valueOf"}, sym) {
				return Known(funcToStr(sym))
			}
		}
	}
	return Unknown[string]()
}

// GetType returns the type of the expression.
func GetType(expr *ast.Expression) TypeValue {
	switch e := expr.Expr.(type) {
	case *ast.AssignExpression:
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
	case *ast.MemberExpression:
		if ident, ok := e.Property.Prop.(*ast.Identifier); ok {
			if ident.Name == "length" {
				switch obj := e.Object.Expr.(type) {
				case *ast.ArrayLiteral, *ast.StringLiteral:
					return TypeValue{Known[Type](NumberType{})}
				case *ast.Identifier:
					if obj.Name == "arguments" {
						return TypeValue{Known[Type](NumberType{})}
					}
				}
			}
		}
	case *ast.SequenceExpression:
		if len(e.Sequence) != 0 {
			return GetType(&e.Sequence[len(e.Sequence)-1])
		}
	case *ast.BinaryExpression:
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
	case *ast.ConditionalExpression:
		ct := GetType(e.Consequent)
		at := GetType(e.Alternate)
		if ct == at {
			return ct
		}
		return TypeValue{Unknown[Type]()}
	case *ast.Identifier:
		switch e.Name {
		case "undefined":
			return TypeValue{Known[Type](UndefinedType{})}
		case "Infinity", "NaN":
			return TypeValue{Known[Type](NumberType{})}
		default:
			return TypeValue{Unknown[Type]()}
		}
	case *ast.NumberLiteral:
		return TypeValue{Known[Type](NumberType{})}
	case *ast.UnaryExpression:
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
	case *ast.UpdateExpression:
		switch e.Operator {
		case token.Increment, token.Decrement:
			return TypeValue{Known[Type](NumberType{})}
		}
	case *ast.BooleanLiteral:
		return TypeValue{Known[Type](BoolType{})}
	case *ast.StringLiteral, *ast.TemplateLiteral:
		return TypeValue{Known[Type](StringType{})}
	case *ast.NullLiteral:
		return TypeValue{Known[Type](NullType{})}
	case *ast.FunctionLiteral, *ast.NewExpression, *ast.ArrayLiteral, *ast.ObjectLiteral, *ast.RegExpLiteral:
		return TypeValue{Known[Type](ObjectType{})}
	}
	return TypeValue{Unknown[Type]()}
}

// IsPureCallee returns true if the expression is a pure function.
func IsPureCallee(expr *ast.Expression) bool {
	if IsGlobalRefTo(expr, "Date") {
		return true
	}
	switch e := expr.Expr.(type) {
	case *ast.MemberExpression:
		if IsGlobalRefTo(e.Object, "Math") {
			return true
		}
		// Some methods of string are pure
		if ident, ok := e.Property.Prop.(*ast.Identifier); ok {
			if slices.Contains([]string{"charAt", "charCodeAt", "concat", "endsWith",
				"includes", "indexOf", "lastIndexOf", "localeCompare", "slice", "split",
				"startsWith", "substr", "substring", "toLocaleLowerCase", "toLocaleUpperCase",
				"toLowerCase", "toString", "toUpperCase", "trim", "trimEnd", "trimStart"}, ident.Name) {
				return true
			}
		}
	case *ast.FunctionLiteral:
		all := true
		for _, decl := range e.ParameterList.List {
			_, ok := decl.Target.Target.(*ast.Identifier)
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
	if IsPureCallee(expr) {
		return false
	}
	switch e := expr.Expr.(type) {
	case *ast.Identifier:
		if e.ScopeContext == resolver.UnresolvedMark &&
			!slices.Contains([]string{"Infinity", "NaN", "Math", "undefined",
				"Object", "Array", "Promise", "Boolean", "Number", "String",
				"BigInt", "Error", "RegExp", "Function", "document"}, e.Name) {
			return true
		}
		return false
	case *ast.StringLiteral, *ast.NumberLiteral, *ast.BooleanLiteral, *ast.NullLiteral, *ast.RegExpLiteral:
		return false
	// Function expression does not have any side effect if it's not used.
	case *ast.FunctionLiteral, *ast.ArrowFunctionLiteral:
		return false
	case *ast.ClassLiteral:
		return classHasSideEffect(e)
	case *ast.ArrayLiteral:
		for _, elem := range e.Value {
			if MayHaveSideEffects(&elem) {
				return true
			}
		}
		return false
	case *ast.UnaryExpression:
		if e.Operator == token.Delete {
			return true
		}
		return MayHaveSideEffects(e.Operand)
	case *ast.BinaryExpression:
		return MayHaveSideEffects(e.Left) || MayHaveSideEffects(e.Right)
	case *ast.MemberExpression:
		switch e.Object.Expr.(type) {
		case *ast.ObjectLiteral, *ast.FunctionLiteral, *ast.ArrowFunctionLiteral, *ast.ClassLiteral:
			if MayHaveSideEffects(e.Object) {
				return true
			}
			switch obj := e.Object.Expr.(type) {
			case *ast.ClassLiteral:
				for _, elem := range obj.Body {
					if elem, ok := elem.Element.(*ast.MethodDefinition); ok && elem.Static {
						if elem.Kind == ast.PropertyKindGet || elem.Kind == ast.PropertyKindSet {
							return true
						}
					}
				}
				return false
			case *ast.ObjectLiteral:
				for _, prop := range obj.Value {
					switch p := prop.Prop.(type) {
					case *ast.SpreadElement:
						return true
					case *ast.PropertyShort:
						if p.Name.Name == "__proto__" {
							return true
						}
					case *ast.PropertyKeyed:
						if strLit, ok := p.Key.Expr.(*ast.StringLiteral); ok && strLit.Value == "__proto__" {
							return true
						}
						if ident, ok := p.Key.Expr.(*ast.Identifier); ok && ident.Name == "__proto__" {
							return true
						}
						if p.Computed {
							return true
						}
					}
				}
				return false
			}

			switch prop := e.Property.Prop.(type) {
			case *ast.Identifier:
				return false
			case *ast.ComputedProperty:
				return MayHaveSideEffects(prop.Expr)
			}
		}

	case *ast.TemplateLiteral:
	case *ast.MetaProperty:
	case *ast.AwaitExpression, *ast.YieldExpression, *ast.SuperExpression, *ast.UpdateExpression, *ast.AssignExpression:

	case *ast.NewExpression:

	case *ast.OptionalChain:
		switch base := e.Base.Expr.(type) {
		case *ast.MemberExpression:
		case *ast.CallExpression:
			if IsPureCallee(base.Callee) {
				for _, arg := range base.ArgumentList {
					if MayHaveSideEffects(&arg) {
						return true
					}
				}
				return false
			}
		}
	case *ast.CallExpression:
		if IsPureCallee(e.Callee) {
			for _, arg := range e.ArgumentList {
				if MayHaveSideEffects(&arg) {
					return true
				}
			}
			return false
		}
	case *ast.SequenceExpression:
		for _, expr := range e.Sequence {
			if MayHaveSideEffects(&expr) {
				return true
			}
		}
		return false
	case *ast.ConditionalExpression:
		return MayHaveSideEffects(e.Test) || MayHaveSideEffects(e.Consequent) || MayHaveSideEffects(e.Alternate)
	case *ast.ObjectLiteral:
		for _, prop := range e.Value {
			switch p := prop.Prop.(type) {
			case *ast.SpreadElement:
				return true
			case *ast.PropertyShort:
			case *ast.PropertyKeyed:
				if p.Computed && MayHaveSideEffects(p.Key) {
					return true
				}
				if MayHaveSideEffects(p.Value) {
					return true
				}
			}
		}
		return false
	case *ast.InvalidExpression:
		return true
	case nil:
		return false
	}
	return true
}
