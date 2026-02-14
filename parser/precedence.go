package parser

import "github.com/t14raptor/go-fast/token"

// Precedence represents operator binding power for Pratt parsing.
//
// Values use a binding-power encoding where even values represent
// left-associative operators and odd values represent right-associative
// operators. The Pratt loop uses a single comparison (lbp <= minBP)
// and the recursive call passes lbp ^ 1 as the new minimum. The XOR
// flips even↔odd:
//
//   - Left-assoc  (even lbp): recursive min = lbp+1 (odd)  → same-level breaks
//   - Right-assoc (odd  lbp): recursive min = lbp-1 (even) → same-level continues
//
// This eliminates the IsRightAssociative branch from the hot loop entirely.
//
// See: https://matklad.github.io/2020/04/13/simple-but-powerful-pratt-parsing.html
type Precedence uint8

const (
	PrecedenceLowest            Precedence = 0
	PrecedenceComma             Precedence = 2  // ,             (left-assoc)
	PrecedenceSpread            Precedence = 4  // ...           (left-assoc)
	PrecedenceYield             Precedence = 6  // yield         (left-assoc)
	PrecedenceAssign            Precedence = 9  // = += -= etc   (right-assoc)
	PrecedenceConditional       Precedence = 11 // ?:            (right-assoc)
	PrecedenceNullishCoalescing Precedence = 12 // ??            (left-assoc)
	PrecedenceLogicalOr         Precedence = 14 // ||            (left-assoc)
	PrecedenceLogicalAnd        Precedence = 16 // &&            (left-assoc)
	PrecedenceBitwiseOr         Precedence = 18 // |             (left-assoc)
	PrecedenceBitwiseXor        Precedence = 20 // ^             (left-assoc)
	PrecedenceBitwiseAnd        Precedence = 22 // &             (left-assoc)
	PrecedenceEquals            Precedence = 24 // == != === !== (left-assoc)
	PrecedenceCompare           Precedence = 26 // < > <= >= instanceof in (left-assoc)
	PrecedenceShift             Precedence = 28 // << >> >>>     (left-assoc)
	PrecedenceAdd               Precedence = 30 // + -           (left-assoc)
	PrecedenceMultiply          Precedence = 32 // * / %         (left-assoc)
	PrecedenceExponentiation    Precedence = 35 // **            (right-assoc)
	PrecedencePrefix            Precedence = 36 // ! ~ + - typeof void delete
	PrecedencePostfix           Precedence = 38 // ++ --
	PrecedenceNew               Precedence = 40 // new
	PrecedenceCall              Precedence = 42 // ()
	PrecedenceMember            Precedence = 44 // . []
)

// tokenPrecedence maps each token kind to its left binding power.
// Zero means the token is not a binary/logical operator.
// The table is 256 bytes (one cache line batch) and stays hot in L1.
var tokenPrecedence [256]Precedence

func init() {
	tokenPrecedence[token.Coalesce] = PrecedenceNullishCoalescing
	tokenPrecedence[token.LogicalOr] = PrecedenceLogicalOr
	tokenPrecedence[token.LogicalAnd] = PrecedenceLogicalAnd
	tokenPrecedence[token.Or] = PrecedenceBitwiseOr
	tokenPrecedence[token.ExclusiveOr] = PrecedenceBitwiseXor
	tokenPrecedence[token.And] = PrecedenceBitwiseAnd
	tokenPrecedence[token.Equal] = PrecedenceEquals
	tokenPrecedence[token.StrictEqual] = PrecedenceEquals
	tokenPrecedence[token.NotEqual] = PrecedenceEquals
	tokenPrecedence[token.StrictNotEqual] = PrecedenceEquals
	tokenPrecedence[token.Less] = PrecedenceCompare
	tokenPrecedence[token.Greater] = PrecedenceCompare
	tokenPrecedence[token.LessOrEqual] = PrecedenceCompare
	tokenPrecedence[token.GreaterOrEqual] = PrecedenceCompare
	tokenPrecedence[token.InstanceOf] = PrecedenceCompare
	tokenPrecedence[token.In] = PrecedenceCompare
	tokenPrecedence[token.ShiftLeft] = PrecedenceShift
	tokenPrecedence[token.ShiftRight] = PrecedenceShift
	tokenPrecedence[token.UnsignedShiftRight] = PrecedenceShift
	tokenPrecedence[token.Plus] = PrecedenceAdd
	tokenPrecedence[token.Minus] = PrecedenceAdd
	tokenPrecedence[token.Multiply] = PrecedenceMultiply
	tokenPrecedence[token.Slash] = PrecedenceMultiply
	tokenPrecedence[token.Remainder] = PrecedenceMultiply
	tokenPrecedence[token.Exponent] = PrecedenceExponentiation
}

// kindToPrecedence returns the left binding power for a token kind.
// Returns 0 if the token is not a binary/logical operator.
//
// This is a single indexed load with no branches, replacing the
// previous multi-arm switch statement.
func kindToPrecedence(kind token.Token) Precedence {
	return tokenPrecedence[kind]
}

// isLogicalOperator returns true if the token is a logical operator (&&, ||, ??).
func isLogicalOperator(kind token.Token) bool {
	return kind == token.LogicalAnd || kind == token.LogicalOr || kind == token.Coalesce
}
