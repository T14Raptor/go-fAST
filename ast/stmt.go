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
		Idx   Idx
		Label *Identifier `optional:"true"`
	}

	ContinueStatement struct {
		Idx   Idx
		Label *Identifier `optional:"true"`
	}

	CaseStatements []CaseStatement

	CaseStatement struct {
		Case       Idx
		Test       *Expression `optional:"true"`
		Consequent Statements
	}

	CatchStatement struct {
		Catch     Idx
		Parameter *BindingTarget `optional:"true"`
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
		Alternate  *Statement `optional:"true"`
	}

	LabelledStatement struct {
		Label     *Identifier
		Colon     Idx
		Statement *Statement
	}

	ReturnStatement struct {
		Return   Idx
		Argument *Expression `optional:"true"`
	}

	SwitchStatement struct {
		Body CaseStatements

		Discriminant *Expression

		Default int

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

		For Idx
	}

	//union:Expression,VariableDeclaration
	ForInto struct {
		ptr  unsafe.Pointer
		kind ForIntoKind
	}
)
