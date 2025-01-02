package scanner

import (
	"github.com/nukilabs/unicodeid"
	"github.com/t14raptor/go-fast/token"
	"unicode"
)

type byteHandler func(s *Scanner) token.Token

var byteHandlers = [256]byteHandler{
	//0    1    2    3    4    5    6    7    8    9    A    B    C    D    E    F
	err, err, err, err, err, err, err, err, err, sps, lin, isp, isp, lin, err, err, // 0
	err, err, err, err, err, err, err, err, err, err, err, err, err, err, err, err, // 1
	sps, exl, qod, has, idt, prc, amp, qos, pno, pnc, atr, pls, com, min, prd, slh, // 2
	zer, dig, dig, dig, dig, dig, dig, dig, dig, dig, col, sem, lss, eql, gtr, qst, // 3
	err, idt, idt, idt, idt, idt, idt, idt, idt, idt, idt, idt, idt, idt, idt, idt, // 4
	idt, idt, idt, idt, idt, idt, idt, idt, idt, idt, idt, bto, esc, btc, crt, idt, // 5
	tpl, l_a, l_b, l_c, l_d, l_e, l_f, idt, idt, l_i, idt, idt, l_l, idt, l_n, l_o, // 6
	idt, idt, l_r, l_s, l_t, idt, l_v, l_w, idt, l_y, idt, beo, pip, bec, tld, err, // 7
	uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, // 8
	uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, // 9
	uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, // A
	uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, // B
	uer, uer, uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, // C
	uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, // D
	uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, uni, // E
	uni, uni, uni, uni, uni, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, uer, // F
}

// `\0` `\1` etc
func err(s *Scanner) token.Token {
	s.ConsumeByte()
	//p.errorUnexpected(p.chr) TODO
	return token.Undetermined
}

// <SPACE> <TAB> Normal Whitespace
func sps(s *Scanner) token.Token {
	s.ConsumeByte()
	return token.Skip
}

// <VT> <FF> Irregular Whitespace
func isp(s *Scanner) token.Token {
	s.ConsumeByte()
	return token.Skip
}

// '\r' '\n'
func lin(s *Scanner) token.Token {
	s.ConsumeByte()
	return s.handleLineBreak()
}

// !
func exl(s *Scanner) token.Token {
	s.ConsumeByte()
	if s.AdvanceIfByteEquals('=') {
		if s.AdvanceIfByteEquals('=') {
			return token.StrictNotEqual
		}
		return token.NotEqual
	}
	return token.Not
}

// "
func qod(s *Scanner) token.Token {
	// SAFETY: This function is only called for `"`
	// String literal
	return s.scanStringLiteralDoubleQuote()
}

// '
func qos(s *Scanner) token.Token {
	// SAFETY: This function is only called for `"`
	// String literal
	return s.scanStringLiteralSingleQuote()
}

// #
func has(s *Scanner) token.Token {
	// Possible shebang (#!)
	//if p.chrOffset == 1 && p.chr == '!' {
	//	s.skipSingleLineComment()
	//		return token.Skip
	//	}
	// Otherwise, private identifier
	s.scanIdentifierTail()
	return token.PrivateIdentifier
}

// `A..=Z`, `a..=z` (except special cases below), `_`, `$`
func idt(s *Scanner) token.Token {
	s.scanIdentifierTail()
	return token.Identifier
}

// %
func prc(s *Scanner) token.Token {
	s.ConsumeByte()
	if s.AdvanceIfByteEquals('=') {
		return token.RemainderAssign
	}
	return token.Remainder
}

// &
func amp(s *Scanner) token.Token {
	s.ConsumeByte()
	if s.AdvanceIfByteEquals('&') {
		if s.AdvanceIfByteEquals('=') {
			// TODO
			return token.LogicalAnd
		}
		return token.LogicalAnd
	} else if s.AdvanceIfByteEquals('=') {
		return token.AndAssign
	}
	return token.And
}

// (
func pno(s *Scanner) token.Token {
	s.ConsumeByte()
	return token.LeftParenthesis
}

// )
func pnc(s *Scanner) token.Token {
	s.ConsumeByte()
	return token.RightParenthesis
}

// *
func atr(s *Scanner) token.Token {
	s.ConsumeByte()
	if s.AdvanceIfByteEquals('*') {
		if s.AdvanceIfByteEquals('=') {
			return token.ExponentAssign
		}
		return token.Exponent
	} else if s.AdvanceIfByteEquals('=') {
		return token.MultiplyAssign
	}
	return token.Multiply
}

// +
func pls(s *Scanner) token.Token {
	s.ConsumeByte()
	if s.AdvanceIfByteEquals('+') {
		return token.Increment
	} else if s.AdvanceIfByteEquals('=') {
		return token.AddAssign
	}
	return token.Plus
}

// ,
func com(s *Scanner) token.Token {
	s.ConsumeByte()
	return token.Comma
}

// -
func min(s *Scanner) token.Token {
	s.ConsumeByte()
	if s.AdvanceIfByteEquals('-') {
		return token.Decrement
	} else if s.AdvanceIfByteEquals('=') {
		return token.SubtractAssign
	}
	return token.Minus
}

// .
func prd(s *Scanner) token.Token {
	s.ConsumeByte()
	return s.readDot()
}

// /
func slh(s *Scanner) token.Token {
	s.ConsumeByte()
	b, ok := s.PeekByte()
	if !ok {
		return token.Eof
	}
	switch b {
	case '/':
		s.ConsumeByte()
		s.skipSingleLineComment()
		return token.Skip
	case '*':
		s.ConsumeByte()
		s.skipMultiLineComment()
		return token.Skip
	}
	// regex is handled separately, see `next_regex`
	if s.AdvanceIfByteEquals('=') {
		return token.QuotientAssign
	}
	return token.Slash
}

// 0
func zer(s *Scanner) token.Token {
	s.ConsumeByte()
	return s.readZero()
}

// 1 to 9
func dig(s *Scanner) token.Token {
	s.ConsumeByte()
	return s.decimalLiteralAfterFirstDigit()
}

// :
func col(s *Scanner) token.Token {
	s.ConsumeByte()
	return token.Colon
}

// ;
func sem(s *Scanner) token.Token {
	s.ConsumeByte()
	return token.Semicolon
}

// <
func lss(s *Scanner) token.Token {
	s.ConsumeByte()

	if s.AdvanceIfByteEquals('=') {
		return token.ShiftLeft
	} else if s.AdvanceIfByteEquals('<') {
		if s.AdvanceIfByteEquals('=') {
			return token.ShiftLeftAssign
		}
		return token.LessOrEqual
	}
	return token.Less
}

// =
func eql(s *Scanner) token.Token {
	s.ConsumeByte()
	if s.AdvanceIfByteEquals('=') {
		if s.AdvanceIfByteEquals('=') {
			return token.StrictEqual
		}
		return token.Equal
	} else if s.AdvanceIfByteEquals('>') {
		return token.Arrow
	}
	return token.Assign
}

// >
func gtr(s *Scanner) token.Token {
	s.ConsumeByte()
	if s.AdvanceIfByteEquals('=') {
		return token.GreaterOrEqual
	}
	if s.AdvanceIfByteEquals('>') {
		if s.AdvanceIfByteEquals('=') {
			return token.ShiftRightAssign
		}
		if s.AdvanceIfByteEquals('>') {
			if s.AdvanceIfByteEquals('=') {
				return token.UnsignedShiftRightAssign
			}
			return token.UnsignedShiftRight
		}
		return token.ShiftRight
	}
	return token.Greater
}

// ?
func qst(s *Scanner) token.Token {
	s.ConsumeByte()

	next2Bytes, ok := s.src.PeekTwoBytes()
	if ok {
		switch next2Bytes[0] {
		case '?':
			if next2Bytes[1] == '=' {
				s.ConsumeByte()
				s.ConsumeByte()
				return token.Coalesce
			}
			s.ConsumeByte()
			return token.Coalesce
			// parse `?.1` as `?` `.1`
		case '.':
			if !unicode.IsDigit(rune(next2Bytes[1])) {
				s.ConsumeByte()
				return token.QuestionDot
			}
		}
		return token.QuestionMark
	}

	// At EOF, or only 1 byte left
	nextByte, ok := s.PeekByte()
	if !ok {
		return token.Eof
	}
	switch nextByte {
	case '?':
		s.ConsumeByte()
		return token.Coalesce
	case '.':
		s.ConsumeByte()
		return token.QuestionDot
	}
	return token.QuestionMark
}

// [
func bto(s *Scanner) token.Token {
	s.ConsumeByte()
	return token.LeftBracket
}

// \
func esc(s *Scanner) token.Token {
	return s.identifierBackslashHandler()
}

// ]
func btc(s *Scanner) token.Token {
	s.ConsumeByte()
	return token.RightBracket
}

// ^
func crt(s *Scanner) token.Token {
	s.ConsumeByte()
	if s.AdvanceIfByteEquals('=') {
		return token.ExclusiveOrAssign
	}
	return token.ExclusiveOr
}

// `
func tpl(s *Scanner) token.Token {
	s.ConsumeByte()
	return token.Backtick
}

// {
func beo(s *Scanner) token.Token {
	s.ConsumeByte()
	return token.LeftBrace
}

// |
func pip(s *Scanner) token.Token {
	s.ConsumeByte()

	b, ok := s.PeekByte()
	if !ok {
		return token.Eof
	}
	switch b {
	case '|':
		s.ConsumeByte()
		return token.LogicalOr
	case '=':
		s.ConsumeByte()
		return token.OrAssign
	}
	return token.Or
}

// }
func bec(s *Scanner) token.Token {
	s.ConsumeByte()
	return token.RightBrace
}

// ~
func tld(s *Scanner) token.Token {
	s.ConsumeByte()
	return token.BitwiseNot
}

// a
func l_a(s *Scanner) token.Token {
	switch s.scanIdentifierTail() {
	case "await":
		return token.Await
	case "async":
		return token.Async
	}
	return token.Identifier
}

// b
func l_b(s *Scanner) token.Token {
	switch s.scanIdentifierTail() {
	case "break":
		return token.Break
	case "boolean":
		return token.Boolean
	}
	return token.Identifier
}

// c
func l_c(s *Scanner) token.Token {
	switch s.scanIdentifierTail() {
	case "case":
		return token.Case
	case "catch":
		return token.Catch
	case "class":
		return token.Class
	case "const":
		return token.Const
	case "continue":
		return token.Continue
	}
	return token.Identifier
}

// d
func l_d(s *Scanner) token.Token {
	switch s.scanIdentifierTail() {
	case "debugger":
		return token.Debugger
	case "default":
		return token.Default
	case "delete":
		return token.Delete
	case "do":
		return token.Do
	}
	return token.Identifier
}

// e
func l_e(s *Scanner) token.Token {
	switch s.scanIdentifierTail() {
	case "else":
		return token.Else
	case "enum":
		return token.Keyword
	case "export":
		return token.Keyword
	case "extends":
		return token.Extends
	}
	return token.Identifier
}

// f
func l_f(s *Scanner) token.Token {
	switch s.scanIdentifierTail() {
	case "false":
		return token.Boolean
	case "finally":
		return token.Finally
	case "for":
		return token.For
	case "function":
		return token.Function
	}
	return token.Identifier
}

// i
func l_i(s *Scanner) token.Token {
	switch s.scanIdentifierTail() {
	case "if":
		return token.If
	case "import":
		return token.Keyword
	case "in":
		return token.In
	case "instanceof":
		return token.InstanceOf
	}
	return token.Identifier
}

// l
func l_l(s *Scanner) token.Token {
	switch s.scanIdentifierTail() {
	case "let":
		return token.Let
	}
	return token.Identifier
}

// n
func l_n(s *Scanner) token.Token {
	switch s.scanIdentifierTail() {
	case "new":
		return token.New
	case "null":
		return token.Null
	case "number":
		return token.Number
	}
	return token.Identifier
}

// o
func l_o(s *Scanner) token.Token {
	switch s.scanIdentifierTail() {
	case "of":
		return token.Of
	}
	return token.Identifier
}

// r
func l_r(s *Scanner) token.Token {
	switch s.scanIdentifierTail() {
	case "return":
		return token.Return
	}
	return token.Identifier
}

// s
func l_s(s *Scanner) token.Token {
	switch s.scanIdentifierTail() {
	case "super":
		return token.Super
	case "static":
		return token.Static
	case "switch":
		return token.Switch
	case "string":
		return token.String
	}
	return token.Identifier
}

// t
func l_t(s *Scanner) token.Token {
	switch s.scanIdentifierTail() {
	case "this":
		return token.This
	case "throw":
		return token.Throw
	case "true":
		return token.Boolean
	case "typeof":
		return token.Typeof
	case "try":
		return token.Try
	}
	return token.Identifier
}

// v
func l_v(s *Scanner) token.Token {
	switch s.scanIdentifierTail() {
	case "var":
		return token.Var
	case "void":
		return token.Void
	}
	return token.Identifier
}

// w
func l_w(s *Scanner) token.Token {
	switch s.scanIdentifierTail() {
	case "while":
		return token.While
	case "with":
		return token.With
	}
	return token.Identifier
}

// y
func l_y(s *Scanner) token.Token {
	switch s.scanIdentifierTail() {
	case "yield":
		return token.Yield
	}
	return token.Identifier
}

// Non-ASCII characters.
// NB: Must not use `ascii_byte_handler!` macro, as this handler is for non-ASCII chars.
func uni(s *Scanner) token.Token {
	switch c, _ := s.PeekRune(); {
	case unicodeid.IsIDStartUnicode(c):
		s.scanIdentifierTailAfterUnicode(s.src.Offset())
		return token.Identifier
	case unicode.IsSpace(c):
		s.ConsumeRune()
		return token.Skip
	case isLineTerminator(c):
		s.ConsumeRune()
		s.token.OnNewLine = true
		return token.Skip
	default:
		s.ConsumeRune()
		//p.errorUnexpected(c) TODO
		return token.Undetermined
	}
}

// UTF-8 continuation bytes (0x80 - 0xBF) (i.e. middle of a multi-byte UTF-8 sequence)
// + and byte values which are not legal in UTF-8 strings (0xC0, 0xC1, 0xF5 - 0xFF).
// `handle_byte()` should only be called with 1st byte of a valid UTF-8 character,
// so something has gone wrong if we get here.
// https://datatracker.ietf.org/doc/html/rfc3629
// NB: Must not use `ascii_byte_handler!` macro, as this handler is for non-ASCII bytes.
func uer(s *Scanner) token.Token {
	panic("unreachable")
}
