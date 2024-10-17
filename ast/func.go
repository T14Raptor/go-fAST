package ast

type (
	FunctionLiteral struct {
		Function      Idx
		Name          *Identifier
		ParameterList ParameterList
		Body          *BlockStatement

		Async, Generator bool

		ScopeContext ScopeContext
	}

	ParameterList struct {
		Opening Idx
		List    VariableDeclarators
		Rest    Expr `optional:"true"`
		Closing Idx
	}
)

func (*FunctionLiteral) _expr() {}
