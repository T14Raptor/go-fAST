package ext

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/resolver"
	"github.com/t14raptor/go-fast/token"
)

// IsNumber returns true if the expression is a number.
func IsNumber(n *ast.Expression) bool {
	_, ok := n.Expr.(*ast.NumberLiteral)
	return ok
}

// IsStr returns true if the expression is a string.
func IsStr(n *ast.Expression) bool {
	switch n := n.Expr.(type) {
	case *ast.StringLiteral, *ast.TemplateLiteral:
		return true
	case *ast.UnaryExpression:
		if n.Operator == token.Typeof {
			return true
		}
	case *ast.BinaryExpression:
		if n.Operator == token.Plus {
			return IsStr(n.Left) || IsStr(n.Right)
		}
	case *ast.AssignExpression:
		if n.Operator == token.Assign || n.Operator == token.AddAssign {
			return IsStr(n.Right)
		}
	case *ast.SequenceExpression:
		if len(n.Sequence) == 0 {
			return false
		}
		return IsStr(&n.Sequence[len(n.Sequence)-1])
	case *ast.ConditionalExpression:
		return IsStr(n.Consequent) && IsStr(n.Alternate)
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
		return ident.Name == id && ident.ScopeContext == resolver.UnresolvedMark
	}
	return false
}

// AsPureBool gets the boolean value if it does not have any side effects.
func AsPureBool(expr *ast.Expression) (value bool, ok bool) {
	if v, ok, pure := CastToBool(expr); pure {
		return v, ok
	}
	return false, false
}

// CastToBool emulates the Boolean() JavaScript cast function.
func CastToBool(expr *ast.Expression) (value bool, ok bool, pure bool) {
	if IsGlobalRefTo(expr, "undefined") {
		return false, true, true
	}
	if IsNaN(expr) {
		return false, true, true
	}

	switch e := expr.Expr.(type) {
	case *ast.AssignExpression:
		if e.Operator == token.Assign {
			v, ok, _ := CastToBool(e.Right)
			return v, ok, false
		}
	case *ast.UnaryExpression:
		switch e.Operator {
		case token.Minus:
			n, ok := AsPureNumber(e.Operand)
			if ok {
				value = !(math.IsNaN(n) || n == 0)
			} else {
				return false, false, false
			}
		case token.Not:
			b, ok, pure := CastToBool(e.Operand)
			return !b, ok, pure
		case token.Void:
			value = false
			ok = true
		}
	case *ast.SequenceExpression:
		if len(e.Sequence) != 0 {
			value, ok, _ = CastToBool(&e.Sequence[len(e.Sequence)-1])
		}
	case *ast.BinaryExpression:
		switch e.Operator {
		case token.Minus:
			nl, okl, pl := CastToNumber(e.Left)
			nr, okr, pr := CastToNumber(e.Right)

			return nl != nr, okl && okr, pl && pr
		case token.Slash:
			nl, okl := AsPureNumber(e.Left)
			nr, okr := AsPureNumber(e.Right)
			if okl && okr {
				// NaN is false
				if nl == 0.0 && nr == 0.0 {
					return false, true, true
				}
				// Infinity is true
				if nr == 0.0 {
					return true, true, true
				}
				n := nl / nr
				return n != 0.0, true, true
			}
		case token.And, token.Or:
			lt, lok := GetType(e.Left)
			rt, rok := GetType(e.Right)
			if lok && rok && lt != BoolType && rt != BoolType {
				return false, false, false
			}

			// TODO: Ignore purity if value cannot be reached.
			lv, lok, lp := CastToBool(e.Left)
			rv, rok, rp := CastToBool(e.Right)

			if e.Operator == token.And {
				value, ok = And(lv, rv, lok, rok)
			} else {
				value, ok = Or(lv, rv, lok, rok)
			}
			if lp && rp {
				return value, ok, true
			}
		case token.LogicalOr:
			lv, lok, lp := CastToBool(e.Left)
			if lok && lv {
				return lv, lok, lp
			}
			rv, rok, rp := CastToBool(e.Right)
			if rok && rv {
				return rv, rok, rp
			}
		case token.LogicalAnd:
			lv, lok, lp := CastToBool(e.Left)
			if lok && !lv {
				return lv, lok, lp
			}
			rv, rok, rp := CastToBool(e.Right)
			if rok && !rv {
				return rv, rok, rp || lp
			}
		case token.Plus:
			if strLit, ok := e.Left.Expr.(*ast.StringLiteral); ok && strLit.Value != "" {
				return true, true, false
			}
			if strLit, ok := e.Right.Expr.(*ast.StringLiteral); ok && strLit.Value != "" {
				return true, true, false
			}
		}
	case *ast.FunctionLiteral, *ast.ClassLiteral, *ast.NewExpression, *ast.ArrayLiteral, *ast.ObjectLiteral:
		value = true
		ok = true
	case *ast.NumberLiteral:
		if e.Value == 0.0 || math.IsNaN(e.Value) {
			return false, true, true
		}
		return true, true, true
	case *ast.BooleanLiteral:
		return e.Value, true, true
	case *ast.StringLiteral:
		return e.Value != "", true, true
	case *ast.NullLiteral:
		return false, true, true
	case *ast.RegExpLiteral:
		return true, true, true
	}

	if MayHaveSideEffects(expr) {
		return value, ok, false
	} else {
		return value, ok, true
	}
}

// AsPureNumber gets the number value if it does not have any side effects.
func AsPureNumber(expr *ast.Expression) (value float64, ok bool) {
	if v, ok, pure := CastToNumber(expr); pure {
		return v, ok
	}
	return 0.0, false
}

// CastToNumber emulates the Number() JavaScript cast function.
func CastToNumber(expr *ast.Expression) (value float64, ok bool, pure bool) {
	switch e := expr.Expr.(type) {
	case *ast.BooleanLiteral:
		if e.Value {
			return 1.0, true, true
		}
		return 0.0, true, true
	case *ast.NumberLiteral:
		return e.Value, true, true
	case *ast.StringLiteral:
		n, ok := numFromStr(e.Value)
		return n, ok, true
	case *ast.NullLiteral:
		return 0.0, false, true
	case *ast.ArrayLiteral:
		s, ok := AsPureString(expr)
		if !ok {
			return 0.0, false, false
		}
		n, ok := numFromStr(s)
		return n, ok, true
	case *ast.Identifier:
		if e.Name == "undefined" || e.Name == "NaN" && e.ScopeContext == resolver.UnresolvedMark {
			return math.NaN(), true, true
		}
		if e.Name == "Infinity" && e.ScopeContext == resolver.UnresolvedMark {
			return math.Inf(1), true, true
		}
		return 0.0, false, true
	case *ast.UnaryExpression:
		switch e.Operator {
		case token.Minus:
			n, ok, pure := CastToNumber(e.Operand)
			if ok && pure {
				return -n, true, true
			}
			return 0.0, false, false
		case token.Not:
			b, ok, pure := CastToBool(e.Operand)
			if ok && pure {
				if b {
					return 0.0, true, true
				}
				return 1.0, true, true
			}
			return 0.0, false, false
		case token.Void:
			if MayHaveSideEffects(e.Operand) {
				return math.NaN(), true, false
			} else {
				return math.NaN(), true, true
			}
		}
	case *ast.TemplateLiteral:
		s, ok := AsPureString(expr)
		if ok {
			n, ok := numFromStr(s)
			return n, ok, true
		}
	case *ast.SequenceExpression:
		if len(e.Sequence) != 0 {
			v, ok, _ := CastToNumber(&e.Sequence[len(e.Sequence)-1])
			return v, ok, false
		}
	}
	return 0.0, false, false
}

// AsPureString gets the string value if it does not have any side effects.
func AsPureString(expr *ast.Expression) (value string, ok bool) {
	objectToStr := func(name string) string {
		return fmt.Sprintf("[object %s]", name)
	}
	funcToStr := func(name string) string {
		return fmt.Sprintf("function %s() { [native code] }", name)
	}

	switch e := expr.Expr.(type) {
	case *ast.StringLiteral:
		return e.Value, true
	case *ast.NumberLiteral:
		if e.Value == 0.0 {
			return "0", true
		}
		return fmt.Sprint(e.Value), true
	case *ast.BooleanLiteral:
		return fmt.Sprint(e.Value), true
	case *ast.NullLiteral:
		return "null", true
	case *ast.TemplateLiteral:
		// TODO:
		// Only convert a template literal if all its expressions can be
		// converted.
	case *ast.Identifier:
		switch e.Name {
		case "undefined", "Infinity", "NaN":
			return e.Name, true
		case "Math", "JSON":
			return objectToStr(e.Name), true
		case "Date":
			return funcToStr(e.Name), true
		}
	case *ast.UnaryExpression:
		switch e.Operator {
		case token.Void:
			return "undefined", true
		case token.Not:
			if b, ok := AsPureBool(e.Operand); ok {
				return fmt.Sprint(!b), true
			}
		}
	case *ast.ArrayLiteral:
		var sb strings.Builder
		// null, undefined is "" in array literal.
		for idx, elem := range e.Value {
			if idx > 0 {
				sb.WriteString(",")
			}
			switch elem := elem.Expr.(type) {
			case *ast.NullLiteral:
				sb.WriteString("")
			case *ast.UnaryExpression:
				if elem.Operator == token.Void {
					if MayHaveSideEffects(elem.Operand) {
						return "", false
					}
					sb.WriteString("")
				}
			case *ast.Identifier:
				if elem.Name == "undefined" {
					sb.WriteString("")
				}
			}
			if s, ok := AsPureString(&elem); ok {
				sb.WriteString(s)
			} else {
				return "", false
			}
		}
		return sb.String(), true
	case *ast.MemberExpression:
		strLit, ok := e.Property.Expr.(*ast.StringLiteral)
		if !ok {
			return "", false
		}
		// Convert some built-in funcs to string.
		switch obj := e.Object.Expr.(type) {
		case *ast.Identifier:
			switch obj.Name {
			case "Math":
				if slices.Contains([]string{"abs", "acos", "acosh", "asin", "asinh", "atan", "atan2", "atanh", "cbrt", "ceil", "clz32", "cos", "cosh", "exp", "expm1", "floor", "fround", "hypot", "imul", "log", "log10", "log1p", "log2", "max", "min", "pow", "random", "round", "sign", "sin", "sinh", "sqrt", "tan", "tanh", "trunc"}, strLit.Value) {
					return funcToStr(strLit.Value), true
				}
			case "JSON":
				if slices.Contains([]string{"parse", "stringify"}, strLit.Value) {
					return funcToStr(strLit.Value), true
				}
			case "Date":
				if slices.Contains([]string{"now", "parse", "UTC"}, strLit.Value) {
					return funcToStr(strLit.Value), true
				}
			}
		case *ast.StringLiteral:
			if slices.Contains([]string{"anchor", "at", "big", "blink", "bold", "charAt", "charCodeAt", "codePointAt", "concat", "endsWith", "fixed", "fontcolor", "fontsize", "includes", "indexOf", "isWellFormed", "italics", "lastIndexOf", "link", "localeCompare", "match", "matchAll", "normalize", "padEnd", "padStart", "repeat", "replace", "replaceAll", "search", "slice", "small", "split", "startsWith", "strike", "sub", "substr", "substring", "sup", "toLocaleLowerCase", "toLocaleUpperCase", "toLowerCase", "toString", "toUpperCase", "toWellFormed", "trim", "trimEnd", "trimStart", "valueOf"}, strLit.Value) {
				return funcToStr(strLit.Value), true
			}
		case *ast.NumberLiteral:
			if slices.Contains([]string{"toExponential", "toFixed", "toLocaleString", "toPrecision", "toString", "valueOf"}, strLit.Value) {
				return funcToStr(strLit.Value), true
			}
		case *ast.BooleanLiteral:
			if slices.Contains([]string{"toString", "valueOf"}, strLit.Value) {
				return funcToStr(strLit.Value), true
			}
		case *ast.ArrayLiteral:
			if slices.Contains([]string{"at", "concat", "copyWithin", "entries", "every", "fill", "filter", "find", "findIndex", "findLast", "findLastIndex", "flat", "flatMap", "forEach", "includes", "indexOf", "join", "keys", "lastIndexOf", "map", "pop", "push", "reduce", "reduceRight", "reverse", "shift", "slice", "some", "sort", "splice", "toLocaleString", "toReversed", "toSorted", "toSpliced", "toString", "unshift", "values", "with"}, strLit.Value) {
				return funcToStr(strLit.Value), true
			}
		case *ast.ObjectLiteral:
			if slices.Contains([]string{"hasOwnProperty", "isPrototypeOf", "propertyIsEnumerable", "toLocaleString", "toString", "valueOf"}, strLit.Value) {
				return funcToStr(strLit.Value), true
			}
		}
	}
	return "", false
}

// GetType returns the type of the expression.
func GetType(expr *ast.Expression) (typ Type, ok bool) {
	switch e := expr.Expr.(type) {
	case *ast.AssignExpression:
		switch e.Operator {
		case token.Assign:
			return GetType(e.Right)
		case token.AddAssign:
			rt, rok := GetType(e.Right)
			if rok && rt == StringType {
				return StringType, true
			}
		case token.AndAssign, token.ExclusiveOrAssign, token.OrAssign,
			token.ShiftLeftAssign, token.ShiftRightAssign, token.UnsignedShiftRightAssign,
			token.SubtractAssign, token.MultiplyAssign, token.ExponentAssign, token.QuotientAssign, token.RemainderAssign:
			return NumberType, true
		}
	case *ast.MemberExpression:
		if strLit, ok := e.Property.Expr.(*ast.StringLiteral); ok {
			if strLit.Value == "length" {
				switch obj := e.Object.Expr.(type) {
				case *ast.ArrayLiteral, *ast.StringLiteral:
					return NumberType, true
				case *ast.Identifier:
					if obj.Name == "arguments" {
						return NumberType, true
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
			lt, lok := GetType(e.Left)
			rt, rok := GetType(e.Right)
			if lok && rok && lt == rt {
				return lt, true
			}
		case token.Plus:
			rt, rok := GetType(e.Right)
			if rok && rt == StringType {
				return StringType, true
			}
			lt, lok := GetType(e.Left)
			if lok && lt == StringType {
				return StringType, true
			}
			// There are some pretty weird cases for object types:
			//   {} + [] === "0"
			//   [] + {} ==== "[object Object]"
			if rok && rt == ObjectType {
				return UndefinedType, false
			}
			if lok && lt == ObjectType {
				return UndefinedType, false
			}
			if rok && lok && !mayBeStr(lt) && !mayBeStr(rt) {
				return NumberType, true
			}
		case token.Or, token.ExclusiveOr, token.And, token.ShiftLeft, token.ShiftRight, token.UnsignedShiftRight,
			token.Minus, token.Multiply, token.Remainder, token.Slash, token.Exponent:
			return NumberType, true
		case token.Equal, token.NotEqual, token.StrictEqual, token.StrictNotEqual, token.Less, token.LessOrEqual,
			token.Greater, token.GreaterOrEqual, token.In, token.InstanceOf:
			return BoolType, true
		}
	case *ast.ConditionalExpression:
		ct, cok := GetType(e.Consequent)
		at, aok := GetType(e.Alternate)
		if cok && aok && ct == at {
			return ct, true
		}
	case *ast.NumberLiteral:
		return NumberType, true
	case *ast.UnaryExpression:
		switch e.Operator {
		case token.Minus, token.Plus, token.BitwiseNot:
			return NumberType, true
		case token.Not, token.Delete:
			return BoolType, true
		case token.Typeof:
			return StringType, true
		case token.Void:
			return UndefinedType, true
		}
	case *ast.UpdateExpression:
		switch e.Operator {
		case token.Increment, token.Decrement:
			return NumberType, true
		}
	case *ast.BooleanLiteral:
		return BoolType, true
	case *ast.StringLiteral, *ast.TemplateLiteral:
		return StringType, true
	case *ast.NullLiteral:
		return NullType, true
	case *ast.FunctionLiteral, *ast.NewExpression, *ast.ArrayLiteral, *ast.ObjectLiteral, *ast.RegExpLiteral:
		return ObjectType, true
	}
	return UndefinedType, false
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
		if strLit, ok := e.Property.Expr.(*ast.StringLiteral); ok {
			if slices.Contains([]string{"charAt", "charCodeAt", "concat", "endsWith",
				"includes", "indexOf", "lastIndexOf", "localeCompare", "slice", "split",
				"startsWith", "substr", "substring", "toLocaleLowerCase", "toLocaleUpperCase",
				"toLowerCase", "toString", "toUpperCase", "trim", "trimEnd", "trimStart"}, strLit.Value) {
				return true
			}
		}
	case *ast.FunctionLiteral:
		if !slices.ContainsFunc(e.ParameterList.List, func(decl ast.VariableDeclarator) bool {
			_, ok := decl.Initializer.Expr.(*ast.Identifier)
			return !ok
		}) && len(e.Body.List) == 0 {
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

			if _, ok := e.Property.Expr.(*ast.StringLiteral); ok {
				return false
			}
			return MayHaveSideEffects(e.Property)
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
	}
	return true
}
