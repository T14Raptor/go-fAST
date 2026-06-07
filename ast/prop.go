package ast

import "unsafe"

type (
	Properties []Property

	//union:PropertyKeyValue,PropertyMethod,PropertyGetter,PropertySetter,PropertyShort,SpreadElement
	Property struct {
		kind PropKind

		ptr unsafe.Pointer
	}

	// PropertyName is an object/class member key. Computed keys (`[expr]`) are the
	// ComputedProperty variant, so there is no separate "computed" flag, and a
	// computed literal key (`["a"]`) stays distinct from a plain one (`"a"`).
	//
	//union:StringLiteral,NumberLiteral,BigIntLiteral,PrivateIdentifier,ComputedProperty
	PropertyName struct {
		kind PropNameKind

		ptr unsafe.Pointer
	}

	// PropertyKeyValue is `{ key: value }` / `{ [key]: value }`.
	PropertyKeyValue struct {
		Key   *PropertyName
		Value *Expression
	}

	// PropertyMethod is a method shorthand: `{ key() {} }`, `{ *key() {} }`,
	// `{ async key() {} }`, `{ async *key() {} }`. async/generator flags live on Body.
	PropertyMethod struct {
		Key  *PropertyName
		Body *FunctionLiteral
	}

	// PropertyGetter is `{ get key() {} }`.
	PropertyGetter struct {
		Key  *PropertyName
		Body *FunctionLiteral
	}

	// PropertySetter is `{ set key(v) {} }`.
	PropertySetter struct {
		Key  *PropertyName
		Body *FunctionLiteral
	}

	// PropertyShort is shorthand `{ a }` (object literal) or `{ a = d }` (only
	// valid once the object is reinterpreted as a destructuring pattern).
	PropertyShort struct {
		Name        *Identifier
		Initializer *Expression `optional:"true"`
	}

	ComputedProperty struct {
		Expr *Expression

		LeftBracket  Idx
		RightBracket Idx
	}
)

func (n *ComputedProperty) Idx0() Idx {
	if n.LeftBracket != 0 {
		return n.LeftBracket
	}
	return n.Expr.Idx0()
}
func (n *ComputedProperty) Idx1() Idx {
	if n.RightBracket != 0 {
		return n.RightBracket + 1
	}
	return n.Expr.Idx1()
}

func (n *PropertyKeyValue) Idx0() Idx { return n.Key.Idx0() }
func (n *PropertyKeyValue) Idx1() Idx { return n.Value.Idx1() }

func (n *PropertyMethod) Idx0() Idx { return n.Key.Idx0() }
func (n *PropertyMethod) Idx1() Idx { return n.Body.Idx1() }

func (n *PropertyGetter) Idx0() Idx { return n.Key.Idx0() }
func (n *PropertyGetter) Idx1() Idx { return n.Body.Idx1() }

func (n *PropertySetter) Idx0() Idx { return n.Key.Idx0() }
func (n *PropertySetter) Idx1() Idx { return n.Body.Idx1() }

func (n *PropertyShort) Idx0() Idx { return n.Name.Idx }
func (n *PropertyShort) Idx1() Idx {
	if n.Initializer != nil {
		return n.Initializer.Idx1()
	}
	return n.Name.Idx1()
}
