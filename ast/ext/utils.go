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
		switch elem := elem.Element.(type) {
		case *ast.MethodDefinition:
			if elem.Computed && MayHaveSideEffects(elem.Key) {
				return true
			}
		case *ast.FieldDefinition:
			if elem.Computed && MayHaveSideEffects(elem.Key) {
				return true
			}
			if elem.Initializer != nil && MayHaveSideEffects(elem.Initializer) {
				return true
			}
		case *ast.ClassStaticBlock:
			for _, stmt := range elem.Block.List {
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
		if !v.allowNonJsonValue && elem.Expr == nil {
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
		return ast.Expression{Expr: &ast.SequenceExpression{Sequence: exprs2}}
	}
}

// ExtractSideEffectsTo adds side effects of expr to to.
func ExtractSideEffectsTo(to *[]ast.Expression, expr *ast.Expression) {
	switch e := expr.Expr.(type) {
	case *ast.StringLiteral, *ast.BooleanLiteral, *ast.NullLiteral, *ast.NumberLiteral, *ast.RegExpLiteral,
		*ast.ThisExpression, *ast.FunctionLiteral, *ast.ArrowFunctionLiteral, *ast.PrivateIdentifier:
	case *ast.Identifier:
		if MayHaveSideEffects(expr) {
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
		ExtractSideEffectsTo(to, e.Operand)
	case *ast.BinaryExpression:
		if e.Operator.MayShortCircuit() {
			*to = append(*to, *expr)
		} else {
			ExtractSideEffectsTo(to, e.Left)
			ExtractSideEffectsTo(to, e.Right)
		}
	case *ast.SequenceExpression:
		for _, expr := range e.Sequence {
			ExtractSideEffectsTo(to, &expr)
		}
	case *ast.ObjectLiteral:
		var hasSpread bool
		e.Value = slices.DeleteFunc(e.Value, func(prop ast.Property) bool {
			switch p := prop.Prop.(type) {
			case *ast.PropertyShort:
				return true
			case *ast.PropertyKeyed:
				if p.Computed && MayHaveSideEffects(p.Key) {
					return false
				}
				return !MayHaveSideEffects(p.Value)
			case *ast.SpreadElement:
				hasSpread = true
				return false
			}
			return true
		})
		if hasSpread {
			*to = append(*to, *expr)
		} else {
			for _, prop := range e.Value {
				switch p := prop.Prop.(type) {
				case *ast.PropertyShort:
				case *ast.PropertyKeyed:
					if p.Computed {
						ExtractSideEffectsTo(to, p.Key)
					}
					ExtractSideEffectsTo(to, p.Value)
				}
			}
		}
	case *ast.ArrayLiteral:
		for _, elem := range e.Value {
			ExtractSideEffectsTo(to, &elem)
		}
	case *ast.TemplateLiteral:
		for _, elem := range e.Expressions {
			ExtractSideEffectsTo(to, &elem)
		}
	case *ast.ClassLiteral:
		panic("add_effects for class expression")
	case *ast.OptionalChain:
		*to = append(*to, *expr)
	}
}

// PropNameEq returns true if the property name of the expression is equal to key.
func PropNameEq(p *ast.Expression, key string) bool {
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
