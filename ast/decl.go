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
