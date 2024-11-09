package simplifier

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/resolver"
	"github.com/t14raptor/go-fast/token"
)

func isNaN(expr *ast.Expression) bool {
	return isGlobalRefTo(expr, "NaN")
}

func isGlobalRefTo(expr *ast.Expression, id string) bool {
	if ident, ok := expr.Expr.(*ast.Identifier); ok {
		return ident.Name == id && ident.ScopeContext == resolver.UnresolvedMark
	}
	return false
}

func and(av, bv bool, aok, bok bool) (value bool, ok bool) {
	if aok && av {
		return bv, bok
	} else if aok && !av {
		return false, true
	} else if bok && !bv {
		return false, true
	}
	return false, false
}

func or(av, bv bool, aok, bok bool) (value bool, ok bool) {
	if aok && av {
		return true, true
	} else if aok && !av {
		return bv, bok
	} else if bok && bv {
		return true, true
	}
	return false, false
}

func asPureBool(expr *ast.Expression) (value bool, ok bool) {
	if v, ok, pure := castToBool(expr); pure {
		return v, ok
	}
	return false, false
}

func castToBool(expr *ast.Expression) (value bool, ok bool, pure bool) {
	if isGlobalRefTo(expr, "undefined") {
		return false, true, true
	}
	if isNaN(expr) {
		return false, true, true
	}

	switch e := expr.Expr.(type) {
	case *ast.AssignExpression:
		if e.Operator == token.Assign {
			v, ok, _ := castToBool(e.Right)
			return v, ok, false
		}
	case *ast.UnaryExpression:
		switch e.Operator {
		case token.Minus:
			n, ok := asPureNumber(e.Operand)
			if ok {
				value = !(math.IsNaN(n) || n == 0)
			} else {
				return false, false, false
			}
		case token.Not:
			b, ok, pure := castToBool(e.Operand)
			return !b, ok, pure
		case token.Void:
			value = false
			ok = true
		}
	case *ast.SequenceExpression:
		if len(e.Sequence) != 0 {
			value, ok, _ = castToBool(&e.Sequence[len(e.Sequence)-1])
		}
	case *ast.BinaryExpression:
		switch e.Operator {
		case token.Minus:
			nl, okl, pl := castToNumber(e.Left)
			nr, okr, pr := castToNumber(e.Right)

			return nl != nr, okl && okr, pl && pr
		case token.Slash:
			nl, okl := asPureNumber(e.Left)
			nr, okr := asPureNumber(e.Right)
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
			lt, lok := getType(e.Left)
			rt, rok := getType(e.Right)
			if lok && rok && lt != BooleanType && rt != BooleanType {
				return false, false, false
			}

			// TODO: Ignore purity if value cannot be reached.
			lv, lok, lp := castToBool(e.Left)
			rv, rok, rp := castToBool(e.Right)

			if e.Operator == token.And {
				value, ok = and(lv, rv, lok, rok)
			} else {
				value, ok = or(lv, rv, lok, rok)
			}
			if lp && rp {
				return value, ok, true
			}
		case token.LogicalOr:
			lv, lok, lp := castToBool(e.Left)
			if lok && lv {
				return lv, lok, lp
			}
			rv, rok, rp := castToBool(e.Right)
			if rok && rv {
				return rv, rok, rp
			}
		case token.LogicalAnd:
			lv, lok, lp := castToBool(e.Left)
			if lok && !lv {
				return lv, lok, lp
			}
			rv, rok, rp := castToBool(e.Right)
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

	if mayHaveSideEffects(expr) {
		return value, ok, false
	} else {
		return value, ok, true
	}
}

func asPureNumber(expr *ast.Expression) (value float64, ok bool) {
	if v, ok, pure := castToNumber(expr); pure {
		return v, ok
	}
	return 0.0, false
}

func castToNumber(expr *ast.Expression) (value float64, ok bool, pure bool) {
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
		s, ok := asPureString(expr)
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
			n, ok, pure := castToNumber(e.Operand)
			if ok && pure {
				return -n, true, true
			}
			return 0.0, false, false
		case token.Not:
			b, ok, pure := castToBool(e.Operand)
			if ok && pure {
				if b {
					return 0.0, true, true
				}
				return 1.0, true, true
			}
			return 0.0, false, false
		case token.Void:
			if mayHaveSideEffects(e.Operand) {
				return math.NaN(), true, false
			} else {
				return math.NaN(), true, true
			}
		}
	case *ast.TemplateLiteral:
		s, ok := asPureString(expr)
		if ok {
			n, ok := numFromStr(s)
			return n, ok, true
		}
	case *ast.SequenceExpression:
		if len(e.Sequence) != 0 {
			v, ok, _ := castToNumber(&e.Sequence[len(e.Sequence)-1])
			return v, ok, false
		}
	}
	return 0.0, false, false
}

func asPureString(expr *ast.Expression) (value string, ok bool) {
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
		if slices.Contains([]string{"undefined", "Infinity", "NaN"}, e.Name) {
			return e.Name, true
		}
	case *ast.UnaryExpression:
		switch e.Operator {
		case token.Void:
			return "undefined", true
		case token.Not:
			if b, ok := asPureBool(e.Operand); ok {
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
					if mayHaveSideEffects(elem.Operand) {
						return "", false
					}
					sb.WriteString("")
				}
			case *ast.Identifier:
				if elem.Name == "undefined" {
					sb.WriteString("")
				}
			}
			if s, ok := asPureString(&elem); ok {
				sb.WriteString(s)
			} else {
				return "", false
			}
		}
		return sb.String(), true
	}
	return "", false
}

type Type int

const (
	UndefinedType Type = iota
	NullType
	BooleanType
	StringType
	SymbolType
	NumberType
	ObjectType
)

func (t Type) CastToNumberOnAdd() bool {
	switch t {
	case BooleanType, NullType, NumberType, UndefinedType:
		return true
	default:
		return false
	}
}

func getType(expr *ast.Expression) (typ Type, ok bool) {
	switch e := expr.Expr.(type) {
	case *ast.AssignExpression:
		switch e.Operator {
		case token.Assign:
			return getType(e.Right)
		case token.AddAssign:
			rt, rok := getType(e.Right)
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
			return getType(&e.Sequence[len(e.Sequence)-1])
		}
	case *ast.BinaryExpression:
		switch e.Operator {
		case token.LogicalAnd, token.LogicalOr:
			lt, lok := getType(e.Left)
			rt, rok := getType(e.Right)
			if lok && rok && lt == rt {
				return lt, true
			}
		case token.Plus:
			rt, rok := getType(e.Right)
			if rok && rt == StringType {
				return StringType, true
			}
			lt, lok := getType(e.Left)
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
			return BooleanType, true
		}
	case *ast.ConditionalExpression:
		ct, cok := getType(e.Consequent)
		at, aok := getType(e.Alternate)
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
			return BooleanType, true
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
		return BooleanType, true
	case *ast.StringLiteral, *ast.TemplateLiteral:
		return StringType, true
	case *ast.NullLiteral:
		return NullType, true
	case *ast.FunctionLiteral, *ast.NewExpression, *ast.ArrayLiteral, *ast.ObjectLiteral, *ast.RegExpLiteral:
		return ObjectType, true
	}
	return UndefinedType, false
}

func isPureCallee(expr *ast.Expression) bool {
	if isGlobalRefTo(expr, "Date") {
		return true
	}
	switch e := expr.Expr.(type) {
	case *ast.MemberExpression:
		if isGlobalRefTo(e.Object, "Math") {
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

func mayHaveSideEffects(expr *ast.Expression) bool {
	if isPureCallee(expr) {
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
			if mayHaveSideEffects(&elem) {
				return true
			}
		}
		return false
	case *ast.UnaryExpression:
		if e.Operator == token.Delete {
			return true
		}
		return mayHaveSideEffects(e.Operand)
	case *ast.BinaryExpression:
		return mayHaveSideEffects(e.Left) || mayHaveSideEffects(e.Right)
	case *ast.MemberExpression:
		switch e.Object.Expr.(type) {
		case *ast.ObjectLiteral, *ast.FunctionLiteral, *ast.ArrowFunctionLiteral, *ast.ClassLiteral:
			if mayHaveSideEffects(e.Object) {
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
			return mayHaveSideEffects(e.Property)
		}

	case *ast.TemplateLiteral:
	case *ast.MetaProperty:
	case *ast.AwaitExpression, *ast.YieldExpression, *ast.SuperExpression, *ast.UpdateExpression, *ast.AssignExpression:

	case *ast.NewExpression:

	case *ast.OptionalChain:
		switch base := e.Base.Expr.(type) {
		case *ast.MemberExpression:
		case *ast.CallExpression:
			if isPureCallee(base.Callee) {
				for _, arg := range base.ArgumentList {
					if mayHaveSideEffects(&arg) {
						return true
					}
				}
				return false
			}
		}
	case *ast.CallExpression:
		if isPureCallee(e.Callee) {
			for _, arg := range e.ArgumentList {
				if mayHaveSideEffects(&arg) {
					return true
				}
			}
			return false
		}
	case *ast.SequenceExpression:
		for _, expr := range e.Sequence {
			if mayHaveSideEffects(&expr) {
				return true
			}
		}
		return false
	case *ast.ConditionalExpression:
		return mayHaveSideEffects(e.Test) || mayHaveSideEffects(e.Consequent) || mayHaveSideEffects(e.Alternate)
	case *ast.ObjectLiteral:
		for _, prop := range e.Value {
			switch p := prop.Prop.(type) {
			case *ast.SpreadElement:
				return true
			case *ast.PropertyShort:
			case *ast.PropertyKeyed:
				if p.Computed && mayHaveSideEffects(p.Key) {
					return true
				}
				if mayHaveSideEffects(p.Value) {
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

func mayHaveSideEffectsStmt(stmt ast.Statement) bool {
	switch s := stmt.Stmt.(type) {
	case *ast.BlockStatement:
		for _, stmt := range s.List {
			if mayHaveSideEffectsStmt(stmt) {
				return true
			}
		}
		return false
	case *ast.EmptyStatement:
		return false
	case *ast.LabelledStatement:
		return mayHaveSideEffectsStmt(*s.Statement)
	case *ast.IfStatement:
		if mayHaveSideEffects(s.Test) || mayHaveSideEffectsStmt(*s.Consequent) {
			return true
		}
		if s.Alternate != nil && mayHaveSideEffectsStmt(*s.Alternate) {
			return true
		}
		return false
	case *ast.SwitchStatement:
		if mayHaveSideEffects(s.Discriminant) {
			return true
		}
		for _, c := range s.Body {
			if mayHaveSideEffects(c.Test) {
				return true
			}
			for _, stmt := range c.Consequent {
				if mayHaveSideEffectsStmt(stmt) {
					return true
				}
			}
		}
		return false
	case *ast.TryStatement:
		for _, stmt := range s.Body.List {
			if mayHaveSideEffectsStmt(stmt) {
				return true
			}
		}
		if s.Catch != nil {
			for _, stmt := range s.Catch.Body.List {
				if mayHaveSideEffectsStmt(stmt) {
					return true
				}
			}
		}
		if s.Finally != nil {
			for _, stmt := range s.Finally.List {
				if mayHaveSideEffectsStmt(stmt) {
					return true
				}
			}
		}
		return false
	case *ast.ClassDeclaration:
		return classHasSideEffect(s.Class)
	case *ast.FunctionDeclaration:
		// TODO: Check in_strict mode like swc
	case *ast.VariableDeclaration:
		return s.Token == token.Var
	case *ast.ExpressionStatement:
		return mayHaveSideEffects(s.Expression)
	}
	return true
}

func classHasSideEffect(class *ast.ClassLiteral) bool {
	if class.SuperClass != nil {
		if mayHaveSideEffects(class.SuperClass) {
			return true
		}
	}
	for _, elem := range class.Body {
		switch elem := elem.Element.(type) {
		case *ast.MethodDefinition:
			if elem.Computed && mayHaveSideEffects(elem.Key) {
				return true
			}
		case *ast.FieldDefinition:
			if elem.Computed && mayHaveSideEffects(elem.Key) {
				return true
			}
			if elem.Initializer != nil && mayHaveSideEffects(elem.Initializer) {
				return true
			}
		case *ast.ClassStaticBlock:
			for _, stmt := range elem.Block.List {
				if mayHaveSideEffectsStmt(stmt) {
					return true
				}
			}
		}
	}
	return false
}

func mayBeStr(ty Type) bool {
	switch ty {
	case BooleanType, NullType, NumberType, UndefinedType:
		return false
	case ObjectType, StringType:
		return true
	case SymbolType:
		return true
	}
	return true
}

func numFromStr(str string) (value float64, ok bool) {
	if strings.ContainsRune(str, '\u000B') {
		return 0, false
	}
	s := strings.TrimSpace(str)
	if s == "" {
		return 0, true
	}
	if len(s) >= 2 && s[0] == '0' {
		switch {
		case s[1] == 'x' || s[1] == 'X':
			if n, err := strconv.ParseInt(s[2:], 16, 64); err == nil {
				return float64(n), true
			} else {
				return math.NaN(), true
			}
		case s[1] == 'o' || s[1] == 'O':
			if n, err := strconv.ParseInt(s[2:], 8, 64); err == nil {
				return float64(n), true
			} else {
				return math.NaN(), true
			}
		case s[1] == 'b' || s[1] == 'B':
			if n, err := strconv.ParseInt(s[2:], 2, 64); err == nil {
				return float64(n), true
			} else {
				return math.NaN(), true
			}
		}
	}
	if (strings.HasPrefix(s, "-") || strings.HasPrefix(s, "+")) &&
		(strings.HasPrefix(s[1:], "0x") || strings.HasPrefix(s[1:], "0X")) {
		return 0, false
	}
	// Firefox and IE treat the "Infinity" differently. Firefox is case
	// insensitive, but IE treats "infinity" as NaN.  So leave it alone.
	switch s {
	case "infinity", "+infinity", "-infinity":
		return 0, false
	}
	if n, err := strconv.ParseFloat(s, 64); err == nil {
		return n, true
	}
	return math.NaN(), true
}

func isLiteral(n ast.VisitableNode) bool {
	_, isLit := calcLiteralCost(n, true)
	return isLit
}

func calcLiteralCost(n ast.VisitableNode, allowNonJsonValue bool) (cost int, isLit bool) {
	visitor := &LiteralVisitor{isLit: true, allowNonJsonValue: allowNonJsonValue}
	visitor.V = visitor
	n.VisitWith(visitor)
	return visitor.cost, visitor.isLit
}

type LiteralVisitor struct {
	ast.NoopVisitor
	isLit             bool
	cost              int
	allowNonJsonValue bool
}

func (v *LiteralVisitor) VisitArrayLiteral(n *ast.ArrayLiteral) {
	if !v.isLit {
		return
	}
	v.cost += 2 + len(n.Value)
	n.VisitChildrenWith(v)
	for _, elem := range n.Value {
		if !v.allowNonJsonValue && elem.Expr == nil {
			v.isLit = false
		}
	}
}
func (v *LiteralVisitor) VisitArrowFunctionLiteral(n *ast.ArrowFunctionLiteral)   { v.isLit = false }
func (v *LiteralVisitor) VisitAssignExpression(n *ast.AssignExpression)           { v.isLit = false }
func (v *LiteralVisitor) VisitAwaitExpression(n *ast.AwaitExpression)             { v.isLit = false }
func (v *LiteralVisitor) VisitBinaryExpression(n *ast.BinaryExpression)           { v.isLit = false }
func (v *LiteralVisitor) VisitCallExpression(n *ast.CallExpression)               { v.isLit = false }
func (v *LiteralVisitor) VisitClassLiteral(n *ast.ClassLiteral)                   { v.isLit = false }
func (v *LiteralVisitor) VisitConditionalExpression(n *ast.ConditionalExpression) { v.isLit = false }
func (v *LiteralVisitor) VisitExpression(n *ast.Expression) {
	if !v.isLit {
		return
	}
	switch e := n.Expr.(type) {
	case *ast.Identifier, *ast.RegExpLiteral:
		v.isLit = false
	case *ast.TemplateLiteral:
		if e.Expressions != nil {
			v.isLit = false
		}
	default:
		n.VisitChildrenWith(v)
	}
}
func (v *LiteralVisitor) VisitFunctionLiteral(n *ast.FunctionLiteral)     { v.isLit = false }
func (v *LiteralVisitor) VisitInvalidExpression(n *ast.InvalidExpression) { v.isLit = false }
func (v *LiteralVisitor) VisitMemberExpression(n *ast.MemberExpression)   { v.isLit = false }
func (v *LiteralVisitor) VisitMetaProperty(n *ast.MetaProperty)           { v.isLit = false }
func (v *LiteralVisitor) VisitNewExpression(n *ast.NewExpression)         { v.isLit = false }
func (v *LiteralVisitor) VisitNumberLiteral(n *ast.NumberLiteral) {
	if !v.allowNonJsonValue && (math.IsNaN(n.Value) || math.IsInf(n.Value, 0)) {
		v.isLit = false
	}
}
func (v *LiteralVisitor) VisitOptionalChain(n *ast.OptionalChain)         { v.isLit = false }
func (v *LiteralVisitor) VisitPrivateIdentifier(n *ast.PrivateIdentifier) { v.isLit = false }
func (v *LiteralVisitor) VisitProperty(n *ast.Property) {
	if !v.isLit {
		return
	}
	n.VisitChildrenWith(v)
	switch p := n.Prop.(type) {
	case *ast.PropertyKeyed:
		switch p := p.Key.Expr.(type) {
		case *ast.StringLiteral:
			v.cost += 2 + len(p.Value)
		case *ast.Identifier:
			v.cost += 2 + len(p.Name)
		case *ast.NumberLiteral:
			v.cost += 2 + len(strconv.FormatFloat(p.Value, 'f', -1, 64))
		}
	}
	switch n.Prop.(type) {
	case *ast.PropertyKeyed:
		v.cost++
	default:
		v.isLit = false
	}
}
func (v *LiteralVisitor) VisitSequenceExpression(n *ast.SequenceExpression) { v.isLit = false }
func (v *LiteralVisitor) VisitSpreadElement(n *ast.SpreadElement)           { v.isLit = false }
func (v *LiteralVisitor) VisitThisExpression(n *ast.ThisExpression)         { v.isLit = false }
func (v *LiteralVisitor) VisitUnaryExpression(n *ast.UnaryExpression)       { v.isLit = false }
func (v *LiteralVisitor) VisitUpdateExpression(n *ast.UpdateExpression)     { v.isLit = false }
func (v *LiteralVisitor) VisitYieldExpression(n *ast.YieldExpression)       { v.isLit = false }

func preserveEffects(val ast.Expression, exprs []ast.Expression) ast.Expression {
	var exprs2 []ast.Expression
	for _, expr := range exprs {
		extractSideEffectsTo(&exprs2, &expr)
	}
	if len(exprs2) == 0 {
		return val
	} else {
		exprs2 = append(exprs2, val)
		return ast.Expression{Expr: &ast.SequenceExpression{Sequence: exprs2}}
	}
}

func extractSideEffectsTo(to *[]ast.Expression, expr *ast.Expression) {
	switch e := expr.Expr.(type) {
	case *ast.StringLiteral, *ast.BooleanLiteral, *ast.NullLiteral, *ast.NumberLiteral, *ast.RegExpLiteral,
		*ast.ThisExpression, *ast.FunctionLiteral, *ast.ArrowFunctionLiteral, *ast.PrivateIdentifier:
	case *ast.Identifier:
		if mayHaveSideEffects(expr) {
			*to = append(*to, *expr)
		}
	// In most case, we can do nothing for this.
	case *ast.UpdateExpression, *ast.AssignExpression, *ast.YieldExpression, *ast.AwaitExpression:
		*to = append(*to, *expr)
	// TODO
	case *ast.MetaProperty:
		*to = append(*to, *expr)
	case *ast.CallExpression:
		*to = append(*to, *expr)
	case *ast.NewExpression:
		// Known constructors
		if ident, ok := e.Callee.Expr.(*ast.Identifier); ok && ident.Name == "Date" && len(e.ArgumentList) == 0 {
			return
		}
		*to = append(*to, *expr)
	case *ast.MemberExpression, *ast.SuperExpression:
		*to = append(*to, *expr)
	// We are at here because we could not determine value of test.
	//TODO: Drop values if it does not have side effects.
	case *ast.ConditionalExpression:
		*to = append(*to, *expr)
	case *ast.UnaryExpression:
		if e.Operator == token.Typeof {
			if _, ok := e.Operand.Expr.(*ast.Identifier); ok {
				return
			}
		}
		extractSideEffectsTo(to, e.Operand)
	case *ast.BinaryExpression:
		if e.Operator.MayShortCircuit() {
			*to = append(*to, *expr)
		} else {
			extractSideEffectsTo(to, e.Left)
			extractSideEffectsTo(to, e.Right)
		}
	case *ast.SequenceExpression:
		for _, expr := range e.Sequence {
			extractSideEffectsTo(to, &expr)
		}
	case *ast.ObjectLiteral:
		var hasSpread bool
		e.Value = slices.DeleteFunc(e.Value, func(prop ast.Property) bool {
			switch p := prop.Prop.(type) {
			case *ast.PropertyShort:
				return false
			case *ast.PropertyKeyed:
				if p.Computed && mayHaveSideEffects(p.Key) {
					return true
				}
				return mayHaveSideEffects(p.Value)
			case *ast.SpreadElement:
				hasSpread = true
				return true
			}
			return false
		})
		if hasSpread {
			*to = append(*to, *expr)
		} else {
			for _, prop := range e.Value {
				switch p := prop.Prop.(type) {
				case *ast.PropertyShort:
				case *ast.PropertyKeyed:
					if p.Computed {
						extractSideEffectsTo(to, p.Key)
					}
					extractSideEffectsTo(to, p.Value)
				}
			}
		}
	case *ast.ArrayLiteral:
		for _, elem := range e.Value {
			extractSideEffectsTo(to, &elem)
		}
	case *ast.TemplateLiteral:
		for _, elem := range e.Expressions {
			extractSideEffectsTo(to, &elem)
		}
	case *ast.ClassLiteral:
		panic("add_effects for class expression")
	case *ast.OptionalChain:
		*to = append(*to, *expr)
	}
}

func propNameEq(p *ast.Expression, key string) bool {
	switch p := p.Expr.(type) {
	case *ast.Identifier:
		return p.Name == key
	case *ast.StringLiteral:
		return p.Value == key
	case *ast.NumberLiteral:
		return strconv.FormatFloat(p.Value, 'f', -1, 64) == key
	}
	return false
}
