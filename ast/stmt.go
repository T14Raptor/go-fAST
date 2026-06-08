package ast

import (
	"unsafe"
)

type (
	Statements []Statement

	//union:BadStatement,BlockStatement,BreakStatement,CaseStatement,CatchStatement,ClassDeclaration,ContinueStatement,DebuggerStatement,DoWhileStatement,EmptyStatement,ExpressionStatement,ForStatement,ForInStatement,ForOfStatement,FunctionDeclaration,IfStatement,LabelledStatement,ReturnStatement,SwitchStatement,ThrowStatement,TryStatement,VariableDeclaration,WhileStatement,WithStatement
	Statement struct {
		kind StmtKind

		ptr unsafe.Pointer
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
		Parameter *Pattern `optional:"true"`
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
		Update      *Expression         `optional:"true"`
		Test        *Expression         `optional:"true"`
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

	//union:Pattern,VariableDeclaration
	ForInto struct {
		ptr  unsafe.Pointer
		kind ForIntoKind
	}
)

func (n *BadStatement) Idx0() Idx { return n.From }
func (n *BadStatement) Idx1() Idx { return n.To }

func (n *BlockStatement) Idx0() Idx { return n.LeftBrace }
func (n *BlockStatement) Idx1() Idx { return n.RightBrace + 1 }

func (n *BreakStatement) Idx0() Idx { return n.Idx }
func (n *BreakStatement) Idx1() Idx {
	if n.Label != nil {
		return n.Label.Idx1()
	}
	return n.Idx + 5 // "break"
}

func (n *ContinueStatement) Idx0() Idx { return n.Idx }
func (n *ContinueStatement) Idx1() Idx {
	if n.Label != nil {
		return n.Label.Idx1()
	}
	return n.Idx + 8 // "continue"
}

func (n *CaseStatement) Idx0() Idx { return n.Case }
func (n *CaseStatement) Idx1() Idx {
	if len(n.Consequent) > 0 {
		return n.Consequent[len(n.Consequent)-1].Idx1()
	}
	if n.Test != nil {
		return n.Test.Idx1() // `case <test>:` with no body
	}
	return n.Case + 7 // bare `default`
}

func (n *CatchStatement) Idx0() Idx { return n.Catch }
func (n *CatchStatement) Idx1() Idx { return n.Body.Idx1() }

func (n *DebuggerStatement) Idx0() Idx { return n.Debugger }
func (n *DebuggerStatement) Idx1() Idx { return n.Debugger + 8 }

func (n *DoWhileStatement) Idx0() Idx { return n.Do }
func (n *DoWhileStatement) Idx1() Idx { return n.Test.Idx1() }

func (n *EmptyStatement) Idx0() Idx { return n.Semicolon }
func (n *EmptyStatement) Idx1() Idx { return n.Semicolon + 1 }

func (n *ExpressionStatement) Idx0() Idx { return n.Expression.Idx0() }
func (n *ExpressionStatement) Idx1() Idx { return n.Expression.Idx1() }

func (n *IfStatement) Idx0() Idx { return n.If }
func (n *IfStatement) Idx1() Idx {
	if n.Alternate != nil {
		return n.Alternate.Idx1()
	}
	return n.Consequent.Idx1()
}

func (n *LabelledStatement) Idx0() Idx { return n.Label.Idx0() }
func (n *LabelledStatement) Idx1() Idx { return n.Statement.Idx1() }

func (n *ReturnStatement) Idx0() Idx { return n.Return }
func (n *ReturnStatement) Idx1() Idx {
	if n.Argument != nil {
		return n.Argument.Idx1()
	}
	return n.Return + 6 // "return"
}

func (n *SwitchStatement) Idx0() Idx { return n.Switch }
func (n *SwitchStatement) Idx1() Idx {
	if len(n.Body) > 0 {
		return n.Body[len(n.Body)-1].Idx1()
	}
	return n.Discriminant.Idx1() // empty `switch (x) {}`
}

func (n *ThrowStatement) Idx0() Idx { return n.Throw }
func (n *ThrowStatement) Idx1() Idx { return n.Argument.Idx1() }

func (n *TryStatement) Idx0() Idx { return n.Try }
func (n *TryStatement) Idx1() Idx {
	if n.Finally != nil {
		return n.Finally.Idx1()
	}
	if n.Catch != nil {
		return n.Catch.Idx1()
	}
	return n.Body.Idx1()
}

func (n *WhileStatement) Idx0() Idx { return n.While }
func (n *WhileStatement) Idx1() Idx { return n.Body.Idx1() }

func (n *WithStatement) Idx0() Idx { return n.With }
func (n *WithStatement) Idx1() Idx { return n.Body.Idx1() }

func (n *ForStatement) Idx0() Idx { return n.For }
func (n *ForStatement) Idx1() Idx { return n.Body.Idx1() }

func (n *ForInStatement) Idx0() Idx { return n.For }
func (n *ForInStatement) Idx1() Idx { return n.Body.Idx1() }

func (n *ForOfStatement) Idx0() Idx { return n.For }
func (n *ForOfStatement) Idx1() Idx { return n.Body.Idx1() }
