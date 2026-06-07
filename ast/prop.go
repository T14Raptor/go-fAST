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
