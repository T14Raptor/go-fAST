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
		Rest *Pattern `optional:"true"`

		Opening Idx
		Closing Idx
	}
)

func (f *FunctionLiteral) Idx0() Idx { return f.Function }
func (f *FunctionLiteral) Idx1() Idx { return f.Body.Idx1() }

func (n *ParameterList) Idx0() Idx { return n.Opening }
func (n *ParameterList) Idx1() Idx { return n.Closing + 1 }
