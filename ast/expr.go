package ast

import "github.com/t14raptor/go-fast/token"

type (
	Expressions []Expression

	// Expression is a struct to allow defining methods on it.
	Expression struct {
		Expr `optional:"true"`
	}

	// All expression nodes implement the Expr interface.
	Expr interface {
		Node
		VisitableNode
		_expr()
	}

	BindingTarget struct {
		Target
	}

	Target interface {
		Expr
		_bindingTarget()
	}

	Pattern interface {
		Target
		_pattern()
	}

	YieldExpression struct {
		Yield    Idx
		Argument *Expression
		Delegate bool
	}

	AwaitExpression struct {
		Await    Idx
		Argument *Expression
	}

	ArrayLiteral struct {
		LeftBracket  Idx
		RightBracket Idx
		Value        Expressions
	}

	ArrayPattern struct {
		LeftBracket  Idx
		RightBracket Idx
		Elements     Expressions
		Rest         *Expression
	}

	AssignExpression struct {
		Operator token.Token
		Left     *Expression
		Right    *Expression
	}

	InvalidExpression struct {
		From Idx
		To   Idx
	}

	BinaryExpression struct {
		Operator token.Token
		Left     *Expression
		Right    *Expression
	}

	MemberExpression struct {
		Object   *Expression
		Property *Expression
	}

	CallExpression struct {
		Callee           *Expression
		LeftParenthesis  Idx
		ArgumentList     Expressions
		RightParenthesis Idx
	}

	ConditionalExpression struct {
		Test       *Expression
		Consequent *Expression
		Alternate  *Expression
	}

	PrivateDotExpression struct {
		Left       *Expression
		Identifier PrivateIdentifier
	}

	OptionalChain struct {
		Base *Expression
	}

	Optional struct {
		Expr *Expression
	}

	ConciseBody struct {
		Body Body
	}

	Body interface {
		Node
		VisitableNode
		_conciseBody()
	}

	ArrowFunctionLiteral struct {
		Start         Idx
		ParameterList ParameterList
		Body          *ConciseBody
		Async         bool

		ScopeContext ScopeContext
	}

	PrivateIdentifier struct {
		Identifier *Identifier
	}

	NewExpression struct {
		New              Idx
		Callee           *Expression
		LeftParenthesis  Idx
		ArgumentList     Expressions
		RightParenthesis Idx
	}

	ObjectLiteral struct {
		LeftBrace  Idx
		RightBrace Idx
		Value      Properties
	}

	ObjectPattern struct {
		LeftBrace  Idx
		RightBrace Idx
		Properties Properties
		Rest       Expr `optional:"true"`
	}

	SpreadElement struct {
		Expression Expression
	}

	SequenceExpression struct {
		Sequence Expressions
	}

	TemplateElements []TemplateElement

	TemplateElement struct {
		Idx     Idx
		Literal string
		Parsed  string
		Valid   bool
	}

	TemplateLiteral struct {
		OpenQuote   Idx
		CloseQuote  Idx
		Tag         *Expression `optional:"true"`
		Elements    TemplateElements
		Expressions Expressions
	}

	ThisExpression struct {
		Idx Idx
	}

	SuperExpression struct {
		Idx Idx
	}

	UnaryExpression struct {
		Operator token.Token
		Idx      Idx
		Operand  *Expression
	}

	UpdateExpression struct {
		Operator token.Token
		Idx      Idx // If a prefix operation
		Operand  *Expression
		Postfix  bool
	}

	MetaProperty struct {
		Meta, Property *Identifier
		Idx            Idx
	}
)

func (*BlockStatement) _conciseBody() {}
func (*Expression) _conciseBody()     {}

func (*ArrayPattern) _pattern()  {}
func (*ObjectPattern) _pattern() {}

func (*ArrayPattern) _bindingTarget()      {}
func (*MemberExpression) _bindingTarget()  {}
func (*ObjectPattern) _bindingTarget()     {}
func (*Identifier) _bindingTarget()        {}
func (*InvalidExpression) _bindingTarget() {}

func (*ArrayLiteral) _expr()          {}
func (*AssignExpression) _expr()      {}
func (*YieldExpression) _expr()       {}
func (*AwaitExpression) _expr()       {}
func (*InvalidExpression) _expr()     {}
func (*BinaryExpression) _expr()      {}
func (*CallExpression) _expr()        {}
func (*ConditionalExpression) _expr() {}
func (*MemberExpression) _expr()      {}
func (*PrivateDotExpression) _expr()  {}
func (*ArrowFunctionLiteral) _expr()  {}
func (*NewExpression) _expr()         {}
func (*ObjectLiteral) _expr()         {}
func (*SequenceExpression) _expr()    {}
func (*TemplateLiteral) _expr()       {}
func (*ThisExpression) _expr()        {}
func (*SuperExpression) _expr()       {}
func (*UnaryExpression) _expr()       {}
func (*UpdateExpression) _expr()      {}
func (*MetaProperty) _expr()          {}
func (*ObjectPattern) _expr()         {}
func (*ArrayPattern) _expr()          {}
func (*VariableDeclarator) _expr()    {}
func (*OptionalChain) _expr()         {}
func (*Optional) _expr()              {}
func (*SpreadElement) _expr()         {}
func (*PrivateIdentifier) _expr()     {}
