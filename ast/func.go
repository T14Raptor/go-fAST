package ast

type (
	FunctionLiteral struct {
		Function      Idx
		Name          *Identifier
		ParameterList ParameterList `optional:"true"`
		Body          *BlockStatement

		Async, Generator bool
	}

	ParameterList struct {
		Opening Idx
		List    VariableDeclarators
		Rest    Expr `optional:"true"`
		Closing Idx
	}
)

func (*FunctionLiteral) _expr() {}
