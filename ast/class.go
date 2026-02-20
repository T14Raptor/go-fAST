package ast

import "unsafe"

type (
	ClassLiteral struct {
		Name       *Identifier `optional:"true"`
		SuperClass *Expression `optional:"true"`
		Body       ClassElements

		Class      Idx
		RightBrace Idx
	}

	ClassElements []ClassElement

	//union:ClassStaticBlock,FieldDefinition,MethodDefinition
	ClassElement struct {
		ptr  unsafe.Pointer
		kind ClassElemKind
	}

	FieldDefinition struct {
		Key         *Expression
		Initializer *Expression `optional:"true"`

		Idx Idx

		Computed bool
		Static   bool
	}

	MethodDefinition struct {
		Key  *Expression
		Kind PropertyKind // "method", "get" or "set"
		Body *FunctionLiteral

		Idx      Idx
		Computed bool
		Static   bool
	}

	ClassStaticBlock struct {
		Block *BlockStatement

		Static Idx
	}
)
