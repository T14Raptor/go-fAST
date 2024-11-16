package ext

type Type int

const (
	UndefinedType Type = iota
	NullType
	BoolType
	StringType
	SymbolType
	NumberType
	ObjectType
)

func (t Type) CastToNumberOnAdd() bool {
	switch t {
	case BoolType, NullType, NumberType, UndefinedType:
		return true
	default:
		return false
	}
}

func And(av, bv bool, aok, bok bool) (value bool, ok bool) {
	if aok && av {
		return bv, bok
	} else if aok && !av {
		return false, true
	} else if bok && !bv {
		return false, true
	}
	return false, false
}

func Or(av, bv bool, aok, bok bool) (value bool, ok bool) {
	if aok && av {
		return true, true
	} else if aok && !av {
		return bv, bok
	} else if bok && bv {
		return true, true
	}
	return false, false
}
