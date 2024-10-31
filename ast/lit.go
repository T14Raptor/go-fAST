package ast

type (
	BooleanLiteral struct {
		Idx   Idx
		Value bool
	}

	NullLiteral struct {
		Idx Idx
	}

	NumberLiteral struct {
		Idx Idx
		// Note: NaN should not be stored here, use an identifier instead.
		Value float64

		Raw *string
	}

	RegExpLiteral struct {
		Idx     Idx
		Literal string
		Pattern string
		Flags   string
	}

	StringLiteral struct {
		Idx   Idx
		Value string

		Raw *string
	}
)

func (*BooleanLiteral) _expr() {}
func (*NullLiteral) _expr()    {}
func (*NumberLiteral) _expr()  {}
func (*RegExpLiteral) _expr()  {}
func (*StringLiteral) _expr()  {}
