package ast

type (
	Id struct {
		Name         string
		ScopeContext ScopeContext
	}

	Identifier struct {
		Name         string
		ScopeContext ScopeContext

		Idx Idx
	}
)

func (n *Identifier) ToId() Id {
	return Id{Name: n.Name, ScopeContext: n.ScopeContext}
}
func (i *Id) String() string { return i.Name }
