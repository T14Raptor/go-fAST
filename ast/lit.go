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
		// Note: NaN should not be stored here, use an identifier instead.
		Value float64

		Raw *string

		Idx Idx
	}

	RegExpLiteral struct {
		Literal string
		Pattern string
		Flags   string

		Idx Idx
	}

	StringLiteral struct {
		Value string

		Raw *string

		Idx Idx
	}
)

func (*BooleanLiteral) _expr() {}
func (*NullLiteral) _expr()    {}
func (*NumberLiteral) _expr()  {}
func (*RegExpLiteral) _expr()  {}
func (*StringLiteral) _expr()  {}
