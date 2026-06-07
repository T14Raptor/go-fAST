package ast

import "unsafe"

// MethodKind distinguishes the kinds of class/object method definitions. The
// zero value is intentionally unused so it can mark "not a method" during
// parsing.
type MethodKind uint8

const (
	MethodKindMethod MethodKind = iota + 1
	MethodKindGet
	MethodKindSet
)

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
		Key         *PropertyName
		Initializer *Expression `optional:"true"`

		Idx Idx

		Static bool
	}

	MethodDefinition struct {
		Key  *PropertyName
		Kind MethodKind
		Body *FunctionLiteral

		Idx    Idx
		Static bool
	}

	ClassStaticBlock struct {
		Block *BlockStatement

		Static Idx
	}
)
