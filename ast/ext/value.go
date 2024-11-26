package ext

type (
	Type interface {
		_type()
	}

	UndefinedType struct{}
	NullType      struct{}
	BoolType      struct{}
	StringType    struct{}
	SymbolType    struct{}
	NumberType    struct{}
	ObjectType    struct{}
)

type Value[T any] struct {
	val     T
	unknown bool
}

func Known[T any](val T) Value[T] {
	return Value[T]{val: val}
}

func Unknown[T any]() Value[T] {
	return Value[T]{unknown: true}
}

func (v Value[T]) Val() T {
	return v.val
}

func (v Value[T]) Known() bool {
	return !v.unknown
}

func (v Value[T]) Unknown() bool {
	return v.unknown
}

type TypeValue struct {
	Value[Type]
}

func (v TypeValue) CastToNumberOnAdd() bool {
	switch v.val.(type) {
	case BoolType, NullType, NumberType, UndefinedType:
		return true
	}
	return false
}

type BoolValue struct {
	Value[bool]
}

func (v BoolValue) And(other BoolValue) BoolValue {
	if v.unknown {
		if other.unknown || other.val {
			return BoolValue{Unknown[bool]()}
		}
		return BoolValue{Known(false)}
	}
	if !v.val {
		return BoolValue{Known(false)}
	}
	return other
}

func (v BoolValue) Or(other BoolValue) BoolValue {
	if v.unknown {
		if other.unknown || !other.val {
			return BoolValue{Unknown[bool]()}
		}
		return BoolValue{Known(true)}
	}
	if v.val {
		return BoolValue{Known(true)}
	}
	return other
}

func (v BoolValue) Not() BoolValue {
	if v.unknown {
		return v
	}
	return BoolValue{Known(!v.val)}
}

func (UndefinedType) _type() {}
func (NullType) _type()      {}
func (BoolType) _type()      {}
func (StringType) _type()    {}
func (SymbolType) _type()    {}
func (NumberType) _type()    {}
func (ObjectType) _type()    {}
