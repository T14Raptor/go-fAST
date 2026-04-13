package token

import (
	"strconv"
)

// Token is the set of lexical tokens in JavaScript (ECMA5).
type Token byte

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

// keyword ...
type keyword struct {
	token         Token
	futureKeyword bool
	strict        bool
}

func MatchKeyword(literal string) Token {
	if k, exists := keywordTable[literal]; exists {
		return k.token
	}
	return Identifier
}

// ID ...
func ID(token Token) bool {
	return token >= Identifier
}

// UnreservedWord ...
func UnreservedWord(token Token) bool {
	return token > EscapedReservedWord
}
