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

func (c *ClassLiteral) Idx0() Idx { return c.Class }
func (c *ClassLiteral) Idx1() Idx { return c.RightBrace + 1 }

func (n *FieldDefinition) Idx0() Idx { return n.Idx }
func (n *FieldDefinition) Idx1() Idx {
	if n.Initializer != nil {
		return n.Initializer.Idx1()
	}
	return n.Key.Idx1()
}

func (n *MethodDefinition) Idx0() Idx { return n.Idx }
func (n *MethodDefinition) Idx1() Idx { return n.Body.Idx1() }

func (n *ClassStaticBlock) Idx0() Idx { return n.Static }
func (n *ClassStaticBlock) Idx1() Idx { return n.Block.Idx1() }
