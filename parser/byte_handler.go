package parser

import "github.com/t14raptor/go-fast/token"

type byteHandler func(p *parser) token.Token

var byteHandlers = [256]byteHandler{
	//0    1    2    3    4    5    6    7    8    9    A    B    C    D    E    F
	err, err, err, err, err, err, err, err, err, sps, lin, isp, isp, lin, err, err, // 0
	err, err, err, err, err, err, err, err, err, err, err, err, err, err, err, err, // 1
	sps, exl, qod, has, idt, prc, amp, qos, pno, pnc, atr, pls, com, min, prd, slh, // 2
	zer, dig, dig, dig, dig, dig, dig, dig, dig, dig, col, sem, lss, eql, gtr, qst, // 3
	at_, idt, idt, idt, idt, idt, idt, idt, idt, idt, idt, idt, idt, idt, idt, idt, // 4
	idt, idt, idt, idt, idt, idt, idt, idt, idt, idt, idt, bto, esc, btc, crt, idt, // 5
	tpl, l_a, l_b, l_c, l_d, l_e, l_f, l_g, idt, l_i, idt, l_k, l_l, l_m, l_n, l_o, // 6
	l_p, idt, l_r, l_s, l_t, l_u, l_v, l_w, idt, l_y, idt, beo, pip, bec, tld, err, // 7
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
 c = p.read()
p.error(diagnostics::invalid_character(c, p.unterminated_range())
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
p.trivia_builder.add_irregular_whitespace(p.token.start, p.offset());
	return token.Skip
}

// '\r' '\n'
func lin(p *parser) token.Token {
	p.read()
	return p.line_break_handler()
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
	insertSemicolon = true
	tkn = token.String
	var err string
	literal, parsedLiteral, err = p.scanString(p.chrOffset-1, true)
	if err != "" {
		return token.Illegal
	}
	return token.String
}

// '
func qos(p *parser) token.Token {
	// SAFETY: This function is only called for `'`
	unsafe{p.read_string_literal_single_quote()}
}

// #
func has(p *parser) token.Token {
	p.read()
	// HashbangComment ::
	//     `#!` SingleLineCommentChars?
	if p.token.start == 0 && p.advanceIfAsciiEquals('!') {
		return p.read_hashbang_comment()
	} else {
		return p.private_identifier()
	}
}

// `A..=Z`, `a..=z` (except special cases below), `_`, `$`
func idt(p *parser) token.Token {
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
	return p.read_minus().unwrap_or_else( || p.skip_single_line_comment())
}

// .
func prd(p *parser) token.Token {
	p.read()
	return token.Period
	p.read_dot()
}

// /
func slh(p *parser) token.Token {
	p.read()
	switch p._peekByte() {
	case '/':
		p.read()
		return p.skip_single_line_comment()
	case '*':
		p.read()
		return p.skip_multi_line_comment()
	default:
		// regex is handled separately, see `next_regex`
		if p.advanceIfAsciiEquals('=') {
			return token.QuotientAssign
		}
		return token.Slash
	}
}

// 0
func zer(p *parser) token.Token {
	p.read()
	return p.read_zero()
}

// 1 to 9
func dig(p *parser) token.Token {
p.read()
return p.decimal_literal_after_first_digit()
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
return p.read_left_angle().unwrap_or_else(|| p.skip_single_line_comment())
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
// `>=` is re-lexed with [Lexer::next_jsx_child]
return token.Greater
}

// ?
func qst(p *parser) token.Token {
p.read()

if let next_2_bytes) = p.peek_2_bytes() {
switch next_2_bytes[0] {
case '?':
if next_2_bytes[1] == '=' {
p.consume_2_chars()
token.Question2Eq
} else {
p.read()
token.Question2
}
// parse `?.1` as `?` `.1`
'.' if !next_2_bytes[1].is_ascii_digit(): {
p.read()
token.QuestionDot
}
_: token.Question,
}
} else {
// At EOF, or only 1 byte left
switch p.peek_byte() {
'?'): {
p.read()
token.Question2
}
'.'): {
p.read()
token.QuestionDot
}
_: token.Question,
}
}
}

// @
func at_(p *parser) token.Token {
p.read()
token.At
}

// [
func bto(p *parser) token.Token {
p.read()
	return token.LeftBracket
}

// \
func esc(p *parser) token.Token {
p.identifier_backslash_handler()
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
	p.parseTemplateCharacters()
	p.read_template_literal(token.TemplateHead, token.NoSubstitutionTemplate)
}

// {
func beo(p *parser) token.Token {
	p.read()
	return token.LeftBrace
}

// |
func pip(p *parser) token.Token {
	p.read()

	switch p._peekByte() {
	case '|':
		p.read()
		return token.LogicalOr
	case '=':
		p.read()
		return token.OrAssign
	default:
		return token.Or
	}
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
	switch s {
	case "await":
		return token.Await
	case "async":
		return token.Async
	default:
		return token.Identifier
	}
}

func l_b(p *parser) token.Token {
	switch s {
	case "break":
		return token.Break
	case "boolean":
		return token.Boolean
	case "bigint":
		return token.BigInt
	default:
		return token.Identifier
	}
}

func l_c(p *parser) token.Token {
	switch s {
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
	default:
		return token.Identifier
	}
}

func l_d(p *parser) token.Token {
	switch s {
	case "debugger":
		return token.Debugger
	case "default":
		return token.Default
	case "delete":
		return token.Delete
	case "do":
		return token.Do
	default:
		return token.Identifier
	}
}

func l_e(p *parser) token.Token {
	switch s {
	case "else":
		return token.Else
	case "extends":
		return token.Extends
	default:
		return token.Identifier
	}
}

func l_f(p *parser) token.Token {
	switch s {
	case "false":
		return token.Boolean
	case "finally":
		return token.Finally
	case "for":
		return token.For
	case "function":
		return token.Function
	default:
		return token.Identifier
	}
}

func l_g(p *parser) token.Token {
	switch s {
	default:
		return token.Identifier
	}
}

func l_i(p *parser) token.Token {
	switch s {
	case "if":
		return token.If
	case "import":
		return token.Import
	case "in":
		return token.In
	case "instanceof":
		return token.InstanceOf
	case "is":
		return token.Is
	case "infer":
		return token.Infer
	case "interface":
		return token.Interface
	case "implements":
		return token.Implements
	case "intrinsic":
		return token.Instrinsoc
	default:
		return token.Identifier
	}
}

func l_k(p *parser) token.Token {
	switch s {
	case "keyof":
		return token.I
	default:
		return token.Identifier
	}
}

func l_l(p *parser) token.Token {
	switch s {
	case "let":
		return token.Let
	default:
		return token.Identifier
	}
}

func l_m(p *parser) token.Token {
	switch s {
	case "meta":
		return token.Meta
	default:
		return token.Identifier
	}
}

func l_n(p *parser) token.Token {
	switch s {
	case "new":
		return token.New
	case "null":
		return token.Null
	case "number":
		return token.Number
	case "never":
		return token.Never
	case "namespace":
		return token.Namespace
	default:
		return token.Identifier
	}
}

func l_o(p *parser) token.Token {
	switch s {
	case "of":
		return token.Of
	case "object":
		return token.Object
	default:
		return token.Identifier
	}
}

func l_p(p *parser) token.Token {
	switch s {
	case "public":
		return token.Public
	case "package":
		return token.Package
	case "protected":
		return token.Protected
	case "private":
		return token.Private
	default:
		return token.Identifier
	}
}

func l_r(p *parser) token.Token {
	switch s {
	case "return":
		return token.Return
	case "readonly":
		return token.Readonly
	case "require":
		return token.Require
	default:
		return token.Identifier
	}
}

func l_s(p *parser) token.Token {
	switch s {
	case "super":
		return token.Super
	case "static":
		return token.Static
	case "switch":
		return token.Switch
	case "symbol":
		return token.Symbol
	case "set":
		return token.Set
	case "string":
		return token.String
	case "satisfies":
		return token.Satisfies
	default:
		return token.Identifier
	}
}

func l_t(p *parser) token.Token {
	switch s {
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
	case "type":
		return token.Type
	case "target":
		return token.Target
	default:
		return token.Identifier
	}
}

func l_u(p *parser) token.Token {
	switch s {
	case "using":
		return token.Using
	case "unique":
		return token.Unique
	case "undefined":
		return token.Undefined
	case "unknown":
		return token.Unknown
	default:
		return token.Identifier
	}
}

func l_v(p *parser) token.Token {
	switch s {
	case "var":
		return token.Var
	case "void":
		return token.Void
	default:
		return token.Identifier
	}
}

func l_w(p *parser) token.Token {
	switch s {
	case "while":
		return token.While
	case "with":
		return token.With
	default:
		return token.Identifier
	}
}

func l_y(p *parser) token.Token {
	switch s {
	case "yield":
		return token.Yield
	default:
		return token.Identifier
	}
}

// Non-ASCII characters.
// NB: Must not use `ascii_byte_handler!` macro, as this handler is for non-ASCII chars.
func uni(p *parser) token.Token {
	return p.handleUnicodeChar()
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