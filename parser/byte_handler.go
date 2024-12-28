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
let c = p.consume_char();
p.error(diagnostics::invalid_character(c, p.unterminated_range())
return token.Undetermined
}

// <SPACE> <TAB> Normal Whitespace
func sps(p *parser) token.Token {
p.consume_char();
return token.Skip
}

// <VT> <FF> Irregular Whitespace
func isp(p *parser) token.Token {
p.consume_char();
p.trivia_builder.add_irregular_whitespace(p.token.start, p.offset());
	return token.Skip
}

// '\r' '\n'
func lin(p *parser) token.Token {
p.consume_char();
p.line_break_handler()
}

// !
func exl(p *parser) token.Token {
	p.consume_char();
	if p.next_ascii_byte_eq('=') {
		if p.next_ascii_byte_eq('=') {
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
unsafe { p.read_string_literal_single_quote() }
}

// #
func has(p *parser) token.Token {
p.consume_char();
// HashbangComment ::
//     `#!` SingleLineCommentChars?
if p.token.start == 0 && p.next_ascii_byte_eq('!') {
p.read_hashbang_comment()
} else {
p.private_identifier()
}
}

// `A..=Z`, `a..=z` (except special cases below), `_`, `$`
ascii_identifier_handler!(IDT(_id_without_first_char) {
token.Ident
}

// %
func prc(p *parser) token.Token {
	p.consume_char();
	if p.next_ascii_byte_eq('=') {
		return token.RemainderAssign
	}

	return token.Remainder
}

// &
func amp(p *parser) token.Token {
	p.consume_char();
	if p.next_ascii_byte_eq('&') {
		if p.next_ascii_byte_eq('=') {
			return token.LogicalAnd
		}
		return token.LogicalAnd
	} else if p.next_ascii_byte_eq('=') {
		return token.AndAssign
	}
	return token.And
}

// (
func pno(p *parser) token.Token {
p.consume_char();
return token.LeftParenthesis
}

// )
func pnc(p *parser) token.Token {
p.consume_char();
return token.RightParenthesis
}

// *
func atr(p *parser) token.Token {
	p.consume_char();
	if p.next_ascii_byte_eq('*') {
		if p.next_ascii_byte_eq('=') {
			return token.ExponentAssign
		}
		return token.Exponent
	} else if p.next_ascii_byte_eq('=') {
		return token.MultiplyAssign
	}
	return token.Multiply
}

// +
func pls(p *parser) token.Token {
	p.consume_char();
	if p.next_ascii_byte_eq('+') {
		return token.Increment
	} else if p.next_ascii_byte_eq('=') {
		return token.AddAssign
	}
	return token.Plus
}

// ,
func com(p *parser) token.Token {
p.consume_char();
return token.Comma
}

// -
func min(p *parser) token.Token {
p.consume_char();
p.read_minus().unwrap_or_else(|| p.skip_single_line_comment())
}

// .
func prd(p *parser) token.Token {
p.consume_char();
return token.Period
p.read_dot()
}

// /
func slh(p *parser) token.Token {
	p.consume_char();
	switch p.peek_byte() {
	case '/':
		p.consume_char();
		return p.skip_single_line_comment()
	case '*':
		p.consume_char();
		return p.skip_multi_line_comment()
	default:
		// regex is handled separately, see `next_regex`
		if p.next_ascii_byte_eq('=') {
			return token.QuotientAssign
		}
		return token.Slash
	}
}

// 0
func zer(p *parser) token.Token {
p.consume_char();
return p.read_zero()
}

// 1 to 9
func dig(p *parser) token.Token {
p.consume_char();
return p.decimal_literal_after_first_digit()
}

// :
func col(p *parser) token.Token {
p.consume_char();
return token.Colon
}

// ;
func sem(p *parser) token.Token {
p.consume_char();
return token.Semicolon
}

// <
func lss(p *parser) token.Token {
p.consume_char();
return p.read_left_angle().unwrap_or_else(|| p.skip_single_line_comment())
}

// =
func eql(p *parser) token.Token {
	p.consume_char();
	if p.next_ascii_byte_eq('=') {
		if p.next_ascii_byte_eq('=') {
			return token.StrictEqual
		}
		return token.Equal
	} else if p.next_ascii_byte_eq('>') {
		return token.Arrow
	}
	return token.Assign
}

// >
func gtr(p *parser) token.Token {
p.consume_char();
// `>=` is re-lexed with [Lexer::next_jsx_child]
return token.Greater
}

// ?
func qst(p *parser) token.Token {
p.consume_char();

if let Some(next_2_bytes) = p.peek_2_bytes() {
switch next_2_bytes[0] {
case '?':
if next_2_bytes[1] == '=' {
p.consume_2_chars();
token.Question2Eq
} else {
p.consume_char();
token.Question2
}
// parse `?.1` as `?` `.1`
'.' if !next_2_bytes[1].is_ascii_digit() => {
p.consume_char();
token.QuestionDot
}
_ => token.Question,
}
} else {
// At EOF, or only 1 byte left
match p.peek_byte() {
Some('?') => {
p.consume_char();
token.Question2
}
Some('.') => {
p.consume_char();
token.QuestionDot
}
_ => token.Question,
}
}
}

// @
func at_(p *parser) token.Token {
p.consume_char();
token.At
}

// [
func bto(p *parser) token.Token {
p.consume_char();
	return token.LeftBracket
}

// \
func esc(p *parser) token.Token {
p.identifier_backslash_handler()
}

// ]
func btc(p *parser) token.Token {
p.consume_char();
return token.RightBracket
}

// ^
func crt(p *parser) token.Token {
	p.consume_char();
	if p.next_ascii_byte_eq('=') {
		return token.ExclusiveOrAssign
	}
	return token.ExclusiveOr
}

// `
func tpl(p *parser) token.Token {
p.consume_char();
p.read_template_literal(token.TemplateHead, token.NoSubstitutionTemplate)
}

// {
func beo(p *parser) token.Token {
p.consume_char();
	return token.LeftBrace
}

// |
func pip(p *parser) token.Token {
p.consume_char();

match p.peek_byte() {
Some('|') => {
p.consume_char();
if p.next_ascii_byte_eq('=') {
token.Pipe2Eq
} else {
token.Pipe2
}
}
Some('=') => {
p.consume_char();
token.PipeEq
}
_ => token.Pipe
}
}

// }
func bec(p *parser) token.Token {
p.consume_char();
return token.RightBrace
}

// ~
func tld(p *parser) token.Token {
p.consume_char();
return token.BitwiseNot
}

ascii_identifier_handler!(L_A(id_without_first_char) match id_without_first_char {
"wait" => token.Await,
"sync" => token.Async,
"bstract" => token.Abstract,
"ccessor" => token.Accessor,
"ny" => token.Any,
"s" => token.As,
"ssert" => token.Assert,
"sserts" => token.Asserts,
_ => token.Ident,
}

ascii_identifier_handler!(L_B(id_without_first_char) match id_without_first_char {
"reak" => token.Break,
"oolean" => token.Boolean,
"igint" => token.BigInt,
_ => token.Ident,
}

ascii_identifier_handler!(L_C(id_without_first_char) match id_without_first_char {
"onst" => token.Const,
"lass" => token.Class,
"ontinue" => token.Continue,
"atch" => token.Catch,
"ase" => token.Case,
"onstructor" => token.Constructor,
_ => token.Ident,
}

ascii_identifier_handler!(L_D(id_without_first_char) match id_without_first_char {
"o" => token.Do,
"elete" => token.Delete,
"eclare" => token.Declare,
"efault" => token.Default,
"ebugger" => token.Debugger,
"efer" => token.Defer,
_ => token.Ident,
}

ascii_identifier_handler!(L_E(id_without_first_char) match id_without_first_char {
"lse" => token.Else,
"num" => token.Enum,
"xport" => token.Export,
"xtends" => token.Extends,
_ => token.Ident,
}

ascii_identifier_handler!(L_F(id_without_first_char) match id_without_first_char {
"unction" => token.Function,
"alse" => token.False,
"or" => token.For,
"inally" => token.Finally,
"rom" => token.From,
_ => token.Ident,
}

ascii_identifier_handler!(L_G(id_without_first_char) match id_without_first_char {
"et" => token.Get,
"lobal" => token.Global,
_ => token.Ident,
}

ascii_identifier_handler!(L_I(id_without_first_char) match id_without_first_char {
"f" => token.If,
"nstanceof" => token.Instanceof,
"n" => token.In,
"mplements" => token.Implements,
"mport" => token.Import,
"nfer" => token.Infer,
"nterface" => token.Interface,
"ntrinsic" => token.Intrinsic,
"s" => token.Is,
_ => token.Ident,
}

ascii_identifier_handler!(L_K(id_without_first_char) match id_without_first_char {
"eyof" => token.KeyOf,
_ => token.Ident,
}

ascii_identifier_handler!(L_L(id_without_first_char) match id_without_first_char {
"et" => token.Let,
_ => token.Ident,
}

ascii_identifier_handler!(L_M(id_without_first_char) match id_without_first_char {
"eta" => token.Meta,
"odule" => token.Module,
_ => token.Ident,
}

ascii_identifier_handler!(L_N(id_without_first_char) match id_without_first_char {
"ull" => token.Null,
"ew" => token.New,
"umber" => token.Number,
"amespace" => token.Namespace,
"ever" => token.Never,
_ => token.Ident,
}

ascii_identifier_handler!(L_O(id_without_first_char) match id_without_first_char {
"f" => token.Of,
"bject" => token.Object,
"ut" => token.Out,
"verride" => token.Override,
_ => token.Ident,
}

ascii_identifier_handler!(L_P(id_without_first_char) match id_without_first_char {
"ackage" => token.Package,
"rivate" => token.Private,
"rotected" => token.Protected,
"ublic" => token.Public,
_ => token.Ident,
}

ascii_identifier_handler!(L_R(id_without_first_char) match id_without_first_char {
"eturn" => token.Return,
"equire" => token.Require,
"eadonly" => token.Readonly,
_ => token.Ident,
}

ascii_identifier_handler!(L_S(id_without_first_char) match id_without_first_char {
"et" => token.Set,
"uper" => token.Super,
"witch" => token.Switch,
"tatic" => token.Static,
"ymbol" => token.Symbol,
"tring" => token.String,
"atisfies" => token.Satisfies,
"ource" => token.Source,
_ => token.Ident,
}

ascii_identifier_handler!(L_T(id_without_first_char) match id_without_first_char {
"his" => token.This,
"rue" => token.True,
"hrow" => token.Throw,
"ry" => token.Try,
"ypeof" => token.Typeof,
"arget" => token.Target,
"ype" => token.Type,
_ => token.Ident,
}

ascii_identifier_handler!(L_U(id_without_first_char) match id_without_first_char {
"ndefined" => token.Undefined,
"sing" => token.Using,
"nique" => token.Unique,
"nknown" => token.Unknown,
_ => token.Ident,
}

ascii_identifier_handler!(L_V(id_without_first_char) match id_without_first_char {
"ar" => token.Var,
"oid" => token.Void,
_ => token.Ident,
}

ascii_identifier_handler!(L_W(id_without_first_char) match id_without_first_char {
"hile" => token.While,
"ith" => token.With,
_ => token.Ident,
}

ascii_identifier_handler!(L_Y(id_without_first_char) match id_without_first_char {
"ield" => token.Yield,
_ => token.Ident,
}

// Non-ASCII characters.
// NB: Must not use `ascii_byte_handler!` macro, as this handler is for non-ASCII chars.
byte_handler!(UNI(p) {
p.unicode_char_handler()
}

// UTF-8 continuation bytes (0x80 - 0xBF) (i.e. middle of a multi-byte UTF-8 sequence)
// + and byte values which are not legal in UTF-8 strings (0xC0, 0xC1, 0xF5 - 0xFF).
// `handle_byte()` should only be called with 1st byte of a valid UTF-8 character,
// so something has gone wrong if we get here.
// https://datatracker.ietf.org/doc/html/rfc3629
// NB: Must not use `ascii_byte_handler!` macro, as this handler is for non-ASCII bytes.
byte_handler!(UER(_p) {
unreachable!();
}
