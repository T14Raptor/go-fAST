package ast

//go:generate go run ast/gen_clone.go
//go:generate go run ast/gen_visit.go
//go:generate go run ast/gen_union.go

// Idx is a compact encoding of a source position within JS code.
type Idx uint32

type Node interface {
	// Idx0 returns the index of the first character belonging to the node.
	Idx0() Idx
	// Idx1 returns the index of the first character immediately after the node.
	Idx1() Idx
}

type VisitableNode interface {
	VisitWith(v Visitor)
	VisitChildrenWith(v Visitor)
}

type Program struct {
	Body Statements
}

func (n *Program) Idx0() Idx { return n.Body[0].Idx0() }
func (n *Program) Idx1() Idx { return n.Body[len(n.Body)-1].Idx1() }
