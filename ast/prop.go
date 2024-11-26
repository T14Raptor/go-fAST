package ast

type PropertyKind string

const (
	PropertyKindValue  PropertyKind = "value"
	PropertyKindGet    PropertyKind = "get"
	PropertyKindSet    PropertyKind = "set"
	PropertyKindMethod PropertyKind = "method"
)

type (
	Properties []Property

	Property struct {
		Prop Prop
	}

	Prop interface {
		Expr
		_property()
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

func (*PropertyShort) _property() {}
func (*PropertyKeyed) _property() {}
func (*SpreadElement) _property() {}

func (*ComputedProperty) _memberProperty() {}
