package ast

type (
	Expressions []Expression

	// Expression is a struct to allow defining methods on it.
	Expression struct {
		Expr Expr `optional:"true"`
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

	ArrayPattern struct {
		Elements Expressions
		Rest     *Expression

		LeftBracket  Idx
		RightBracket Idx
	}

	AssignExpression struct {
		Left  *Expression
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
		Prop MemberProp
	}

	MemberProp interface {
		VisitableNode
		_memberProperty()
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

	ConciseBody struct {
		Body Body
	}

	Body interface {
		Node
		VisitableNode
		_conciseBody()
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
		Rest       Expr `optional:"true"`

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

		Operator UnaryOperator

		Idx Idx
	}

	UpdateExpression struct {
		Operand *Expression

		Operator UpdateOperator
		Postfix  bool

		Idx Idx // If a prefix operation
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
func (*LogicalExpression) _expr()     {}
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
