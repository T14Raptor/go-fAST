package ast

import "math/big"

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

	// BigIntLiteral is a BigInt numeric literal such as 42n, 0xffn, 0b11n, 0o7n.
	// Value holds the parsed arbitrary-precision integer; Raw (if set) is the
	// original source text including the trailing 'n'.
	BigIntLiteral struct {
		Value *big.Int

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
func (*BigIntLiteral) _expr()  {}
func (*RegExpLiteral) _expr()  {}
func (*StringLiteral) _expr()  {}
