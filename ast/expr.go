package ast

import (
	"github.com/t14raptor/go-fast/token"
	"unsafe"
)

type (
	Expressions []Expression

	//union:ArrayLiteral,ArrayPattern,ArrowFunctionLiteral,AssignExpression,AwaitExpression,BinaryExpression,BooleanLiteral,CallExpression,ClassLiteral,ConditionalExpression,FunctionLiteral,Identifier,InvalidExpression,MemberExpression,MetaProperty,NewExpression,NullLiteral,NumberLiteral,ObjectLiteral,ObjectPattern,OptionalChain,Optional,PrivateDotExpression,PrivateIdentifier,PropertyKeyed,PropertyShort,RegExpLiteral,SequenceExpression,SpreadElement,StringLiteral,SuperExpression,ThisExpression,TemplateLiteral,UnaryExpression,UpdateExpression,VariableDeclarator,YieldExpression
	Expression struct {
		ptr unsafe.Pointer

		kind ExprKind
	}

	//union:ArrayPattern,Identifier,InvalidExpression,MemberExpression,ObjectPattern
	BindingTarget struct {
		ptr  unsafe.Pointer
		kind BindingTargetKind
	}

	YieldExpression struct {
		Yield    Idx
		Argument *Expression
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

	ArrayPattern struct {
		Elements Expressions
		Rest     *Expression

		LeftBracket  Idx
		RightBracket Idx
	}

	AssignExpression struct {
		Left  *Expression
		Right *Expression

		Operator token.Token
	}

	InvalidExpression struct {
		From Idx
		To   Idx
	}

	BinaryExpression struct {
		Left  *Expression
		Right *Expression

		Operator token.Token
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
		ptr  unsafe.Pointer
		kind ConciseBodyKind
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

	ObjectPattern struct {
		Properties Properties
		Rest       *Expression `optional:"true"`

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

		Idx   Idx
		Valid bool
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

		Operator token.Token
	}

	UpdateExpression struct {
		Operand *Expression

		Idx Idx // If a prefix operation

		Operator token.Token
		Postfix  bool
	}

	MetaProperty struct {
		Meta, Property *Identifier
		Idx            Idx
	}
)

func BindingTargetFromExpression(expr *Expression) BindingTarget {
	switch expr.Kind() {
	case ExprArrPat:
		return NewArrPatBindingTarget((*ArrayPattern)(expr.ptr))
	case ExprMember:
		return NewMemberBindingTarget((*MemberExpression)(expr.ptr))
	case ExprObjPat:
		return NewObjPatBindingTarget((*ObjectPattern)(expr.ptr))
	case ExprIdent:
		return NewIdentBindingTarget((*Identifier)(expr.ptr))
	case ExprInvalid:
		return NewInvalidBindingTarget((*InvalidExpression)(expr.ptr))
	}
	return BindingTarget{}
}

func (bt *BindingTarget) IsPattern() bool {
	return bt.kind == BindingTargetArrPat || bt.kind == BindingTargetObjPat
}
