package parser

import (
	"fmt"
	"github.com/nukilabs/unicodeid"
	"github.com/t14raptor/go-fast/token"
	"unicode"
)

type byteHandler func(p *parser) token.Token

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
func err(p *parser) token.Token {
	p.read()
	p.errorUnexpected(p.chr)
	return token.Undetermined
}

// <SPACE> <TAB> Normal Whitespace
func sps(p *parser) token.Token {
	p.read()
	return token.Skip
}

// <VT> <FF> Irregular Whitespace
func isp(p *parser) token.Token {
	p.read()
	return token.Skip
}

// '\r' '\n'
func lin(p *parser) token.Token {
	p.read()
	return token.Skip
}

// !
func exl(p *parser) token.Token {
	p.read()
	if p.advanceIfAsciiEquals('=') {
		if p.advanceIfAsciiEquals('=') {
			return token.StrictNotEqual
		}
		return token.NotEqual
	}
	return token.Not
}

// "
func qod(p *parser) token.Token {
	// SAFETY: This function is only called for `"`
	// String literal
	p.read()

	p.read()
	var err string
	p.literal, p.parsedLiteral, err = p.scanString(p.chrOffset-1, true)
	if err != "" {
		return token.Illegal
	}
	return token.String
}

// '
func qos(p *parser) token.Token {
	// SAFETY: This function is only called for `"`
	// String literal
	p.insertSemicolon = true
	p.read()

	p.read()
	var err string
	p.literal, p.parsedLiteral, err = p.scanString(p.chrOffset-1, true)
	if err != "" {
		return token.Illegal
	}
	return token.String
}

// #
func has(p *parser) token.Token {
	// Possible shebang (#!)
	if p.chrOffset == 1 && p.chr == '!' {
		p.skipSingleLineComment()
		return token.Skip
	}
	// Otherwise, private identifier
	var err string
	literal, _, err := p.scanIdentifierTail(p.offset)
	if err != "" || literal == "" {
		return token.Illegal
	}
	p.insertSemicolon = true
	return token.PrivateIdentifier
}

// `A..=Z`, `a..=z` (except special cases below), `_`, `$`
func idt(p *parser) token.Token {
	p.literal, _, _ = p.scanIdentifierTail(p.offset)
	return token.Identifier
}

// %
func prc(p *parser) token.Token {
	p.read()
	if p.advanceIfAsciiEquals('=') {
		return token.RemainderAssign
	}
	return token.Remainder
}

// &
func amp(p *parser) token.Token {
	p.read()
	if p.advanceIfAsciiEquals('&') {
		if p.advanceIfAsciiEquals('=') {
			return token.LogicalAnd
		}
		return token.LogicalAnd
	} else if p.advanceIfAsciiEquals('=') {
		return token.AndAssign
	}
	return token.And
}

// (
func pno(p *parser) token.Token {
	p.read()
	return token.LeftParenthesis
}

// )
func pnc(p *parser) token.Token {
	p.read()
	return token.RightParenthesis
}

// *
func atr(p *parser) token.Token {
	p.read()
	if p.advanceIfAsciiEquals('*') {
		if p.advanceIfAsciiEquals('=') {
			return token.ExponentAssign
		}
		return token.Exponent
	} else if p.advanceIfAsciiEquals('=') {
		return token.MultiplyAssign
	}
	return token.Multiply
}

// +
func pls(p *parser) token.Token {
	p.read()
	if p.advanceIfAsciiEquals('+') {
		return token.Increment
	} else if p.advanceIfAsciiEquals('=') {
		return token.AddAssign
	}
	return token.Plus
}

// ,
func com(p *parser) token.Token {
	p.read()
	return token.Comma
}

// -
func min(p *parser) token.Token {
	p.read()
	if p.advanceIfAsciiEquals('-') {
		return token.Decrement
	} else if p.advanceIfAsciiEquals('=') {
		return token.SubtractAssign
	}
	return token.Minus
}

// .
func prd(p *parser) token.Token {
	p.read()

	if digitValue(p.chr) < 10 {
		p.insertSemicolon = true
		tkn, literal := p.scanNumericLiteral(true)
		p.literal = literal
		return tkn
	}
	if p.advanceIfAsciiEquals('.') {
		if p.advanceIfAsciiEquals('.') {
			return token.Ellipsis
		}
		return token.Illegal
	}
	return token.Period
}

// /
func slh(p *parser) token.Token {
	p.read()
	b, ok := p._peekByte()
	if !ok {
		return token.Eof
	}
	switch b {
	case '/':
		p.read()
		p.skipSingleLineComment()
		return token.Skip
	case '*':
		p.read()
		p.skipMultiLineComment()
		return token.Skip
	}
	// regex is handled separately, see `next_regex`
	if p.advanceIfAsciiEquals('=') {
		return token.QuotientAssign
	}
	return token.Slash
}

// 0
func zer(p *parser) token.Token {
	p.read()
	return p.readZero()
}

// 1 to 9
func dig(p *parser) token.Token {
	p.scanNumericLiteral(false)
	return token.Number
}

// :
func col(p *parser) token.Token {
	p.read()
	return token.Colon
}

// ;
func sem(p *parser) token.Token {
	p.read()
	return token.Semicolon
}

// <
func lss(p *parser) token.Token {
	p.read()

	if p.advanceIfAsciiEquals('=') {
		return token.ShiftLeft
	} else if p.advanceIfAsciiEquals('<') {
		if p.advanceIfAsciiEquals('=') {
			return token.ShiftLeftAssign
		}
		return token.LessOrEqual
	}
	return token.Less
}

// =
func eql(p *parser) token.Token {
	p.read()
	if p.advanceIfAsciiEquals('=') {
		if p.advanceIfAsciiEquals('=') {
			return token.StrictEqual
		}
		return token.Equal
	} else if p.advanceIfAsciiEquals('>') {
		return token.Arrow
	}
	return token.Assign
}

// >
func gtr(p *parser) token.Token {
	p.read()
	if p.advanceIfAsciiEquals('=') {
		return token.GreaterOrEqual
	}
	if p.advanceIfAsciiEquals('>') {
		if p.advanceIfAsciiEquals('=') {
			return token.ShiftRightAssign
		}
		if p.advanceIfAsciiEquals('>') {
			if p.advanceIfAsciiEquals('=') {
				return token.UnsignedShiftRightAssign
			}
			return token.UnsignedShiftRight
		}
		return token.ShiftRight
	}
	return token.Greater
}

// ?
func qst(p *parser) token.Token {
	p.read()

	next2Bytes, ok := p.peek2Bytes()
	if ok {
		switch next2Bytes[0] {
		case '?':
			if next2Bytes[1] == '=' {
				p.read()
				p.read()
				return token.Coalesce
			}
			p.read()
			return token.Coalesce
			// parse `?.1` as `?` `.1`
		case '.':
			if !unicode.IsDigit(rune(next2Bytes[1])) {
				p.read()
				return token.QuestionDot
			}
		}
		return token.QuestionMark
	}

	// At EOF, or only 1 byte left
	nextByte, ok := p._peekByte()
	if !ok {
		return token.Eof
	}
	switch nextByte {
	case '?':
		p.read()
		return token.Coalesce
	case '.':
		p.read()
		return token.QuestionDot
	}
	return token.QuestionMark
}

// [
func bto(p *parser) token.Token {
	p.read()
	return token.LeftBracket
}

// \
func esc(p *parser) token.Token {
	// todo
	return token.Identifier
}

// ]
func btc(p *parser) token.Token {
	p.read()
	return token.RightBracket
}

// ^
func crt(p *parser) token.Token {
	p.read()
	if p.advanceIfAsciiEquals('=') {
		return token.ExclusiveOrAssign
	}
	return token.ExclusiveOr
}

// `
func tpl(p *parser) token.Token {
	p.read()
	return token.Backtick
}

// {
func beo(p *parser) token.Token {
	p.read()
	return token.LeftBrace
}

// |
func pip(p *parser) token.Token {
	p.read()

	b, ok := p._peekByte()
	if !ok {
		return token.Eof
	}
	switch b {
	case '|':
		p.read()
		return token.LogicalOr
	case '=':
		p.read()
		return token.OrAssign
	}
	return token.Or
}

// }
func bec(p *parser) token.Token {
	p.read()
	return token.RightBrace
}

// ~
func tld(p *parser) token.Token {
	p.read()
	return token.BitwiseNot
}

func l_a(p *parser) token.Token {
	p.literal, _, _ = p.scanIdentifierTail(p.offset)
	switch p.literal {
	case "await":
		return token.Await
	case "async":
		return token.Async
	}
	return token.Identifier
}

func l_b(p *parser) token.Token {
	p.literal, _, _ = p.scanIdentifierTail(p.offset)
	switch p.literal {
	case "break":
		return token.Break
	case "boolean":
		return token.Boolean
	}
	return token.Identifier
}

func l_c(p *parser) token.Token {
	p.literal, _, _ = p.scanIdentifierTail(p.offset)
	switch p.literal {
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

func l_d(p *parser) token.Token {
	p.literal, _, _ = p.scanIdentifierTail(p.offset)
	switch p.literal {
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

func l_e(p *parser) token.Token {
	p.literal, _, _ = p.scanIdentifierTail(p.offset)
	switch p.literal {
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

func l_f(p *parser) token.Token {
	p.literal, _, _ = p.scanIdentifierTail(p.offset)
	fmt.Println(p.literal, "hellooo")
	switch p.literal {
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

func l_i(p *parser) token.Token {
	p.literal, _, _ = p.scanIdentifierTail(p.offset)
	switch p.literal {
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

func l_l(p *parser) token.Token {
	p.literal, _, _ = p.scanIdentifierTail(p.offset)
	switch p.literal {
	case "let":
		return token.Let
	}
	return token.Identifier
}

func l_n(p *parser) token.Token {
	p.literal, _, _ = p.scanIdentifierTail(p.offset)
	switch p.literal {
	case "new":
		return token.New
	case "null":
		return token.Null
	case "number":
		return token.Number
	}
	return token.Identifier
}

func l_o(p *parser) token.Token {
	p.literal, _, _ = p.scanIdentifierTail(p.offset)
	switch p.literal {
	case "of":
		return token.Of
	}
	return token.Identifier
}

func l_r(p *parser) token.Token {
	p.literal, _, _ = p.scanIdentifierTail(p.offset)
	switch p.literal {
	case "return":
		return token.Return
	}
	return token.Identifier
}

func l_s(p *parser) token.Token {
	p.literal, _, _ = p.scanIdentifierTail(p.offset)
	switch p.literal {
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

func l_t(p *parser) token.Token {
	p.literal, _, _ = p.scanIdentifierTail(p.offset)
	switch p.literal {
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

func l_v(p *parser) token.Token {
	p.literal, _, _ = p.scanIdentifierTail(p.offset)
	switch p.literal {
	case "var":
		return token.Var
	case "void":
		return token.Void
	}
	return token.Identifier
}

func l_w(p *parser) token.Token {
	p.literal, _, _ = p.scanIdentifierTail(p.offset)
	switch p.literal {
	case "while":
		return token.While
	case "with":
		return token.With
	}
	return token.Identifier
}

func l_y(p *parser) token.Token {
	p.literal, _, _ = p.scanIdentifierTail(p.offset)
	switch p.literal {
	case "yield":
		return token.Yield
	}
	return token.Identifier
}

// Non-ASCII characters.
// NB: Must not use `ascii_byte_handler!` macro, as this handler is for non-ASCII chars.
func uni(p *parser) token.Token {
	switch c := p._peek(); {
	case unicodeid.IsIDStartUnicode(c):
		p.scanIdentifierTail(p.offset)
		return token.Identifier
	case unicode.IsSpace(c):
		p.read()
		return token.Skip
	case isLineTerminator(c):
		p.read()
		return token.Skip
	default:
		p.read()
		p.errorUnexpected(c)
		return token.Undetermined
	}
}

// UTF-8 continuation bytes (0x80 - 0xBF) (i.e. middle of a multi-byte UTF-8 sequence)
// + and byte values which are not legal in UTF-8 strings (0xC0, 0xC1, 0xF5 - 0xFF).
// `handle_byte()` should only be called with 1st byte of a valid UTF-8 character,
// so something has gone wrong if we get here.
// https://datatracker.ietf.org/doc/html/rfc3629
// NB: Must not use `ascii_byte_handler!` macro, as this handler is for non-ASCII bytes.
func uer(p *parser) token.Token {
	panic("unreachable")
}
