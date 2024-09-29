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
		Idx     Idx
		Literal string
		Value   any
	}

	RegExpLiteral struct {
		Idx     Idx
		Literal string
		Pattern string
		Flags   string
	}

	StringLiteral struct {
		Idx     Idx
		Literal string
		Value   string
	}
)

func (*BooleanLiteral) _expr() {}
func (*NullLiteral) _expr()    {}
func (*NumberLiteral) _expr()  {}
func (*RegExpLiteral) _expr()  {}
func (*StringLiteral) _expr()  {}
