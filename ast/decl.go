package ast

type (
	FunctionDeclaration struct {
		Function *FunctionLiteral
	}

	ClassDeclaration struct {
		Class *ClassLiteral
	}

	VariableDeclaration struct {
		List    VariableDeclarators
		Comment string

		Idx  Idx
		Kind VarKind
	}

	VariableDeclarators []VariableDeclarator

	VariableDeclarator struct {
		Target      *Pattern
		Initializer *Expression `optional:"true"`
	}
)

type VarKind uint8

const (
	VarKindVar VarKind = iota
	VarKindLet
	VarKindConst
)

func (k VarKind) String() string {
	switch k {
	case VarKindVar:
		return "var"
	case VarKindLet:
		return "let"
	case VarKindConst:
		return "const"
	}
	return ""
}

func (n *FunctionDeclaration) Idx0() Idx { return n.Function.Idx0() }
func (n *FunctionDeclaration) Idx1() Idx { return n.Function.Idx1() }

func (n *ClassDeclaration) Idx0() Idx { return n.Class.Idx0() }
func (n *ClassDeclaration) Idx1() Idx { return n.Class.Idx1() }

func (n *VariableDeclaration) Idx0() Idx { return n.Idx }
func (n *VariableDeclaration) Idx1() Idx { return n.List[len(n.List)-1].Idx1() }

func (b *VariableDeclarator) Idx0() Idx { return b.Target.Idx0() }
func (b *VariableDeclarator) Idx1() Idx {
	if b.Initializer != nil {
		return b.Initializer.Idx1()
	}
	return b.Target.Idx1()
}
