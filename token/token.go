package token

import (
	"strconv"
)

// Token is the set of lexical tokens in JavaScript (ECMA5).
type Token int

// String returns the string corresponding to the token.
func (t Token) String() string {
	if t == 0 {
		return "UNKNOWN"
	}
	if t < Token(len(token2string)) {
		return token2string[t]
	}
	return "token(" + strconv.Itoa(int(t)) + ")"
}

// Precedence ...
func (t Token) Precedence(in bool) int {
	switch t {
	case LogicalOr:
		return 1
	case LogicalAnd:
		return 2
	case Or, OrAssign:
		return 3
	case ExclusiveOr:
		return 4
	case And, AndAssign:
		return 5
	case Equal,
		NotEqual,
		StrictEqual,
		StrictNotEqual:
		return 6
	case Less, Greater, LessOrEqual, GreaterOrEqual, InstanceOf:
		return 7
	case In:
		if in {
			return 7
		}
		return 0
	case ShiftLeft, ShiftRight, UnsignedShiftRight:
		fallthrough
	case ShiftLeftAssign, ShiftRightAssign, UnsignedShiftRightAssign:
		return 8
	case Plus, Minus, AddAssign, SubtractAssign:
		return 9
	case Multiply, Slash, Remainder, MultiplyAssign, QuotientAssign, RemainderAssign:
		return 11
	}
	return 0
}

// keyword ...
type keyword struct {
	token         Token
	futureKeyword bool
	strict        bool
}

// LiteralKeyword returns the keyword token if literal is a keyword, a Keyword token. If the literal is a future keyword
// (const, let, class, super, ...), or 0 if the literal is not a keyword.
func LiteralKeyword(literal string) (Token, bool) {
	if k, exists := keywordTable[literal]; exists {
		if k.futureKeyword {
			return Keyword, k.strict
		}
		return k.token, false
	}
	return 0, false
}

// ID ...
func ID(token Token) bool {
	return token >= Identifier
}

// UnreservedWord ...
func UnreservedWord(token Token) bool {
	return token > EscapedReservedWord
}
