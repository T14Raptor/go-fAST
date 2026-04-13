package parser

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

// Precedence represents operator binding power for Pratt parsing.
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
func kindToPrecedence(kind token.Token) Precedence {
	return tokenPrecedence[kind]
}

func isBinaryOperator(kind token.Token) bool {
	switch kind {
	case token.LogicalAnd, token.Equal, token.NotEqual, token.StrictEqual, token.StrictNotEqual,
		token.Less, token.LessOrEqual, token.Greater, token.GreaterOrEqual, token.Plus, token.Minus,
		token.Multiply, token.Slash, token.Remainder, token.Exponent, token.ShiftLeft, token.ShiftRight,
		token.UnsignedShiftRight, token.Or, token.ExclusiveOr, token.And, token.In, token.InstanceOf:
		return true
	}
	return false
}

func toBinaryOperator(kind token.Token) ast.BinaryOperator {
	switch kind {
	case token.Equal:
		return ast.BinaryEquality
	case token.NotEqual:
		return ast.BinaryInequality
	case token.StrictEqual:
		return ast.BinaryStrictEquality
	case token.StrictNotEqual:
		return ast.BinaryStrictInequality
	case token.Less:
		return ast.BinaryLessThan
	case token.LessOrEqual:
		return ast.BinaryLessEqualThan
	case token.Greater:
		return ast.BinaryGreaterThan
	case token.GreaterOrEqual:
		return ast.BinaryGreaterEqualThan
	case token.Plus:
		return ast.BinaryAddition
	case token.Minus:
		return ast.BinarySubtraction
	case token.Multiply:
		return ast.BinaryMultiplication
	case token.Slash:
		return ast.BinaryDivision
	case token.Remainder:
		return ast.BinaryRemainder
	case token.Exponent:
		return ast.BinaryExponential
	case token.ShiftLeft:
		return ast.BinaryShiftLeft
	case token.ShiftRight:
		return ast.BinaryShiftRight
	case token.UnsignedShiftRight:
		return ast.BinaryUnsignedShiftRight
	case token.Or:
		return ast.BinaryBitwiseOR
	case token.ExclusiveOr:
		return ast.BinaryBitwiseXOR
	case token.And:
		return ast.BinaryBitwiseAnd
	case token.In:
		return ast.BinaryIn
	case token.InstanceOf:
		return ast.BinaryInstanceof
	}
	panic("invalid token")
}

func isUnaryOperator(kind token.Token) bool {
	switch kind {
	case token.Plus, token.Minus, token.Not, token.BitwiseNot, token.Delete, token.Void, token.Typeof:
		return true
	}
	return false
}

func toUnaryOperator(kind token.Token) ast.UnaryOperator {
	switch kind {
	case token.Plus:
		return ast.UnaryPlus
	case token.Minus:
		return ast.UnaryNegation
	case token.Not:
		return ast.UnaryLogicalNot
	case token.BitwiseNot:
		return ast.UnaryBitwiseNot
	case token.Delete:
		return ast.UnaryDelete
	case token.Void:
		return ast.UnaryVoid
	case token.Typeof:
		return ast.UnaryTypeof
	}
	panic("invalid token")
}

func isUpdateOperator(kind token.Token) bool {
	switch kind {
	case token.Increment, token.Decrement:
		return true
	}
	return false
}

func toUpdateOperator(kind token.Token) ast.UpdateOperator {
	switch kind {
	case token.Increment:
		return ast.UpdateIncrement
	case token.Decrement:
		return ast.UpdateDecrement
	}
	panic("invalid token")
}

// isLogicalOperator returns true if the token is a logical operator (&&, ||, ??).
func isLogicalOperator(kind token.Token) bool {
	switch kind {
	case token.LogicalAnd, token.LogicalOr, token.Coalesce:
		return true
	}
	return false
}

func toLogicalOperator(kind token.Token) ast.LogicalOperator {
	switch kind {
	case token.LogicalAnd:
		return ast.LogicalAnd
	case token.LogicalOr:
		return ast.LogicalOr
	case token.Coalesce:
		return ast.LogicalCoalesce
	}
	panic("invalid token")
}

func isAssignOperator(kind token.Token) bool {
	switch kind {
	case token.Assign, token.AddAssign, token.SubtractAssign, token.MultiplyAssign, token.ExponentAssign,
		token.QuotientAssign, token.RemainderAssign, token.AndAssign, token.OrAssign, token.ExclusiveOrAssign,
		token.ShiftLeftAssign, token.ShiftRightAssign, token.UnsignedShiftRightAssign, token.LogicalAndAssign,
		token.LogicalOrAssign, token.CoalesceAssign:
		return true
	}
	return false
}

func toAssignOperator(kind token.Token) ast.AssignmentOperator {
	switch kind {
	case token.Assign:
		return ast.AssignmentAssign
	case token.AddAssign:
		return ast.AssignmentAddition
	case token.SubtractAssign:
		return ast.AssignmentSubtraction
	case token.MultiplyAssign:
		return ast.AssignmentMultiplication
	case token.ExponentAssign:
		return ast.AssignmentExponential
	case token.QuotientAssign:
		return ast.AssignmentDivision
	case token.RemainderAssign:
		return ast.AssignmentRemainder
	case token.AndAssign:
		return ast.AssignmentBitwiseAnd
	case token.OrAssign:
		return ast.AssignmentBitwiseOR
	case token.ExclusiveOrAssign:
		return ast.AssignmentBitwiseXOR
	case token.ShiftLeftAssign:
		return ast.AssignmentShiftLeft
	case token.ShiftRightAssign:
		return ast.AssignmentShiftRight
	case token.UnsignedShiftRightAssign:
		return ast.AssignmentUnsignedShiftRight
	case token.LogicalAndAssign:
		return ast.AssignmentLogicalAnd
	case token.LogicalOrAssign:
		return ast.AssignmentLogicalOr
	case token.CoalesceAssign:
		return ast.AssignmentLogicalNullish
	}
	panic("invalid token")
}
