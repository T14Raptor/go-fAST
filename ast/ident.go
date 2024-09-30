package ast

type (
	ScopeContext int

	Id struct {
		Name         string
		ScopeContext ScopeContext
	}

	Identifier struct {
		Idx          Idx
		Name         string
		ScopeContext ScopeContext
	}
)

func (n *Identifier) ToId() Id {
	return Id{Name: n.Name, ScopeContext: n.ScopeContext}
}

func (*Identifier) _expr() {}
