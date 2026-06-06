package ast

import "unsafe"

type (
	Expressions []Expression

	//union:ArrayLiteral,ArrowFunctionLiteral,AssignExpression,AwaitExpression,BigIntLiteral,BinaryExpression,BooleanLiteral,CallExpression,ClassLiteral,ConditionalExpression,FunctionLiteral,Identifier,InvalidExpression,LogicalExpression,MemberExpression,MetaProperty,NewExpression,NullLiteral,NumberLiteral,ObjectLiteral,OptionalChain,Optional,PrivateDotExpression,PrivateIdentifier,RegExpLiteral,SequenceExpression,SpreadElement,StringLiteral,SuperExpression,ThisExpression,TemplateLiteral,UnaryExpression,UpdateExpression,VariableDeclarator,YieldExpression
	Expression struct {
		kind ExprKind

		ptr unsafe.Pointer
	}

	YieldExpression struct {
		Argument *Expression

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

	//union:ComputedProperty,Identifier
	MemberProperty struct {
		ptr  unsafe.Pointer
		kind MemPropKind
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

	//union:BlockStatement,Expression
	ConciseBody struct {
		kind ConciseBodyKind
		ptr  unsafe.Pointer
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

	MetaProperty struct {
		Meta, Property *Identifier
		Idx            Idx
	}
)

// ExpressionFromPattern unwraps the simple-target variants of a pattern
// (identifier, member, private-dot, invalid) back into an Expression. It is used
// by the generator, which emits those leaf targets through the expression path.
// Array/object/assignment patterns are emitted directly and are not handled here.
func ExpressionFromPattern(p *Pattern) Expression {
	switch p.Kind() {
	case PatternIdent:
		return NewIdentExpr((*Identifier)(p.ptr))
	case PatternMember:
		return NewMemberExpr((*MemberExpression)(p.ptr))
	case PatternPrivDot:
		return NewPrivDotExpr((*PrivateDotExpression)(p.ptr))
	case PatternInvalid:
		return NewInvalidExpr((*InvalidExpression)(p.ptr))
	}
	return Expression{}
}
