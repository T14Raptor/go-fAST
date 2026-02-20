package ast

type (
	FunctionLiteral struct {
		Name          *Identifier `optional:"true"`
		ParameterList *ParameterList
		Body          *BlockStatement

		ScopeContext ScopeContext

		Function Idx

		Async, Generator bool
	}

	ParameterList struct {
		List VariableDeclarators
		Rest *Expression `optional:"true"`

		Opening Idx
		Closing Idx
	}
)
