package ast

import "github.com/t14raptor/go-fast/token"

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

		Idx   Idx
		Token token.Token
	}

	VariableDeclarators []VariableDeclarator

	VariableDeclarator struct {
		Target      *BindingTarget
		Initializer *Expression `optional:"true"`
	}
)
