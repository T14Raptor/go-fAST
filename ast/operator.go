package ast

// AssignmentOperator is the operator in an assignment expression.
// See https://tc39.es/ecma262/#sec-assignment-operators.
type AssignmentOperator uint8

const (
	AssignmentAssign             AssignmentOperator = iota // =
	AssignmentAddition                                     // +=
	AssignmentSubtraction                                  // -=
	AssignmentMultiplication                               // *=
	AssignmentDivision                                     // /=
	AssignmentRemainder                                    // %=
	AssignmentExponential                                  // **=
	AssignmentShiftLeft                                    // <<=
	AssignmentShiftRight                                   // >>=
	AssignmentUnsignedShiftRight                           // >>>=
	AssignmentBitwiseOR                                    // |=
	AssignmentBitwiseXOR                                   // ^=
	AssignmentBitwiseAnd                                   // &=
	AssignmentLogicalOr                                    // ||=
	AssignmentLogicalAnd                                   // &&=
	AssignmentLogicalNullish                               // ??=
)

// String returns op as it appears in source.
func (op AssignmentOperator) String() string {
	switch op {
	case AssignmentAssign:
		return "="
	case AssignmentAddition:
		return "+="
	case AssignmentSubtraction:
		return "-="
	case AssignmentMultiplication:
		return "*="
	case AssignmentDivision:
		return "/="
	case AssignmentRemainder:
		return "%="
	case AssignmentExponential:
		return "**="
	case AssignmentShiftLeft:
		return "<<="
	case AssignmentShiftRight:
		return ">>="
	case AssignmentUnsignedShiftRight:
		return ">>>="
	case AssignmentBitwiseOR:
		return "|="
	case AssignmentBitwiseXOR:
		return "^="
	case AssignmentBitwiseAnd:
		return "&="
	case AssignmentLogicalOr:
		return "||="
	case AssignmentLogicalAnd:
		return "&&="
	case AssignmentLogicalNullish:
		return "??="
	}
	return ""
}

// IsAssign reports whether op is plain =.
func (op AssignmentOperator) IsAssign() bool { return op == AssignmentAssign }

// IsArithmetic reports whether op is one of +=, -=, *=, /=, %=, **=.
func (op AssignmentOperator) IsArithmetic() bool {
	switch op {
	case AssignmentAddition,
		AssignmentSubtraction,
		AssignmentMultiplication,
		AssignmentDivision,
		AssignmentRemainder,
		AssignmentExponential:
		return true
	}
	return false
}

// IsBitwise reports whether op is one of |=, ^=, &=, <<=, >>=, >>>=.
func (op AssignmentOperator) IsBitwise() bool {
	switch op {
	case AssignmentShiftLeft,
		AssignmentShiftRight,
		AssignmentUnsignedShiftRight,
		AssignmentBitwiseOR,
		AssignmentBitwiseXOR,
		AssignmentBitwiseAnd:
		return true
	}
	return false
}

// IsLogical reports whether op is one of ||=, &&=, ??=.
func (op AssignmentOperator) IsLogical() bool {
	switch op {
	case AssignmentLogicalOr, AssignmentLogicalAnd, AssignmentLogicalNullish:
		return true
	}
	return false
}

// ToBinaryOperator returns the underlying [BinaryOperator] for compound
// arithmetic, bitwise, and shift assignments. ok is false for =, ||=, &&=, ??=.
func (op AssignmentOperator) ToBinaryOperator() (_ BinaryOperator, ok bool) {
	switch op {
	case AssignmentAddition:
		return BinaryAddition, true
	case AssignmentSubtraction:
		return BinarySubtraction, true
	case AssignmentMultiplication:
		return BinaryMultiplication, true
	case AssignmentDivision:
		return BinaryDivision, true
	case AssignmentRemainder:
		return BinaryRemainder, true
	case AssignmentExponential:
		return BinaryExponential, true
	case AssignmentShiftLeft:
		return BinaryShiftLeft, true
	case AssignmentShiftRight:
		return BinaryShiftRight, true
	case AssignmentUnsignedShiftRight:
		return BinaryUnsignedShiftRight, true
	case AssignmentBitwiseOR:
		return BinaryBitwiseOR, true
	case AssignmentBitwiseXOR:
		return BinaryBitwiseXOR, true
	case AssignmentBitwiseAnd:
		return BinaryBitwiseAnd, true
	}
	return 0, false
}

// ToLogicalOperator returns the underlying [LogicalOperator] for ||=, &&=, ??=.
// ok is false for any other assignment.
func (op AssignmentOperator) ToLogicalOperator() (_ LogicalOperator, ok bool) {
	switch op {
	case AssignmentLogicalOr:
		return LogicalOr, true
	case AssignmentLogicalAnd:
		return LogicalAnd, true
	case AssignmentLogicalNullish:
		return LogicalCoalesce, true
	}
	return 0, false
}

// BinaryOperator is the operator in a binary expression. The short-circuiting
// ||, &&, and ?? live separately on [LogicalOperator].
type BinaryOperator uint8

const (
	BinaryEquality           BinaryOperator = iota // ==
	BinaryInequality                               // !=
	BinaryStrictEquality                           // ===
	BinaryStrictInequality                         // !==
	BinaryLessThan                                 // <
	BinaryLessEqualThan                            // <=
	BinaryGreaterThan                              // >
	BinaryGreaterEqualThan                         // >=
	BinaryAddition                                 // +
	BinarySubtraction                              // -
	BinaryMultiplication                           // *
	BinaryDivision                                 // /
	BinaryRemainder                                // %
	BinaryExponential                              // **
	BinaryShiftLeft                                // <<
	BinaryShiftRight                               // >>
	BinaryUnsignedShiftRight                       // >>>
	BinaryBitwiseOR                                // |
	BinaryBitwiseXOR                               // ^
	BinaryBitwiseAnd                               // &
	BinaryIn                                       // in
	BinaryInstanceof                               // instanceof
)

// String returns op as it appears in source.
func (op BinaryOperator) String() string {
	switch op {
	case BinaryEquality:
		return "=="
	case BinaryInequality:
		return "!="
	case BinaryStrictEquality:
		return "==="
	case BinaryStrictInequality:
		return "!=="
	case BinaryLessThan:
		return "<"
	case BinaryLessEqualThan:
		return "<="
	case BinaryGreaterThan:
		return ">"
	case BinaryGreaterEqualThan:
		return ">="
	case BinaryAddition:
		return "+"
	case BinarySubtraction:
		return "-"
	case BinaryMultiplication:
		return "*"
	case BinaryDivision:
		return "/"
	case BinaryRemainder:
		return "%"
	case BinaryExponential:
		return "**"
	case BinaryShiftLeft:
		return "<<"
	case BinaryShiftRight:
		return ">>"
	case BinaryUnsignedShiftRight:
		return ">>>"
	case BinaryBitwiseOR:
		return "|"
	case BinaryBitwiseXOR:
		return "^"
	case BinaryBitwiseAnd:
		return "&"
	case BinaryIn:
		return "in"
	case BinaryInstanceof:
		return "instanceof"
	}
	return ""
}

// IsEquality reports whether op is one of ==, !=, ===, !==.
func (op BinaryOperator) IsEquality() bool {
	switch op {
	case BinaryEquality, BinaryInequality, BinaryStrictEquality, BinaryStrictInequality:
		return true
	}
	return false
}

// IsCompare reports whether op is one of <, <=, >, >=.
func (op BinaryOperator) IsCompare() bool {
	switch op {
	case BinaryLessThan, BinaryLessEqualThan, BinaryGreaterThan, BinaryGreaterEqualThan:
		return true
	}
	return false
}

// IsArithmetic reports whether op is one of +, -, *, /, %, **.
func (op BinaryOperator) IsArithmetic() bool {
	switch op {
	case BinaryAddition,
		BinarySubtraction,
		BinaryMultiplication,
		BinaryDivision,
		BinaryRemainder,
		BinaryExponential:
		return true
	}
	return false
}

// IsMultiplicative reports whether op is one of *, /, %.
func (op BinaryOperator) IsMultiplicative() bool {
	switch op {
	case BinaryMultiplication, BinaryDivision, BinaryRemainder:
		return true
	}
	return false
}

// IsBitwise reports whether op is any bitwise operator, including the bit shifts.
func (op BinaryOperator) IsBitwise() bool {
	if op.IsBitshift() {
		return true
	}
	switch op {
	case BinaryBitwiseOR, BinaryBitwiseXOR, BinaryBitwiseAnd:
		return true
	}
	return false
}

// IsBitshift reports whether op is one of <<, >>, >>>.
func (op BinaryOperator) IsBitshift() bool {
	switch op {
	case BinaryShiftLeft, BinaryShiftRight, BinaryUnsignedShiftRight:
		return true
	}
	return false
}

// IsNumericOrStringBinaryOperator reports whether op is an arithmetic or
// bitwise operator — i.e. one whose result is a number or string.
func (op BinaryOperator) IsNumericOrStringBinaryOperator() bool {
	return op.IsArithmetic() || op.IsBitwise()
}

// IsRelational reports whether op is in or instanceof.
func (op BinaryOperator) IsRelational() bool {
	switch op {
	case BinaryIn, BinaryInstanceof:
		return true
	}
	return false
}

// IsIn reports whether op is the in operator.
func (op BinaryOperator) IsIn() bool { return op == BinaryIn }

// IsInstanceOf reports whether op is the instanceof operator.
func (op BinaryOperator) IsInstanceOf() bool { return op == BinaryInstanceof }

// IsKeyword reports whether op is spelled as a keyword (in, instanceof) rather
// than punctuation.
func (op BinaryOperator) IsKeyword() bool {
	switch op {
	case BinaryIn, BinaryInstanceof:
		return true
	}
	return false
}

// CompareInverseOperator returns the operator that flips a comparison
// (< ↔ >, <= ↔ >=). ok is false for non-comparison operators.
func (op BinaryOperator) CompareInverseOperator() (_ BinaryOperator, ok bool) {
	switch op {
	case BinaryLessThan:
		return BinaryGreaterThan, true
	case BinaryLessEqualThan:
		return BinaryGreaterEqualThan, true
	case BinaryGreaterThan:
		return BinaryLessThan, true
	case BinaryGreaterEqualThan:
		return BinaryLessEqualThan, true
	}
	return 0, false
}

// EqualityInverseOperator returns the operator that negates an equality test
// (== ↔ !=, === ↔ !==). ok is false for non-equality operators.
func (op BinaryOperator) EqualityInverseOperator() (_ BinaryOperator, ok bool) {
	switch op {
	case BinaryEquality:
		return BinaryInequality, true
	case BinaryInequality:
		return BinaryEquality, true
	case BinaryStrictEquality:
		return BinaryStrictInequality, true
	case BinaryStrictInequality:
		return BinaryStrictEquality, true
	}
	return 0, false
}

// ToAssignmentOperator returns the compound-assignment form of op (e.g. + → +=).
// ok is false for the equality, comparison, and relational operators.
func (op BinaryOperator) ToAssignmentOperator() (_ AssignmentOperator, ok bool) {
	switch op {
	case BinaryAddition:
		return AssignmentAddition, true
	case BinarySubtraction:
		return AssignmentSubtraction, true
	case BinaryMultiplication:
		return AssignmentMultiplication, true
	case BinaryDivision:
		return AssignmentDivision, true
	case BinaryRemainder:
		return AssignmentRemainder, true
	case BinaryExponential:
		return AssignmentExponential, true
	case BinaryShiftLeft:
		return AssignmentShiftLeft, true
	case BinaryShiftRight:
		return AssignmentShiftRight, true
	case BinaryUnsignedShiftRight:
		return AssignmentUnsignedShiftRight, true
	case BinaryBitwiseOR:
		return AssignmentBitwiseOR, true
	case BinaryBitwiseXOR:
		return AssignmentBitwiseXOR, true
	case BinaryBitwiseAnd:
		return AssignmentBitwiseAnd, true
	}
	return 0, false
}

// Precedence returns op's precedence.
func (op BinaryOperator) Precedence() Precedence {
	switch op {
	case BinaryBitwiseOR:
		return PrecedenceBitwiseOr
	case BinaryBitwiseXOR:
		return PrecedenceBitwiseXor
	case BinaryBitwiseAnd:
		return PrecedenceBitwiseAnd
	case BinaryEquality, BinaryInequality, BinaryStrictEquality, BinaryStrictInequality:
		return PrecedenceEquals
	case BinaryLessThan,
		BinaryLessEqualThan,
		BinaryGreaterThan,
		BinaryGreaterEqualThan,
		BinaryInstanceof,
		BinaryIn:
		return PrecedenceCompare
	case BinaryShiftLeft, BinaryShiftRight, BinaryUnsignedShiftRight:
		return PrecedenceShift
	case BinarySubtraction, BinaryAddition:
		return PrecedenceAdd
	case BinaryMultiplication, BinaryRemainder, BinaryDivision:
		return PrecedenceMultiply
	case BinaryExponential:
		return PrecedenceExponentiation
	}
	return PrecedenceLowest
}

// LowerPrecedence returns the precedence one level below op's own.
// See [BinaryOperator.Precedence] for op's actual precedence.
func (op BinaryOperator) LowerPrecedence() Precedence {
	switch op {
	case BinaryBitwiseOR:
		return PrecedenceLogicalAnd
	case BinaryBitwiseXOR:
		return PrecedenceBitwiseOr
	case BinaryBitwiseAnd:
		return PrecedenceBitwiseXor
	case BinaryEquality, BinaryInequality, BinaryStrictEquality, BinaryStrictInequality:
		return PrecedenceBitwiseAnd
	case BinaryLessThan,
		BinaryLessEqualThan,
		BinaryGreaterThan,
		BinaryGreaterEqualThan,
		BinaryInstanceof,
		BinaryIn:
		return PrecedenceEquals
	case BinaryShiftLeft, BinaryShiftRight, BinaryUnsignedShiftRight:
		return PrecedenceCompare
	case BinaryAddition, BinarySubtraction:
		return PrecedenceShift
	case BinaryMultiplication, BinaryRemainder, BinaryDivision:
		return PrecedenceAdd
	case BinaryExponential:
		return PrecedenceMultiply
	}
	return PrecedenceLowest
}

// LogicalOperator is one of the short-circuiting binary operators.
type LogicalOperator uint8

const (
	LogicalOr       LogicalOperator = iota // ||
	LogicalAnd                             // &&
	LogicalCoalesce                        // ??
)

// String returns op as it appears in source.
func (op LogicalOperator) String() string {
	switch op {
	case LogicalOr:
		return "||"
	case LogicalAnd:
		return "&&"
	case LogicalCoalesce:
		return "??"
	}
	return ""
}

// IsOr reports whether op is ||.
func (op LogicalOperator) IsOr() bool { return op == LogicalOr }

// IsAnd reports whether op is &&.
func (op LogicalOperator) IsAnd() bool { return op == LogicalAnd }

// IsCoalesce reports whether op is ??.
func (op LogicalOperator) IsCoalesce() bool { return op == LogicalCoalesce }

// ToAssignmentOperator returns the matching ||=, &&=, or ??=.
func (op LogicalOperator) ToAssignmentOperator() AssignmentOperator {
	switch op {
	case LogicalOr:
		return AssignmentLogicalOr
	case LogicalAnd:
		return AssignmentLogicalAnd
	case LogicalCoalesce:
		return AssignmentLogicalNullish
	}
	return AssignmentAssign
}

// Precedence returns op's precedence.
func (op LogicalOperator) Precedence() Precedence {
	switch op {
	case LogicalOr:
		return PrecedenceLogicalOr
	case LogicalAnd:
		return PrecedenceLogicalAnd
	case LogicalCoalesce:
		return PrecedenceNullishCoalescing
	}
	return PrecedenceLowest
}

// LowerPrecedence returns the precedence one level below op's own.
func (op LogicalOperator) LowerPrecedence() Precedence {
	switch op {
	case LogicalOr:
		return PrecedenceNullishCoalescing
	case LogicalAnd:
		return PrecedenceLogicalOr
	case LogicalCoalesce:
		return PrecedenceConditional
	}
	return PrecedenceLowest
}

// UpdateOperator is the ++ or -- self-modifying unary operator.
type UpdateOperator uint8

const (
	UpdateIncrement UpdateOperator = iota // ++
	UpdateDecrement                       // --
)

// String returns op as it appears in source.
func (op UpdateOperator) String() string {
	switch op {
	case UpdateIncrement:
		return "++"
	case UpdateDecrement:
		return "--"
	}
	return ""
}

// UnaryOperator is the operator in a unary expression. The self-modifying ++
// and -- live on [UpdateOperator]. See https://tc39.es/ecma262/#sec-unary-operators.
type UnaryOperator uint8

const (
	UnaryPlus       UnaryOperator = iota // +
	UnaryNegation                        // -
	UnaryLogicalNot                      // !
	UnaryBitwiseNot                      // ~
	UnaryTypeof                          // typeof
	UnaryVoid                            // void
	UnaryDelete                          // delete
)

// String returns op as it appears in source.
func (op UnaryOperator) String() string {
	switch op {
	case UnaryPlus:
		return "+"
	case UnaryNegation:
		return "-"
	case UnaryLogicalNot:
		return "!"
	case UnaryBitwiseNot:
		return "~"
	case UnaryTypeof:
		return "typeof"
	case UnaryVoid:
		return "void"
	case UnaryDelete:
		return "delete"
	}
	return ""
}

// IsArithmetic reports whether op is unary + or -.
func (op UnaryOperator) IsArithmetic() bool {
	switch op {
	case UnaryPlus, UnaryNegation:
		return true
	}
	return false
}

// IsNot reports whether op is the ! operator.
func (op UnaryOperator) IsNot() bool { return op == UnaryLogicalNot }

// IsBitwise reports whether op is the ~ operator.
func (op UnaryOperator) IsBitwise() bool { return op == UnaryBitwiseNot }

// IsTypeof reports whether op is the typeof operator.
func (op UnaryOperator) IsTypeof() bool { return op == UnaryTypeof }

// IsVoid reports whether op is the void operator.
func (op UnaryOperator) IsVoid() bool { return op == UnaryVoid }

// IsDelete reports whether op is the delete operator.
func (op UnaryOperator) IsDelete() bool { return op == UnaryDelete }

// IsKeyword reports whether op is spelled as a keyword (typeof, void, delete)
// rather than punctuation.
func (op UnaryOperator) IsKeyword() bool {
	switch op {
	case UnaryTypeof, UnaryVoid, UnaryDelete:
		return true
	}
	return false
}

// Precedence is the precedence of an ECMAScript operator. Only the relative
// ordering is meaningful; the absolute values mirror esbuild's js_ast.OpPrec
// and are derived bottom-up from the ECMA grammar starting at the comma
// operator.
type Precedence uint8

const (
	PrecedenceLowest Precedence = iota
	PrecedenceComma
	PrecedenceSpread
	PrecedenceYield
	PrecedenceAssign
	PrecedenceConditional
	PrecedenceNullishCoalescing
	PrecedenceLogicalOr
	PrecedenceLogicalAnd
	PrecedenceBitwiseOr
	PrecedenceBitwiseXor
	PrecedenceBitwiseAnd
	PrecedenceEquals
	PrecedenceCompare
	PrecedenceShift
	PrecedenceAdd
	PrecedenceMultiply
	PrecedenceExponentiation
	PrecedencePrefix
	PrecedencePostfix
	PrecedenceNew
	PrecedenceCall
	PrecedenceMember
)

// IsRightAssociative reports whether p binds right-to-left.
func (p Precedence) IsRightAssociative() bool {
	switch p {
	case PrecedenceExponentiation, PrecedenceConditional, PrecedenceAssign:
		return true
	}
	return false
}

// IsLeftAssociative reports whether p binds left-to-right.
func (p Precedence) IsLeftAssociative() bool {
	switch p {
	case PrecedenceLowest,
		PrecedenceComma,
		PrecedenceSpread,
		PrecedenceYield,
		PrecedenceNullishCoalescing,
		PrecedenceLogicalOr,
		PrecedenceLogicalAnd,
		PrecedenceBitwiseOr,
		PrecedenceBitwiseXor,
		PrecedenceBitwiseAnd,
		PrecedenceEquals,
		PrecedenceCompare,
		PrecedenceShift,
		PrecedenceAdd,
		PrecedenceMultiply,
		PrecedencePrefix,
		PrecedencePostfix,
		PrecedenceNew,
		PrecedenceCall,
		PrecedenceMember:
		return true
	}
	return false
}

// IsBitwise reports whether p is one of the bitwise precedences (|, ^, &).
func (p Precedence) IsBitwise() bool {
	switch p {
	case PrecedenceBitwiseOr, PrecedenceBitwiseXor, PrecedenceBitwiseAnd:
		return true
	}
	return false
}

// IsShift reports whether p is the bit-shift precedence.
func (p Precedence) IsShift() bool { return p == PrecedenceShift }

// IsAdditive reports whether p is the additive precedence.
func (p Precedence) IsAdditive() bool { return p == PrecedenceAdd }
