package ast

import (
	"unsafe"
)

type (
	Statements []Statement

	//union:BadStatement,BlockStatement,BreakStatement,CaseStatement,CatchStatement,ClassDeclaration,ContinueStatement,DebuggerStatement,DoWhileStatement,EmptyStatement,ExpressionStatement,ForStatement,ForInStatement,ForOfStatement,FunctionDeclaration,IfStatement,LabelledStatement,ReturnStatement,SwitchStatement,ThrowStatement,TryStatement,VariableDeclaration,WhileStatement,WithStatement
	Statement struct {
		ptr unsafe.Pointer

		kind StmtKind
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
		Body         CaseStatements
		Default      int

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

	//union:Expression,VariableDeclaration
	ForLoopInitializer struct {
		ptr  unsafe.Pointer
		kind ForInitKind
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

		For   Idx
		Await bool
	}

	//union:Expression,VariableDeclaration
	ForInto struct {
		ptr  unsafe.Pointer
		kind ForIntoKind
	}
)
