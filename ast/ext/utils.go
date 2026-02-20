package ext

import (
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

// classHasSideEffect returns true if the class may have side effects.
func classHasSideEffect(class *ast.ClassLiteral) bool {
	if class.SuperClass != nil {
		if MayHaveSideEffects(class.SuperClass) {
			return true
		}
	}
	for _, elem := range class.Body {
		switch elem.Kind() {
		case ast.ClassElemMethod:
			method := elem.MustMethod()
			if method.Computed && MayHaveSideEffects(method.Key) {
				return true
			}
		case ast.ClassElemField:
			field := elem.MustField()
			if field.Computed && MayHaveSideEffects(field.Key) {
				return true
			}
			if field.Initializer != nil && MayHaveSideEffects(field.Initializer) {
				return true
			}
		case ast.ClassElemStaticBlock:
			sb := elem.MustStaticBlock()
			for _, stmt := range sb.Block.List {
				if MayHaveSideEffectsStmt(stmt) {
					return true
				}
			}
		}
	}
	return false
}

// mayBeStr returns if the node is possibly a string.
func mayBeStr(ty Type) bool {
	switch ty.(type) {
	case BoolType, NullType, NumberType, UndefinedType:
		return false
	case ObjectType, StringType:
		return true
	case SymbolType:
		return true
	}
	return true
}

// numFromStr converts a string to a number.
func numFromStr(str string) Value[float64] {
	if strings.ContainsRune(str, '\u000B') {
		return Value[float64]{unknown: true}
	}
	s := strings.TrimSpace(str)
	if s == "" {
		return Value[float64]{}
	}
	if len(s) >= 2 && s[0] == '0' {
		switch {
		case s[1] == 'x' || s[1] == 'X':
			if n, err := strconv.ParseInt(s[2:], 16, 64); err == nil {
				return Value[float64]{val: float64(n)}
			} else {
				return Value[float64]{val: math.NaN()}
			}
		case s[1] == 'o' || s[1] == 'O':
			if n, err := strconv.ParseInt(s[2:], 8, 64); err == nil {
				return Value[float64]{val: float64(n)}
			} else {
				return Value[float64]{val: math.NaN()}
			}
		case s[1] == 'b' || s[1] == 'B':
			if n, err := strconv.ParseInt(s[2:], 2, 64); err == nil {
				return Value[float64]{val: float64(n)}
			} else {
				return Value[float64]{val: math.NaN()}
			}
		}
	}
	if (strings.HasPrefix(s, "-") || strings.HasPrefix(s, "+")) &&
		(strings.HasPrefix(s[1:], "0x") || strings.HasPrefix(s[1:], "0X")) {
		return Value[float64]{unknown: true}
	}
	// Firefox and IE treat the "Infinity" differently. Firefox is case
	// insensitive, but IE treats "infinity" as NaN.  So leave it alone.
	switch s {
	case "infinity", "+infinity", "-infinity":
		return Value[float64]{unknown: true}
	}
	if n, err := strconv.ParseFloat(s, 64); err == nil {
		return Value[float64]{val: n}
	}
	return Value[float64]{val: math.NaN()}
}

// IsLiteral returns true if the node is a literal.
func IsLiteral(n ast.VisitableNode) bool {
	_, isLit := CalcLiteralCost(n, true)
	return isLit
}

// CalcLiteralCost calculates the cost of the node if it is a literal.
func CalcLiteralCost(n ast.VisitableNode, allowNonJsonValue bool) (cost int, isLit bool) {
	visitor := &literalVisitor{isLit: true, allowNonJsonValue: allowNonJsonValue}
	visitor.V = visitor
	n.VisitWith(visitor)
	return visitor.cost, visitor.isLit
}

type literalVisitor struct {
	ast.NoopVisitor
	isLit             bool
	cost              int
	allowNonJsonValue bool
}

func (v *literalVisitor) VisitArrayLiteral(n *ast.ArrayLiteral) {
	if !v.isLit {
		return
	}
	v.cost += 2 + len(n.Value)
	n.VisitChildrenWith(v)
	for _, elem := range n.Value {
		if !v.allowNonJsonValue && elem.IsNone() {
			v.isLit = false
		}
	}
}
func (v *literalVisitor) VisitArrowFunctionLiteral(n *ast.ArrowFunctionLiteral)   { v.isLit = false }
func (v *literalVisitor) VisitAssignExpression(n *ast.AssignExpression)           { v.isLit = false }
func (v *literalVisitor) VisitAwaitExpression(n *ast.AwaitExpression)             { v.isLit = false }
func (v *literalVisitor) VisitBinaryExpression(n *ast.BinaryExpression)           { v.isLit = false }
func (v *literalVisitor) VisitCallExpression(n *ast.CallExpression)               { v.isLit = false }
func (v *literalVisitor) VisitClassLiteral(n *ast.ClassLiteral)                   { v.isLit = false }
func (v *literalVisitor) VisitConditionalExpression(n *ast.ConditionalExpression) { v.isLit = false }
func (v *literalVisitor) VisitExpression(n *ast.Expression) {
	if !v.isLit {
		return
	}
	switch n.Kind() {
	case ast.ExprIdent, ast.ExprRegExpLit:
		v.isLit = false
	case ast.ExprTmplLit:
		if n.MustTmplLit().Expressions != nil {
			v.isLit = false
		}
	default:
		n.VisitChildrenWith(v)
	}
}
func (v *literalVisitor) VisitFunctionLiteral(n *ast.FunctionLiteral)     { v.isLit = false }
func (v *literalVisitor) VisitInvalidExpression(n *ast.InvalidExpression) { v.isLit = false }
func (v *literalVisitor) VisitMemberExpression(n *ast.MemberExpression)   { v.isLit = false }
func (v *literalVisitor) VisitMetaProperty(n *ast.MetaProperty)           { v.isLit = false }
func (v *literalVisitor) VisitNewExpression(n *ast.NewExpression)         { v.isLit = false }
func (v *literalVisitor) VisitNumberLiteral(n *ast.NumberLiteral) {
	if !v.allowNonJsonValue && (math.IsNaN(n.Value) || math.IsInf(n.Value, 0)) {
		v.isLit = false
	}
}
func (v *literalVisitor) VisitOptionalChain(n *ast.OptionalChain)         { v.isLit = false }
func (v *literalVisitor) VisitPrivateIdentifier(n *ast.PrivateIdentifier) { v.isLit = false }
func (v *literalVisitor) VisitProperty(n *ast.Property) {
	if !v.isLit {
		return
	}
	n.VisitChildrenWith(v)
	switch n.Kind() {
	case ast.PropKeyed:
		p := n.MustKeyed()
		switch p.Key.Kind() {
		case ast.ExprStrLit:
			v.cost += 2 + len(p.Key.MustStrLit().Value)
		case ast.ExprIdent:
			v.cost += 2 + len(p.Key.MustIdent().Name)
		case ast.ExprNumLit:
			v.cost += 2 + len(strconv.FormatFloat(p.Key.MustNumLit().Value, 'f', -1, 64))
		}
		v.cost++
	default:
		v.isLit = false
	}
}
func (v *literalVisitor) VisitSequenceExpression(n *ast.SequenceExpression) { v.isLit = false }
func (v *literalVisitor) VisitSpreadElement(n *ast.SpreadElement)           { v.isLit = false }
func (v *literalVisitor) VisitThisExpression(n *ast.ThisExpression)         { v.isLit = false }
func (v *literalVisitor) VisitUnaryExpression(n *ast.UnaryExpression) {
	if !v.isLit {
		return
	}
	switch n.Operator {
	case token.Minus, token.Plus, token.BitwiseNot, token.Not:
		v.cost++
		n.VisitChildrenWith(v)
	default:
		v.isLit = false
	}
}
func (v *literalVisitor) VisitUpdateExpression(n *ast.UpdateExpression) { v.isLit = false }
func (v *literalVisitor) VisitYieldExpression(n *ast.YieldExpression)   { v.isLit = false }

// PreserveEffects makes a new expression which evaluates val preserving side effects, if any.
func PreserveEffects(val ast.Expression, exprs []ast.Expression) ast.Expression {
	var exprs2 []ast.Expression
	for _, expr := range exprs {
		ExtractSideEffectsTo(&exprs2, &expr)
	}
	if len(exprs2) == 0 {
		return val
	} else {
		exprs2 = append(exprs2, val)
		return ast.NewSequenceExpr(&ast.SequenceExpression{Sequence: exprs2})
	}
}

// ExtractSideEffectsTo adds side effects of expr to to.
func ExtractSideEffectsTo(to *[]ast.Expression, expr *ast.Expression) {
	if expr == nil || expr.IsNone() {
		return
	}
	switch expr.Kind() {
	case ast.ExprStrLit, ast.ExprBoolLit, ast.ExprNullLit, ast.ExprNumLit, ast.ExprRegExpLit,
		ast.ExprThis, ast.ExprFuncLit, ast.ExprArrowFuncLit, ast.ExprPrivIdent:
	case ast.ExprIdent:
		if MayHaveSideEffects(expr) {
			*to = append(*to, *expr)
		}
	case ast.ExprUpdate, ast.ExprAssign, ast.ExprYield, ast.ExprAwait:
		*to = append(*to, *expr)
	case ast.ExprMetaProp:
		*to = append(*to, *expr)
	case ast.ExprCall:
		*to = append(*to, *expr)
	case ast.ExprNew:
		e := expr.MustNew()
		if id, ok := e.Callee.Ident(); ok && id.Name == "Date" && len(e.ArgumentList) == 0 {
			return
		}
		*to = append(*to, *expr)
	case ast.ExprMember, ast.ExprSuper:
		*to = append(*to, *expr)
	case ast.ExprConditional:
		*to = append(*to, *expr)
	case ast.ExprUnary:
		e := expr.MustUnary()
		if e.Operator == token.Typeof {
			if _, ok := e.Operand.Ident(); ok {
				return
			}
		}
		ExtractSideEffectsTo(to, e.Operand)
	case ast.ExprBinary:
		e := expr.MustBinary()
		if e.Operator.MayShortCircuit() {
			*to = append(*to, *expr)
		} else {
			ExtractSideEffectsTo(to, e.Left)
			ExtractSideEffectsTo(to, e.Right)
		}
	case ast.ExprSequence:
		e := expr.MustSequence()
		for _, expr := range e.Sequence {
			ExtractSideEffectsTo(to, &expr)
		}
	case ast.ExprObjLit:
		e := expr.MustObjLit()
		var hasSpread bool
		e.Value = slices.DeleteFunc(e.Value, func(prop ast.Property) bool {
			switch prop.Kind() {
			case ast.PropShort:
				return true
			case ast.PropKeyed:
				p := prop.MustKeyed()
				if p.Computed && MayHaveSideEffects(p.Key) {
					return false
				}
				return !MayHaveSideEffects(p.Value)
			case ast.PropSpread:
				hasSpread = true
				return false
			}
			return true
		})
		if hasSpread {
			*to = append(*to, *expr)
		} else {
			for _, prop := range e.Value {
				switch prop.Kind() {
				case ast.PropShort:
				case ast.PropKeyed:
					p := prop.MustKeyed()
					if p.Computed {
						ExtractSideEffectsTo(to, p.Key)
					}
					ExtractSideEffectsTo(to, p.Value)
				}
			}
		}
	case ast.ExprArrLit:
		e := expr.MustArrLit()
		for _, elem := range e.Value {
			ExtractSideEffectsTo(to, &elem)
		}
	case ast.ExprTmplLit:
		e := expr.MustTmplLit()
		for _, elem := range e.Expressions {
			ExtractSideEffectsTo(to, &elem)
		}
	case ast.ExprClassLit:
		panic("add_effects for class expression")
	case ast.ExprOptChain:
		*to = append(*to, *expr)
	}
}

// PropNameEq returns true if the property name of the expression is equal to key.
func PropNameEq(p *ast.Expression, key string) bool {
	switch p.Kind() {
	case ast.ExprIdent:
		return p.MustIdent().Name == key
	case ast.ExprStrLit:
		return p.MustStrLit().Value == key
	case ast.ExprNumLit:
		return strconv.FormatFloat(p.MustNumLit().Value, 'f', -1, 64) == key
	}
	return false
}
