package ast

import (
	"math/big"

	"github.com/nukilabs/ftoa"
)

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

func (n *BooleanLiteral) Idx0() Idx { return n.Idx }
func (n *BooleanLiteral) Idx1() Idx { return Idx(int(n.Idx) + 4) }

func (n *NullLiteral) Idx0() Idx { return n.Idx }
func (n *NullLiteral) Idx1() Idx { return Idx(int(n.Idx) + 4) } // "null"

func (n *NumberLiteral) Idx0() Idx { return n.Idx }
func (n *NumberLiteral) Idx1() Idx {
	if n.Raw != nil {
		return Idx(int(n.Idx) + len(*n.Raw))
	}
	raw := ftoa.FormatFloat(n.Value, 'g', -1, 64)
	return Idx(int(n.Idx) + len(raw))
}

func (n *BigIntLiteral) Idx0() Idx { return n.Idx }
func (n *BigIntLiteral) Idx1() Idx {
	if n.Raw != nil {
		return Idx(int(n.Idx) + len(*n.Raw))
	}
	// Value string + trailing 'n'.
	if n.Value != nil {
		return Idx(int(n.Idx) + len(n.Value.String()) + 1)
	}
	return n.Idx
}

func (n *RegExpLiteral) Idx0() Idx { return n.Idx }
func (n *RegExpLiteral) Idx1() Idx { return Idx(int(n.Idx) + len(n.Literal)) }

func (n *StringLiteral) Idx0() Idx { return n.Idx }
func (n *StringLiteral) Idx1() Idx {
	if n.Raw != nil {
		return Idx(int(n.Idx) + len(*n.Raw))
	}
	return Idx(int(n.Idx) + len(n.Value) + 2) // +2 for the quotes
}
