package ast

import "unsafe"

type PropertyKind string

const (
	PropertyKindValue  PropertyKind = "value"
	PropertyKindGet    PropertyKind = "get"
	PropertyKindSet    PropertyKind = "set"
	PropertyKindMethod PropertyKind = "method"
)

type (
	Properties []Property

	//union:PropertyKeyed,PropertyShort,SpreadElement
	Property struct {
		ptr  unsafe.Pointer
		kind PropKind
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

