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

func (v Value[T]) Value() T {
	return v.val
}

func (v Value[T]) Unknown() bool {
	return v.unknown
}

type TypeValue = Value[Type]

func (v TypeValue) CastToNumberOnAdd() bool {
	switch any(v.Value()).(type) {
	case BoolType, NullType, NumberType, UndefinedType:
		return true
	}
	return false
}

type BoolValue = Value[bool]

func (v BoolValue) And(other BoolValue) BoolValue {
	if v.Unknown() {
		if other.Unknown() {
			return BoolValue{unknown: true}
		}
		return other
	}
	if v.Value() {
		return other
	}
	return BoolValue{}
}

func (v BoolValue) Or(other BoolValue) BoolValue {
	if v.Unknown() {
		if other.Unknown() || !other.Value() {
			return BoolValue{unknown: true}
		}
		return other
	}
	if v.Value() {
		return v
	}
	return other
}

func (v BoolValue) Not() BoolValue {
	if v.Unknown() {
		return v
	}
	return BoolValue{val: !v.Value()}
}

func (UndefinedType) _type() {}
func (NullType) _type()      {}
func (BoolType) _type()      {}
func (StringType) _type()    {}
func (SymbolType) _type()    {}
func (NumberType) _type()    {}
func (ObjectType) _type()    {}
