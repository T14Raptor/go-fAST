package ast

type (
	Statements []Statement

	Statement struct {
		Stmt
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
		LeftBrace  Idx
		List       Statements
		RightBrace Idx
	}

	BreakStatement struct {
		Idx   Idx
		Label *Identifier
	}

	ContinueStatement struct {
		Idx   Idx
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

	ForLoopInitializer struct {
		Initializer ForLoopInit
	}

	ForLoopInit interface {
		VisitableNode
		_forLoopInitializer()
	}

	ForInto interface {
		VisitableNode
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

func (*VariableDeclaration) _forLoopInitializer() {}
func (*Expression) _forLoopInitializer()          {}

func (*ForIntoVar) _forInto()        {}
func (*ForDeclaration) _forInto()    {}
func (*ForIntoExpression) _forInto() {}

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
