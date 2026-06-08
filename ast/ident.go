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

func (i *Identifier) Idx0() Idx { return i.Idx }
func (i *Identifier) Idx1() Idx { return Idx(int(i.Idx) + len(i.Name)) }
