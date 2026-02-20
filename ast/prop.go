package ast

import "unsafe"

type PropertyKind uint8

const (
	PropertyKindValue  PropertyKind = iota + 1
	PropertyKindGet
	PropertyKindSet
	PropertyKindMethod
)

type (
	Properties []Property

	//union:PropertyKeyed,PropertyShort,SpreadElement
	Property struct {
		kind PropKind

		ptr unsafe.Pointer
	}

	PropertyShort struct {
		Name        *Identifier
		Initializer *Expression
	}

	PropertyKeyed struct {
		Key      *Expression
		Kind     PropertyKind
		Value    *Expression
		Computed bool
	}

	ComputedProperty struct {
		Expr *Expression
	}
)
