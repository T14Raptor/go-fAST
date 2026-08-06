package ast

import "unsafe"

type (
	Expressions []Expression

	Expression struct {
		kind ExprKind
		ptr  unsafe.Pointer
	}

	// Expression union variants (gen_union.go): field = tag, type = payload.
	//union:Expression
	_ struct {
		ArrayLit       ArrayLiteral
		ArrowFuncLit   ArrowFunctionLiteral
		Assign         AssignExpression
		Await          AwaitExpression
		BigIntLit      BigIntLiteral
		Binary         BinaryExpression
		BoolLit        BooleanLiteral
		Call           CallExpression
		ClassLit       ClassLiteral
		Conditional    ConditionalExpression
		FuncLit        FunctionLiteral
		Identifier     Identifier
		Logical        LogicalExpression
		Member         MemberExpression
		MetaProp       MetaProperty
		New            NewExpression
		NullLit        NullLiteral
		NumberLit      NumberLiteral
		ObjectLit      ObjectLiteral
		Optional       Optional
		OptionalChain  OptionalChain
		Paren          ParenthesizedExpression
		PrivDot        PrivateDotExpression
		PrivIdentifier PrivateIdentifier
		RegExpLit      RegExpLiteral
		Sequence       SequenceExpression
		Spread         SpreadElement
		StringLit      StringLiteral
		Super          SuperExpression
		This           ThisExpression
		TmplLit        TemplateLiteral
		Unary          UnaryExpression
		Update         UpdateExpression
		Yield          YieldExpression
		Invalid        InvalidExpression
	}

	YieldExpression struct {
		Argument *Expression `optional:"true"`

		Yield Idx

		Delegate bool
	}

	AwaitExpression struct {
		Argument *Expression

		Await Idx
	}

	ArrayLiteral struct {
		Value Expressions

		LeftBracket  Idx
		RightBracket Idx
	}

	AssignExpression struct {
		Left  *Pattern
		Right *Expression

		Operator AssignmentOperator
	}

	InvalidExpression struct {
		From Idx
		To   Idx
	}

	BinaryExpression struct {
		Left  *Expression
		Right *Expression

		Operator BinaryOperator
	}

	LogicalExpression struct {
		Left  *Expression
		Right *Expression

		Operator LogicalOperator
	}

	MemberExpression struct {
		Object   *Expression
		Property *MemberProperty
	}

	MemberProperty struct {
		kind MemPropKind
		ptr  unsafe.Pointer
	}

	// MemberProperty union variants (gen_union.go): field = tag, type = payload.
	//union:MemberProperty
	_ struct {
		Computed   ComputedProperty
		Identifier Identifier
	}

	CallExpression struct {
		Callee       *Expression
		ArgumentList Expressions

		LeftParenthesis  Idx
		RightParenthesis Idx
	}

	ConditionalExpression struct {
		Test       *Expression
		Consequent *Expression
		Alternate  *Expression
	}

	// ParenthesizedExpression is a grouping `( Expression )`. It is only
	// produced when parsing with [parser.PreserveParens]; by default the
	// parser discards grouping parentheses and yields the inner expression
	// directly. Keeping the node makes a parenthesized expression's source
	// span balanced, since the inner node's own offsets exclude the parens.
	ParenthesizedExpression struct {
		Expression *Expression

		LeftParenthesis  Idx
		RightParenthesis Idx
	}

	PrivateDotExpression struct {
		Left       *Expression
		Identifier *PrivateIdentifier
	}

	OptionalChain struct {
		Base *Expression
	}

	Optional struct {
		Expr *Expression
	}

	ConciseBody struct {
		kind ConciseBodyKind
		ptr  unsafe.Pointer
	}

	// ConciseBody union variants (gen_union.go): field = tag, type = payload.
	//union:ConciseBody
	_ struct {
		Block BlockStatement
		Expr  Expression
	}

	ArrowFunctionLiteral struct {
		ParameterList *ParameterList
		Body          *ConciseBody

		ScopeContext ScopeContext

		Start Idx
		Async bool
	}

	PrivateIdentifier struct {
		Identifier *Identifier
	}

	NewExpression struct {
		Callee       *Expression
		ArgumentList Expressions

		New              Idx
		LeftParenthesis  Idx
		RightParenthesis Idx
	}

	ObjectLiteral struct {
		Value Properties

		LeftBrace  Idx
		RightBrace Idx
	}

	SpreadElement struct {
		Expression *Expression
	}

	SequenceExpression struct {
		Sequence Expressions
	}

	TemplateElements []TemplateElement

	TemplateElement struct {
		Literal string
		Parsed  string

		Idx Idx
	}

	TemplateLiteral struct {
		Tag         *Expression `optional:"true"`
		Elements    TemplateElements
		Expressions Expressions

		OpenQuote  Idx
		CloseQuote Idx
	}

	ThisExpression struct {
		Idx Idx
	}

	SuperExpression struct {
		Idx Idx
	}

	UnaryExpression struct {
		Operand *Expression

		Idx Idx

		Operator UnaryOperator
	}

	UpdateExpression struct {
		Operand *Expression

		Idx Idx // If a prefix operation

		Operator UpdateOperator
		Postfix  bool
	}

	// MetaProperty is a meta-property: new.target or import.meta.
	MetaProperty struct {
		Kind MetaPropertyKind
		Idx  Idx
	}
)

// MetaPropertyKind discriminates the meta-properties new.target and import.meta.
type MetaPropertyKind uint8

const (
	MetaPropertyNewTarget  MetaPropertyKind = iota // new.target
	MetaPropertyImportMeta                         // import.meta
)

func (k MetaPropertyKind) String() string {
	switch k {
	case MetaPropertyNewTarget:
		return "new.target"
	case MetaPropertyImportMeta:
		return "import.meta"
	}
	return "MetaPropertyKind(?)"
}

// ExpressionFromPattern unwraps the simple-target variants of a pattern
// (identifier, member, private-dot, invalid) back into an Expression. It is used
// by the generator, which emits those leaf targets through the expression path.
// Array/object/assignment patterns are emitted directly and are not handled here.
func ExpressionFromPattern(p *Pattern) Expression {
	switch p.Kind() {
	case PatternIdentifier:
		return NewIdentifierExpr((*Identifier)(p.ptr))
	case PatternMember:
		return NewMemberExpr((*MemberExpression)(p.ptr))
	case PatternPrivDot:
		return NewPrivDotExpr((*PrivateDotExpression)(p.ptr))
	case PatternInvalid:
		return NewInvalidExpr((*InvalidExpression)(p.ptr))
	}
	return Expression{}
}

func (o *Optional) Idx0() Idx { return o.Expr.Idx0() }
func (o *Optional) Idx1() Idx { return o.Expr.Idx1() }

func (n *OptionalChain) Idx0() Idx { return n.Base.Idx0() }
func (n *OptionalChain) Idx1() Idx { return n.Base.Idx1() }

func (a *ArrayLiteral) Idx0() Idx { return a.LeftBracket }
func (a *ArrayLiteral) Idx1() Idx { return a.RightBracket + 1 }

func (y *YieldExpression) Idx0() Idx { return y.Yield }
func (y *YieldExpression) Idx1() Idx {
	if y.Argument != nil {
		return y.Argument.Idx1()
	}
	return y.Yield + 5
}

func (a *AwaitExpression) Idx0() Idx { return a.Await }
func (a *AwaitExpression) Idx1() Idx { return a.Argument.Idx1() }

func (a *AssignExpression) Idx0() Idx { return a.Left.Idx0() }
func (a *AssignExpression) Idx1() Idx { return a.Right.Idx1() }

func (b *BinaryExpression) Idx0() Idx { return b.Left.Idx0() }
func (b *BinaryExpression) Idx1() Idx { return b.Right.Idx1() }

func (b *LogicalExpression) Idx0() Idx { return b.Left.Idx0() }
func (b *LogicalExpression) Idx1() Idx { return b.Right.Idx1() }

func (n *CallExpression) Idx0() Idx { return n.Callee.Idx0() }
func (n *CallExpression) Idx1() Idx { return n.RightParenthesis + 1 }

func (n *ConditionalExpression) Idx0() Idx { return n.Test.Idx0() }
func (n *ConditionalExpression) Idx1() Idx { return n.Alternate.Idx1() }

func (n *ParenthesizedExpression) Idx0() Idx { return n.LeftParenthesis }
func (n *ParenthesizedExpression) Idx1() Idx { return n.RightParenthesis + 1 }

func (p *PrivateDotExpression) Idx0() Idx { return p.Left.Idx0() }
func (p *PrivateDotExpression) Idx1() Idx { return p.Identifier.Idx1() }

func (a *ArrowFunctionLiteral) Idx0() Idx { return a.Start }
func (a *ArrowFunctionLiteral) Idx1() Idx { return a.Body.Idx1() }

func (n *InvalidExpression) Idx0() Idx { return n.From }
func (n *InvalidExpression) Idx1() Idx { return n.To }

func (n *NewExpression) Idx0() Idx { return n.New }
func (n *NewExpression) Idx1() Idx {
	if n.ArgumentList != nil {
		return n.RightParenthesis + 1
	}
	return n.Callee.Idx1()
}

func (n *ObjectLiteral) Idx0() Idx { return n.LeftBrace }
func (n *ObjectLiteral) Idx1() Idx { return n.RightBrace + 1 }

func (n *SequenceExpression) Idx0() Idx { return n.Sequence[0].Idx0() }
func (n *SequenceExpression) Idx1() Idx { return n.Sequence[len(n.Sequence)-1].Idx1() }

func (n *TemplateElement) Idx0() Idx { return n.Idx }
func (n *TemplateElement) Idx1() Idx { return n.Idx + Idx(len(n.Literal)) }

func (n *TemplateLiteral) Idx0() Idx { return n.OpenQuote }
func (n *TemplateLiteral) Idx1() Idx { return n.CloseQuote + 1 }

func (n *ThisExpression) Idx0() Idx { return n.Idx }
func (n *ThisExpression) Idx1() Idx { return n.Idx + 4 }

func (n *SuperExpression) Idx0() Idx { return n.Idx }
func (n *SuperExpression) Idx1() Idx { return n.Idx + 5 }

func (n *UnaryExpression) Idx0() Idx { return n.Idx }
func (n *UnaryExpression) Idx1() Idx { return n.Operand.Idx1() }

func (n *UpdateExpression) Idx0() Idx { return n.Idx }
func (n *UpdateExpression) Idx1() Idx {
	if n.Postfix {
		return n.Operand.Idx1() + 2 // x++ x--
	}
	return n.Operand.Idx1()
}

func (n *MetaProperty) Idx0() Idx { return n.Idx }
func (n *MetaProperty) Idx1() Idx { return n.Idx + Idx(len(n.Kind.String())) }

func (m *MemberExpression) Idx0() Idx { return m.Object.Idx0() }
func (m *MemberExpression) Idx1() Idx { return m.Property.Idx1() }

func (n *SpreadElement) Idx0() Idx { return n.Expression.Idx0() }
func (n *SpreadElement) Idx1() Idx { return n.Expression.Idx1() }

func (n *PrivateIdentifier) Idx0() Idx { return n.Identifier.Idx0() }
func (n *PrivateIdentifier) Idx1() Idx { return n.Identifier.Idx1() }
