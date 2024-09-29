package ast

type (
	FunctionLiteral struct {
		Function      Idx
		Name          *Identifier
		ParameterList ParameterList
		Body          *BlockStatement

		Async, Generator bool
	}

	ParameterList struct {
		Opening Idx
		List    VariableDeclarators
		Rest    Expr
		Closing Idx
	}
)

func (*FunctionLiteral) _expr() {}
