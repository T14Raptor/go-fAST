package ast

type (
	Statements []Statement

	Statement struct {
		Stmt Stmt `optional:"true"`
	}

	// All statement nodes implement the Stmt interface.
	Stmt interface {
		Node
		VisitableNode
		_stmt()
	}

	BadStatement struct {
		From Idx
		To   Idx
	}

	BlockStatement struct {
		List Statements

		ScopeContext ScopeContext

		LeftBrace  Idx
		RightBrace Idx
	}

	BreakStatement struct {
		Label *Identifier `optional:"true"`

		Idx Idx
	}

	ContinueStatement struct {
		Label *Identifier `optional:"true"`

		Idx Idx
	}

	CaseStatements []CaseStatement

	CaseStatement struct {
		Test       *Expression `optional:"true"`
		Consequent Statements

		Case Idx
	}

	CatchStatement struct {
		Parameter *BindingTarget `optional:"true"`
		Body      *BlockStatement

		Catch Idx
	}

	DebuggerStatement struct {
		Debugger Idx
	}

	DoWhileStatement struct {
		Test *Expression
		Body *Statement

		Do Idx
	}

	EmptyStatement struct {
		Semicolon Idx
	}

	ExpressionStatement struct {
		Expression *Expression
		Comment    string
	}

	IfStatement struct {
		Test       *Expression
		Consequent *Statement
		Alternate  *Statement `optional:"true"`

		If Idx
	}

	LabelledStatement struct {
		Label     *Identifier
		Statement *Statement

		Colon Idx
	}

	ReturnStatement struct {
		Argument *Expression `optional:"true"`

		Return Idx
	}

	SwitchStatement struct {
		Discriminant *Expression
		Default      int
		Body         CaseStatements

		Switch Idx
	}

	ThrowStatement struct {
		Argument *Expression

		Throw Idx
	}

	TryStatement struct {
		Body    *BlockStatement
		Catch   *CatchStatement `optional:"true"`
		Finally *BlockStatement `optional:"true"`

		Try Idx
	}

	WhileStatement struct {
		Test *Expression
		Body *Statement

		While Idx
	}

	WithStatement struct {
		Object *Expression
		Body   *Statement

		With Idx
	}

	ForStatement struct {
		Initializer *ForLoopInitializer `optional:"true"`
		Update      *Expression
		Test        *Expression
		Body        *Statement

		For Idx
	}

	ForLoopInitializer struct {
		Initializer ForLoopInit
	}

	ForLoopInit interface {
		VisitableNode
		_forLoopInitializer()
	}

	ForInStatement struct {
		Into   *ForInto
		Source *Expression
		Body   *Statement

		For Idx
	}

	ForOfStatement struct {
		Into   *ForInto
		Source *Expression
		Body   *Statement

		For Idx
	}

	ForInto struct {
		Into
	}

	Into interface {
		VisitableNode
		_forInto()
	}
)

func (*VariableDeclaration) _forLoopInitializer() {}
func (*Expression) _forLoopInitializer()          {}

func (*VariableDeclaration) _forInto() {}
func (*Expression) _forInto()          {}

func (*BadStatement) _stmt()        {}
func (*BlockStatement) _stmt()      {}
func (*BreakStatement) _stmt()      {}
func (*CaseStatement) _stmt()       {}
func (*ContinueStatement) _stmt()   {}
func (*CatchStatement) _stmt()      {}
func (*DebuggerStatement) _stmt()   {}
func (*DoWhileStatement) _stmt()    {}
func (*EmptyStatement) _stmt()      {}
func (*ExpressionStatement) _stmt() {}
func (*ForInStatement) _stmt()      {}
func (*ForOfStatement) _stmt()      {}
func (*ForStatement) _stmt()        {}
func (*IfStatement) _stmt()         {}
func (*LabelledStatement) _stmt()   {}
func (*ReturnStatement) _stmt()     {}
func (*SwitchStatement) _stmt()     {}
func (*ThrowStatement) _stmt()      {}
func (*TryStatement) _stmt()        {}
func (*WhileStatement) _stmt()      {}
func (*WithStatement) _stmt()       {}
