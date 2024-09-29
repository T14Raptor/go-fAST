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

func (i *Identifier) ToId() Id {
	return Id{Name: i.Name, ScopeContext: i.ScopeContext}
}

func (*Identifier) _expr() {}
