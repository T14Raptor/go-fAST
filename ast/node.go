package ast

import (
	"github.com/t14raptor/go-fast/token"
	"github.com/t14raptor/go-fast/unistring"
)

type PropertyKind string

const (
	PropertyKindValue  PropertyKind = "value"
	PropertyKindGet    PropertyKind = "get"
	PropertyKindSet    PropertyKind = "set"
	PropertyKindMethod PropertyKind = "method"
)

// Idx is a compact encoding of a source position within JS code.
type Idx int

// Node ...
type Node interface {
	// Idx0 returns the index of the first character belonging to the node.
	Idx0() Idx
	// Idx1 returns the index of the first character immediately after the node.
	Idx1() Idx
}

type (
	Expressions []Expression
	Statements  []Statement
	// Expression is a struct to allow defining methods on it.
	Expression struct {
		Expr
	}
	Statement struct {
		Stmt
	}
)

type (
	// All expression nodes implement the Expr interface.
	Expr interface {
		Node
		VisitableNode
		_expr()
	}

	BindingTarget interface {
		Expr
		_bindingTarget()
	}

	VariableDeclarators []*VariableDeclarator

	VariableDeclarator struct {
		Target      BindingTarget
		Initializer *Expression
	}

	Pattern interface {
		BindingTarget
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
		Operator   token.Token
		Left       *Expression
		Right      *Expression
		Comparison bool
	}

	BooleanLiteral struct {
		Idx     Idx
		Literal string
		Value   bool
	}

	// DEPRECATED: use MemberExpression
	BracketExpression struct {
		Left         *Expression
		Member       *Expression
		LeftBracket  Idx
		RightBracket Idx
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

	// DEPRECATED: use MemberExpression
	DotExpression struct {
		Left       *Expression
		Identifier Identifier
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

	FunctionLiteral struct {
		Function      Idx
		Name          *Identifier
		ParameterList ParameterList
		Body          *BlockStatement
		Source        string

		Async, Generator bool
	}

	ClassLiteral struct {
		Class      Idx
		RightBrace Idx
		Name       *Identifier
		SuperClass *Expression
		Body       []ClassElement
		Source     string
	}

	ConciseBody interface {
		Node
		_conciseBody()
	}

	ExpressionBody struct {
		Expression Expression
	}

	ArrowFunctionLiteral struct {
		Start         Idx
		ParameterList ParameterList
		Body          ConciseBody
		Source        string
		Async         bool
	}

	Identifier struct {
		Idx  Idx
		Name unistring.String
	}

	PrivateIdentifier struct {
		*Identifier
	}

	NewExpression struct {
		New              Idx
		Callee           *Expression
		LeftParenthesis  Idx
		ArgumentList     Expressions
		RightParenthesis Idx
	}

	NullLiteral struct {
		Idx     Idx
		Literal string
	}

	NumberLiteral struct {
		Idx     Idx
		Literal string
		Value   any
	}

	ObjectLiteral struct {
		LeftBrace  Idx
		RightBrace Idx
		Value      []Property
	}

	ObjectPattern struct {
		LeftBrace  Idx
		RightBrace Idx
		Properties []Property
		Rest       Expr
	}

	ParameterList struct {
		Opening Idx
		List    VariableDeclarators
		Rest    Expr
		Closing Idx
	}

	Property interface {
		Expr
		_property()
	}

	PropertyShort struct {
		Name        *Identifier
		Initializer *Expression
	}

	PropertyKeyed struct {
		Key      *Expression
		Kind     PropertyKind
		Value    *Expression
		Computed bool
	}

	SpreadElement struct {
		Expression Expression
	}

	RegExpLiteral struct {
		Idx     Idx
		Literal string
		Pattern string
		Flags   string
	}

	SequenceExpression struct {
		Sequence Expressions
	}

	StringLiteral struct {
		Idx     Idx
		Literal string
		Value   unistring.String
	}

	TemplateElement struct {
		Idx     Idx
		Literal string
		Parsed  unistring.String
		Valid   bool
	}

	TemplateLiteral struct {
		OpenQuote   Idx
		CloseQuote  Idx
		Tag         *Expression
		Elements    []TemplateElement
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
		Idx      Idx // If a prefix operation
		Operand  *Expression
		Postfix  bool
	}

	MetaProperty struct {
		Meta, Property *Identifier
		Idx            Idx
	}
)

func (n SpreadElement) Idx0() Idx {
	return n.Expression.Expr.Idx0()
}

func (n SpreadElement) Idx1() Idx {
	return n.Expression.Expr.Idx1()
}

func (n SpreadElement) _expr() {}

func (MemberExpression) _bindingTarget() {}

// _expr

func (ArrayLiteral) _expr()          {}
func (AssignExpression) _expr()      {}
func (YieldExpression) _expr()       {}
func (AwaitExpression) _expr()       {}
func (InvalidExpression) _expr()     {}
func (BinaryExpression) _expr()      {}
func (BooleanLiteral) _expr()        {}
func (BracketExpression) _expr()     {}
func (CallExpression) _expr()        {}
func (ConditionalExpression) _expr() {}
func (DotExpression) _expr()         {}
func (MemberExpression) _expr()      {}
func (PrivateDotExpression) _expr()  {}
func (FunctionLiteral) _expr()       {}
func (ClassLiteral) _expr()          {}
func (ArrowFunctionLiteral) _expr()  {}
func (Identifier) _expr()            {}
func (NewExpression) _expr()         {}
func (NullLiteral) _expr()           {}
func (NumberLiteral) _expr()         {}
func (ObjectLiteral) _expr()         {}
func (RegExpLiteral) _expr()         {}
func (SequenceExpression) _expr()    {}
func (StringLiteral) _expr()         {}
func (TemplateLiteral) _expr()       {}
func (ThisExpression) _expr()        {}
func (SuperExpression) _expr()       {}
func (UnaryExpression) _expr()       {}
func (MetaProperty) _expr()          {}
func (ObjectPattern) _expr()         {}
func (ArrayPattern) _expr()          {}
func (VariableDeclarator) _expr()    {}
func (OptionalChain) _expr()         {}
func (Optional) _expr()              {}
func (PropertyShort) _expr()         {}
func (PropertyKeyed) _expr()         {}

// ========= //
// Stmt //
// ========= //

type (
	// All statement nodes implement the Stmt interface.
	Stmt interface {
		Node
		VisitableNode
		_statementNode()
	}

	BadStatement struct {
		From Idx
		To   Idx
	}

	BlockStatement struct {
		LeftBrace  Idx
		List       Statements
		RightBrace Idx
	}

	BranchStatement struct {
		Idx   Idx
		Token token.Token
		Label *Identifier
	}

	CaseStatement struct {
		Case       Idx
		Test       *Expression
		Consequent Statements
	}

	CatchStatement struct {
		Catch     Idx
		Parameter *BindingTarget
		Body      *BlockStatement
	}

	DebuggerStatement struct {
		Debugger Idx
	}

	DoWhileStatement struct {
		Do   Idx
		Test *Expression
		Body *Statement
	}

	EmptyStatement struct {
		Semicolon Idx
	}

	ExpressionStatement struct {
		Expression *Expression
		Comment    string
	}

	ForInStatement struct {
		For    Idx
		Into   *ForInto
		Source *Expression
		Body   *Statement
	}

	ForOfStatement struct {
		For    Idx
		Into   *ForInto
		Source *Expression
		Body   *Statement
	}

	ForStatement struct {
		For         Idx
		Initializer *ForLoopInitializer
		Update      *Expression
		Test        *Expression
		Body        *Statement
	}

	IfStatement struct {
		If         Idx
		Test       *Expression
		Consequent *Statement
		Alternate  *Statement
	}

	LabelledStatement struct {
		Label     *Identifier
		Colon     Idx
		Statement *Statement
	}

	ReturnStatement struct {
		Return   Idx
		Argument *Expression
	}

	SwitchStatement struct {
		Switch       Idx
		Discriminant *Expression
		Default      int
		Body         []CaseStatement
	}

	ThrowStatement struct {
		Throw    Idx
		Argument *Expression
	}

	TryStatement struct {
		Try     Idx
		Body    *BlockStatement
		Catch   *CatchStatement
		Finally *BlockStatement
	}

	VariableStatement struct {
		Var  Idx
		List VariableDeclarators
	}

	LexicalDeclaration struct {
		Idx     Idx
		Token   token.Token
		List    VariableDeclarators
		Comment string
	}

	WhileStatement struct {
		While Idx
		Test  *Expression
		Body  *Statement
	}

	WithStatement struct {
		With   Idx
		Object *Expression
		Body   *Statement
	}

	FunctionDeclaration struct {
		Function *FunctionLiteral
	}

	ClassDeclaration struct {
		Class *ClassLiteral
	}
)

// _statementNode

func (BadStatement) _statementNode()        {}
func (BlockStatement) _statementNode()      {}
func (BranchStatement) _statementNode()     {}
func (CaseStatement) _statementNode()       {}
func (CatchStatement) _statementNode()      {}
func (DebuggerStatement) _statementNode()   {}
func (DoWhileStatement) _statementNode()    {}
func (EmptyStatement) _statementNode()      {}
func (ExpressionStatement) _statementNode() {}
func (ForInStatement) _statementNode()      {}
func (ForOfStatement) _statementNode()      {}
func (ForStatement) _statementNode()        {}
func (IfStatement) _statementNode()         {}
func (LabelledStatement) _statementNode()   {}
func (ReturnStatement) _statementNode()     {}
func (SwitchStatement) _statementNode()     {}
func (ThrowStatement) _statementNode()      {}
func (TryStatement) _statementNode()        {}
func (VariableStatement) _statementNode()   {}
func (WhileStatement) _statementNode()      {}
func (WithStatement) _statementNode()       {}
func (LexicalDeclaration) _statementNode()  {}
func (FunctionDeclaration) _statementNode() {}
func (ClassDeclaration) _statementNode()    {}

// =========== //
// Declaration //
// =========== //

type (
	ClassElement interface {
		Node
		_classElement()
	}

	FieldDefinition struct {
		Idx         Idx
		Key         *Expression
		Initializer *Expression
		Computed    bool
		Static      bool
	}

	MethodDefinition struct {
		Idx      Idx
		Key      *Expression
		Kind     PropertyKind // "method", "get" or "set"
		Body     *FunctionLiteral
		Computed bool
		Static   bool
	}

	ClassStaticBlock struct {
		Static Idx
		Block  *BlockStatement
		Source string
	}
)

type (
	ForLoopInitializer interface {
		Node
		_forLoopInitializer()
	}

	ForLoopInitializerExpression struct {
		Expression *Expression
	}

	ForLoopInitializerVarDeclList struct {
		List VariableDeclarators
	}

	ForLoopInitializerLexicalDecl struct {
		LexicalDeclaration LexicalDeclaration
	}

	ForInto interface {
		Node
		_forInto()
	}

	ForIntoVar struct {
		Binding *VariableDeclarator
	}

	ForDeclaration struct {
		Idx     Idx
		IsConst bool
		Target  BindingTarget
	}

	ForIntoExpression struct {
		Expression *Expression
	}
)

func (ForLoopInitializerExpression) _forLoopInitializer()  {}
func (ForLoopInitializerVarDeclList) _forLoopInitializer() {}
func (ForLoopInitializerLexicalDecl) _forLoopInitializer() {}

func (ForIntoVar) _forInto()        {}
func (ForDeclaration) _forInto()    {}
func (ForIntoExpression) _forInto() {}

func (ArrayPattern) _pattern()       {}
func (ArrayPattern) _bindingTarget() {}

func (ObjectPattern) _pattern()       {}
func (ObjectPattern) _bindingTarget() {}

func (InvalidExpression) _bindingTarget() {}

func (PropertyShort) _property() {}
func (PropertyKeyed) _property() {}
func (SpreadElement) _property() {}

func (Identifier) _bindingTarget() {}

func (BlockStatement) _conciseBody() {}
func (ExpressionBody) _conciseBody() {}

func (FieldDefinition) _classElement()  {}
func (MethodDefinition) _classElement() {}
func (ClassStaticBlock) _classElement() {}

// ==== //
// Node //
// ==== //

type Program struct {
	Body Statements
}

// ==== //
// Idx0 //
// ==== //

func (self Optional) Idx0() Idx              { return (*self.Expr).Expr.Idx0() }
func (self OptionalChain) Idx0() Idx         { return (*self.Base).Expr.Idx0() }
func (self ObjectPattern) Idx0() Idx         { return self.LeftBrace }
func (self ParameterList) Idx0() Idx         { return self.Opening }
func (self ArrayLiteral) Idx0() Idx          { return self.LeftBracket }
func (self ArrayPattern) Idx0() Idx          { return self.LeftBracket }
func (self YieldExpression) Idx0() Idx       { return self.Yield }
func (self AwaitExpression) Idx0() Idx       { return self.Await }
func (self AssignExpression) Idx0() Idx      { return (*self.Left).Expr.Idx0() }
func (self BinaryExpression) Idx0() Idx      { return (*self.Left).Expr.Idx0() }
func (self BooleanLiteral) Idx0() Idx        { return self.Idx }
func (self BracketExpression) Idx0() Idx     { return (*self.Left).Expr.Idx0() }
func (self CallExpression) Idx0() Idx        { return (*self.Callee).Expr.Idx0() }
func (self ConditionalExpression) Idx0() Idx { return (*self.Test).Expr.Idx0() }
func (self DotExpression) Idx0() Idx         { return (*self.Left).Expr.Idx0() }
func (self PrivateDotExpression) Idx0() Idx  { return (*self.Left).Expr.Idx0() }
func (self FunctionLiteral) Idx0() Idx       { return self.Function }
func (self ClassLiteral) Idx0() Idx          { return self.Class }
func (self ArrowFunctionLiteral) Idx0() Idx  { return self.Start }
func (self Identifier) Idx0() Idx            { return self.Idx }
func (self InvalidExpression) Idx0() Idx     { return self.From }
func (self NewExpression) Idx0() Idx         { return self.New }
func (self NullLiteral) Idx0() Idx           { return self.Idx }
func (self NumberLiteral) Idx0() Idx         { return self.Idx }
func (self ObjectLiteral) Idx0() Idx         { return self.LeftBrace }
func (self RegExpLiteral) Idx0() Idx         { return self.Idx }
func (self SequenceExpression) Idx0() Idx    { return self.Sequence[0].Expr.Idx0() }
func (self StringLiteral) Idx0() Idx         { return self.Idx }
func (self TemplateElement) Idx0() Idx       { return self.Idx }
func (self TemplateLiteral) Idx0() Idx       { return self.OpenQuote }
func (self ThisExpression) Idx0() Idx        { return self.Idx }
func (self SuperExpression) Idx0() Idx       { return self.Idx }
func (self UnaryExpression) Idx0() Idx       { return self.Idx }
func (self MetaProperty) Idx0() Idx          { return self.Idx }
func (self MemberExpression) Idx0() Idx      { return 0 }
func (self MemberExpression) Idx1() Idx      { return 0 }

func (self BadStatement) Idx0() Idx        { return self.From }
func (self BlockStatement) Idx0() Idx      { return self.LeftBrace }
func (self BranchStatement) Idx0() Idx     { return self.Idx }
func (self CaseStatement) Idx0() Idx       { return self.Case }
func (self CatchStatement) Idx0() Idx      { return self.Catch }
func (self DebuggerStatement) Idx0() Idx   { return self.Debugger }
func (self DoWhileStatement) Idx0() Idx    { return self.Do }
func (self EmptyStatement) Idx0() Idx      { return self.Semicolon }
func (self ExpressionStatement) Idx0() Idx { return (*self.Expression).Expr.Idx0() }
func (self ForInStatement) Idx0() Idx      { return self.For }
func (self ForOfStatement) Idx0() Idx      { return self.For }
func (self ForStatement) Idx0() Idx        { return self.For }
func (self IfStatement) Idx0() Idx         { return self.If }
func (self LabelledStatement) Idx0() Idx   { return self.Label.Idx0() }
func (self Program) Idx0() Idx             { return self.Body[0].Idx0() }
func (self ReturnStatement) Idx0() Idx     { return self.Return }
func (self SwitchStatement) Idx0() Idx     { return self.Switch }
func (self ThrowStatement) Idx0() Idx      { return self.Throw }
func (self TryStatement) Idx0() Idx        { return self.Try }
func (self VariableStatement) Idx0() Idx   { return self.Var }
func (self WhileStatement) Idx0() Idx      { return self.While }
func (self WithStatement) Idx0() Idx       { return self.With }
func (self LexicalDeclaration) Idx0() Idx  { return self.Idx }
func (self FunctionDeclaration) Idx0() Idx { return self.Function.Idx0() }
func (self ClassDeclaration) Idx0() Idx    { return self.Class.Idx0() }
func (self VariableDeclarator) Idx0() Idx  { return self.Target.Idx0() }

func (self ForLoopInitializerExpression) Idx0() Idx  { return (*self.Expression).Expr.Idx0() }
func (self ForLoopInitializerVarDeclList) Idx0() Idx { return self.List[0].Idx0() }
func (self ForLoopInitializerLexicalDecl) Idx0() Idx { return self.LexicalDeclaration.Idx0() }
func (self PropertyShort) Idx0() Idx                 { return self.Name.Idx }
func (self PropertyKeyed) Idx0() Idx                 { return (*self.Key).Expr.Idx0() }
func (self ExpressionBody) Idx0() Idx                { return self.Expression.Expr.Idx0() }

func (self FieldDefinition) Idx0() Idx  { return self.Idx }
func (self MethodDefinition) Idx0() Idx { return self.Idx }
func (self ClassStaticBlock) Idx0() Idx { return self.Static }

func (self ForDeclaration) Idx0() Idx    { return self.Idx }
func (self ForIntoVar) Idx0() Idx        { return self.Binding.Idx0() }
func (self ForIntoExpression) Idx0() Idx { return (*self.Expression).Expr.Idx0() }

// ==== //
// Idx1 //
// ==== //

func (self Optional) Idx1() Idx              { return (*self.Expr).Expr.Idx1() }
func (self OptionalChain) Idx1() Idx         { return (*self.Base).Expr.Idx1() }
func (self ArrayLiteral) Idx1() Idx          { return self.RightBracket + 1 }
func (self ArrayPattern) Idx1() Idx          { return self.RightBracket + 1 }
func (self AssignExpression) Idx1() Idx      { return (*self.Right).Expr.Idx1() }
func (self AwaitExpression) Idx1() Idx       { return (*self.Argument).Expr.Idx1() }
func (self InvalidExpression) Idx1() Idx     { return self.To }
func (self BinaryExpression) Idx1() Idx      { return (*self.Right).Expr.Idx1() }
func (self BooleanLiteral) Idx1() Idx        { return Idx(int(self.Idx) + len(self.Literal)) }
func (self BracketExpression) Idx1() Idx     { return self.RightBracket + 1 }
func (self CallExpression) Idx1() Idx        { return self.RightParenthesis + 1 }
func (self ConditionalExpression) Idx1() Idx { return (*self.Test).Expr.Idx1() }
func (self DotExpression) Idx1() Idx         { return self.Identifier.Idx1() }
func (self PrivateDotExpression) Idx1() Idx  { return self.Identifier.Idx1() }
func (self FunctionLiteral) Idx1() Idx       { return self.Body.Idx1() }
func (self ClassLiteral) Idx1() Idx          { return self.RightBrace + 1 }
func (self ArrowFunctionLiteral) Idx1() Idx  { return self.Body.Idx1() }
func (self Identifier) Idx1() Idx            { return Idx(int(self.Idx) + len(self.Name)) }
func (self NewExpression) Idx1() Idx {
	if self.ArgumentList != nil {
		return self.RightParenthesis + 1
	} else {
		return (*self.Callee).Expr.Idx1()
	}
}
func (self NullLiteral) Idx1() Idx        { return Idx(int(self.Idx) + 4) } // "null"
func (self NumberLiteral) Idx1() Idx      { return Idx(int(self.Idx) + len(self.Literal)) }
func (self ObjectLiteral) Idx1() Idx      { return self.RightBrace + 1 }
func (self ObjectPattern) Idx1() Idx      { return self.RightBrace + 1 }
func (self ParameterList) Idx1() Idx      { return self.Closing + 1 }
func (self RegExpLiteral) Idx1() Idx      { return Idx(int(self.Idx) + len(self.Literal)) }
func (self SequenceExpression) Idx1() Idx { return self.Sequence[len(self.Sequence)-1].Expr.Idx1() }
func (self StringLiteral) Idx1() Idx      { return Idx(int(self.Idx) + len(self.Literal)) }
func (self TemplateElement) Idx1() Idx    { return Idx(int(self.Idx) + len(self.Literal)) }
func (self TemplateLiteral) Idx1() Idx    { return self.CloseQuote + 1 }
func (self ThisExpression) Idx1() Idx     { return self.Idx + 4 }
func (self SuperExpression) Idx1() Idx    { return self.Idx + 5 }
func (self UnaryExpression) Idx1() Idx {
	if self.Postfix {
		return (*self.Operand).Expr.Idx1() + 2 // ++ --
	}
	return (*self.Operand).Expr.Idx1()
}
func (self MetaProperty) Idx1() Idx {
	return self.Property.Idx1()
}

func (self BadStatement) Idx1() Idx        { return self.To }
func (self BlockStatement) Idx1() Idx      { return self.RightBrace + 1 }
func (self BranchStatement) Idx1() Idx     { return self.Idx }
func (self CaseStatement) Idx1() Idx       { return self.Consequent[len(self.Consequent)-1].Idx1() }
func (self CatchStatement) Idx1() Idx      { return self.Body.Idx1() }
func (self DebuggerStatement) Idx1() Idx   { return self.Debugger + 8 }
func (self DoWhileStatement) Idx1() Idx    { return (*self.Test).Expr.Idx1() }
func (self EmptyStatement) Idx1() Idx      { return self.Semicolon + 1 }
func (self ExpressionStatement) Idx1() Idx { return (*self.Expression).Expr.Idx1() }
func (self ForInStatement) Idx1() Idx      { return (*self.Body).Idx1() }
func (self ForOfStatement) Idx1() Idx      { return (*self.Body).Idx1() }
func (self ForStatement) Idx1() Idx        { return (*self.Body).Idx1() }
func (self IfStatement) Idx1() Idx {
	if self.Alternate != nil {
		return (*self.Alternate).Idx1()
	}
	return (*self.Consequent).Idx1()
}
func (self LabelledStatement) Idx1() Idx { return self.Colon + 1 }
func (self Program) Idx1() Idx           { return self.Body[len(self.Body)-1].Idx1() }
func (self ReturnStatement) Idx1() Idx   { return self.Return + 6 }
func (self SwitchStatement) Idx1() Idx   { return self.Body[len(self.Body)-1].Idx1() }
func (self ThrowStatement) Idx1() Idx    { return (*self.Argument).Expr.Idx1() }
func (self TryStatement) Idx1() Idx {
	if self.Finally != nil {
		return self.Finally.Idx1()
	}
	if self.Catch != nil {
		return self.Catch.Idx1()
	}
	return self.Body.Idx1()
}
func (self VariableStatement) Idx1() Idx   { return self.List[len(self.List)-1].Idx1() }
func (self WhileStatement) Idx1() Idx      { return (*self.Body).Idx1() }
func (self WithStatement) Idx1() Idx       { return (*self.Body).Idx1() }
func (self LexicalDeclaration) Idx1() Idx  { return self.List[len(self.List)-1].Idx1() }
func (self FunctionDeclaration) Idx1() Idx { return self.Function.Idx1() }
func (self ClassDeclaration) Idx1() Idx    { return self.Class.Idx1() }
func (self VariableDeclarator) Idx1() Idx {
	if self.Initializer != nil {
		return (*self.Initializer).Expr.Idx1()
	}
	return self.Target.Idx1()
}

func (self ForLoopInitializerExpression) Idx1() Idx { return (*self.Expression).Expr.Idx1() }
func (self ForLoopInitializerVarDeclList) Idx1() Idx {
	return self.List[len(self.List)-1].Idx1()
}
func (self ForLoopInitializerLexicalDecl) Idx1() Idx { return self.LexicalDeclaration.Idx1() }

func (self PropertyShort) Idx1() Idx {
	if self.Initializer != nil {
		return (*self.Initializer).Expr.Idx1()
	}
	return self.Name.Idx1()
}

func (self PropertyKeyed) Idx1() Idx { return (*self.Value).Expr.Idx1() }

func (self ExpressionBody) Idx1() Idx { return self.Expression.Expr.Idx1() }

func (self FieldDefinition) Idx1() Idx {
	if self.Initializer != nil {
		return (*self.Initializer).Expr.Idx1()
	}
	return (*self.Key).Expr.Idx1()
}

func (self MethodDefinition) Idx1() Idx {
	return self.Body.Idx1()
}

func (self ClassStaticBlock) Idx1() Idx {
	return self.Block.Idx1()
}

func (self YieldExpression) Idx1() Idx {
	if self.Argument != nil {
		return (*self.Argument).Expr.Idx1()
	}
	return self.Yield + 5
}

func (self ForDeclaration) Idx1() Idx    { return self.Target.Idx1() }
func (self ForIntoVar) Idx1() Idx        { return self.Binding.Idx1() }
func (self ForIntoExpression) Idx1() Idx { return (*self.Expression).Expr.Idx1() }
