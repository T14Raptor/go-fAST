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
let c = lexer.consume_char();
lexer.error(diagnostics::invalid_character(c, lexer.unterminated_range())
return Kind::Undetermined
}

// <SPACE> <TAB> Normal Whitespace
func sps(p *parser) token.Token {
lexer.consume_char();
return token.Skip
}

// <VT> <FF> Irregular Whitespace
func isp(p *parser) token.Token {
lexer.consume_char();
lexer.trivia_builder.add_irregular_whitespace(lexer.token.start, lexer.offset());
	return token.Skip
}

// '\r' '\n'
func lin(p *parser) token.Token {
lexer.consume_char();
lexer.line_break_handler()
}

// !
func exl(p *parser) token.Token {
lexer.consume_char();
if lexer.next_ascii_byte_eq(b'=') {
if lexer.next_ascii_byte_eq(b'=') {
Kind::Neq2
} else {
Kind::Neq
}
} else {
Kind::Bang
}
}

// "
func qod(p *parser) token.Token {
// SAFETY: This function is only called for `"`
unsafe { lexer.read_string_literal_double_quote() }
}

// '
func qos(p *parser) token.Token {
// SAFETY: This function is only called for `'`
unsafe { lexer.read_string_literal_single_quote() }
}

// #
func has(p *parser) token.Token {
lexer.consume_char();
// HashbangComment ::
//     `#!` SingleLineCommentChars?
if lexer.token.start == 0 && lexer.next_ascii_byte_eq(b'!') {
lexer.read_hashbang_comment()
} else {
lexer.private_identifier()
}
}

// `A..=Z`, `a..=z` (except special cases below), `_`, `$`
ascii_identifier_handler!(IDT(_id_without_first_char) {
Kind::Ident
}

// %
func prc(p *parser) token.Token {
	lexer.consume_char();
	if lexer.next_ascii_byte_eq('=') {
		return token.RemainderAssign
	}

	return token.Remainder
}

// &
func amp(p *parser) token.Token {
	lexer.consume_char();
	if lexer.next_ascii_byte_eq('&') {
		if lexer.next_ascii_byte_eq('=') {
			return token.LogicalAnd
		} else {
			return token.LogicalAnd
		}
	} else if lexer.next_ascii_byte_eq('=') {
		return token.AndAssign
	} else {
		return token.And
	}
}

// (
func pno(p *parser) token.Token {
lexer.consume_char();
return token.LeftParenthesis
}

// )
func pnc(p *parser) token.Token {
lexer.consume_char();
return token.RightParenthesis
}

// *
func atr(p *parser) token.Token {
lexer.consume_char();
if lexer.next_ascii_byte_eq(b'*') {
if lexer.next_ascii_byte_eq(b'=') {
Kind::Star2Eq
} else {
Kind::Star2
}
} else if lexer.next_ascii_byte_eq(b'=') {
Kind::StarEq
} else {
Kind::Star
}
}

// +
func pls(p *parser) token.Token {
lexer.consume_char();
if lexer.next_ascii_byte_eq(b'+') {
Kind::Plus2
} else if lexer.next_ascii_byte_eq(b'=') {
Kind::PlusEq
} else {
Kind::Plus
}
}

// ,
func com(p *parser) token.Token {
lexer.consume_char();
return token.Comma
}

// -
func min(p *parser) token.Token {
lexer.consume_char();
lexer.read_minus().unwrap_or_else(|| lexer.skip_single_line_comment())
}

// .
func prd(p *parser) token.Token {
lexer.consume_char();
return token.Period
lexer.read_dot()
}

// /
func slh(p *parser) token.Token {
lexer.consume_char();
match lexer.peek_byte() {
Some(b'/') => {
lexer.consume_char();
lexer.skip_single_line_comment()
}
Some(b'*') => {
lexer.consume_char();
lexer.skip_multi_line_comment()
}
_ => {
// regex is handled separately, see `next_regex`
if lexer.next_ascii_byte_eq(b'=') {
Kind::SlashEq
} else {
Kind::Slash
}
}
}
}

// 0
func zer(p *parser) token.Token {
lexer.consume_char();
lexer.read_zero()
}

// 1 to 9
func dig(p *parser) token.Token {
lexer.consume_char();
lexer.decimal_literal_after_first_digit()
}

// :
func col(p *parser) token.Token {
lexer.consume_char();
return token.Colon
}

// ;
func sem(p *parser) token.Token {
lexer.consume_char();
return token.Semicolon
}

// <
func lss(p *parser) token.Token {
lexer.consume_char();
lexer.read_left_angle().unwrap_or_else(|| lexer.skip_single_line_comment())
}

// =
func eql(p *parser) token.Token {
lexer.consume_char();
if lexer.next_ascii_byte_eq(b'=') {
if lexer.next_ascii_byte_eq(b'=') {
Kind::Eq3
} else {
Kind::Eq2
}
} else if lexer.next_ascii_byte_eq(b'>') {
Kind::Arrow
} else {
Kind::Eq
}
}

// >
func gtr(p *parser) token.Token {
lexer.consume_char();
// `>=` is re-lexed with [Lexer::next_jsx_child]
return token.Greater
}

// ?
func qst(p *parser) token.Token {
lexer.consume_char();

if let Some(next_2_bytes) = lexer.peek_2_bytes() {
match next_2_bytes[0] {
b'?' => {
if next_2_bytes[1] == b'=' {
lexer.consume_2_chars();
Kind::Question2Eq
} else {
lexer.consume_char();
Kind::Question2
}
}
// parse `?.1` as `?` `.1`
b'.' if !next_2_bytes[1].is_ascii_digit() => {
lexer.consume_char();
Kind::QuestionDot
}
_ => Kind::Question,
}
} else {
// At EOF, or only 1 byte left
match lexer.peek_byte() {
Some(b'?') => {
lexer.consume_char();
Kind::Question2
}
Some(b'.') => {
lexer.consume_char();
Kind::QuestionDot
}
_ => Kind::Question,
}
}
}

// @
func at_(p *parser) token.Token {
lexer.consume_char();
Kind::At
}

// [
func bto(p *parser) token.Token {
lexer.consume_char();
	return token.LeftBracket
}

// \
func esc(p *parser) token.Token {
lexer.identifier_backslash_handler()
}

// ]
func btc(p *parser) token.Token {
lexer.consume_char();
return token.RightBracket
}

// ^
func crt(p *parser) token.Token {
lexer.consume_char();
if lexer.next_ascii_byte_eq(b'=') {
Kind::CaretEq
} else {
Kind::Caret
}
}

// `
func tpl(p *parser) token.Token {
lexer.consume_char();
lexer.read_template_literal(Kind::TemplateHead, Kind::NoSubstitutionTemplate)
}

// {
func beo(p *parser) token.Token {
lexer.consume_char();
	return token.LeftBrace
}

// |
func pip(p *parser) token.Token {
lexer.consume_char();

match lexer.peek_byte() {
Some(b'|') => {
lexer.consume_char();
if lexer.next_ascii_byte_eq(b'=') {
Kind::Pipe2Eq
} else {
Kind::Pipe2
}
}
Some(b'=') => {
lexer.consume_char();
Kind::PipeEq
}
_ => Kind::Pipe
}
}

// }
func bec(p *parser) token.Token {
lexer.consume_char();
return token.RightBrace
}

// ~
func tld(p *parser) token.Token {
lexer.consume_char();
return token.BitwiseNot
}

ascii_identifier_handler!(L_A(id_without_first_char) match id_without_first_char {
"wait" => Kind::Await,
"sync" => Kind::Async,
"bstract" => Kind::Abstract,
"ccessor" => Kind::Accessor,
"ny" => Kind::Any,
"s" => Kind::As,
"ssert" => Kind::Assert,
"sserts" => Kind::Asserts,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_B(id_without_first_char) match id_without_first_char {
"reak" => Kind::Break,
"oolean" => Kind::Boolean,
"igint" => Kind::BigInt,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_C(id_without_first_char) match id_without_first_char {
"onst" => Kind::Const,
"lass" => Kind::Class,
"ontinue" => Kind::Continue,
"atch" => Kind::Catch,
"ase" => Kind::Case,
"onstructor" => Kind::Constructor,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_D(id_without_first_char) match id_without_first_char {
"o" => Kind::Do,
"elete" => Kind::Delete,
"eclare" => Kind::Declare,
"efault" => Kind::Default,
"ebugger" => Kind::Debugger,
"efer" => Kind::Defer,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_E(id_without_first_char) match id_without_first_char {
"lse" => Kind::Else,
"num" => Kind::Enum,
"xport" => Kind::Export,
"xtends" => Kind::Extends,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_F(id_without_first_char) match id_without_first_char {
"unction" => Kind::Function,
"alse" => Kind::False,
"or" => Kind::For,
"inally" => Kind::Finally,
"rom" => Kind::From,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_G(id_without_first_char) match id_without_first_char {
"et" => Kind::Get,
"lobal" => Kind::Global,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_I(id_without_first_char) match id_without_first_char {
"f" => Kind::If,
"nstanceof" => Kind::Instanceof,
"n" => Kind::In,
"mplements" => Kind::Implements,
"mport" => Kind::Import,
"nfer" => Kind::Infer,
"nterface" => Kind::Interface,
"ntrinsic" => Kind::Intrinsic,
"s" => Kind::Is,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_K(id_without_first_char) match id_without_first_char {
"eyof" => Kind::KeyOf,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_L(id_without_first_char) match id_without_first_char {
"et" => Kind::Let,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_M(id_without_first_char) match id_without_first_char {
"eta" => Kind::Meta,
"odule" => Kind::Module,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_N(id_without_first_char) match id_without_first_char {
"ull" => Kind::Null,
"ew" => Kind::New,
"umber" => Kind::Number,
"amespace" => Kind::Namespace,
"ever" => Kind::Never,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_O(id_without_first_char) match id_without_first_char {
"f" => Kind::Of,
"bject" => Kind::Object,
"ut" => Kind::Out,
"verride" => Kind::Override,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_P(id_without_first_char) match id_without_first_char {
"ackage" => Kind::Package,
"rivate" => Kind::Private,
"rotected" => Kind::Protected,
"ublic" => Kind::Public,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_R(id_without_first_char) match id_without_first_char {
"eturn" => Kind::Return,
"equire" => Kind::Require,
"eadonly" => Kind::Readonly,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_S(id_without_first_char) match id_without_first_char {
"et" => Kind::Set,
"uper" => Kind::Super,
"witch" => Kind::Switch,
"tatic" => Kind::Static,
"ymbol" => Kind::Symbol,
"tring" => Kind::String,
"atisfies" => Kind::Satisfies,
"ource" => Kind::Source,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_T(id_without_first_char) match id_without_first_char {
"his" => Kind::This,
"rue" => Kind::True,
"hrow" => Kind::Throw,
"ry" => Kind::Try,
"ypeof" => Kind::Typeof,
"arget" => Kind::Target,
"ype" => Kind::Type,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_U(id_without_first_char) match id_without_first_char {
"ndefined" => Kind::Undefined,
"sing" => Kind::Using,
"nique" => Kind::Unique,
"nknown" => Kind::Unknown,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_V(id_without_first_char) match id_without_first_char {
"ar" => Kind::Var,
"oid" => Kind::Void,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_W(id_without_first_char) match id_without_first_char {
"hile" => Kind::While,
"ith" => Kind::With,
_ => Kind::Ident,
}

ascii_identifier_handler!(L_Y(id_without_first_char) match id_without_first_char {
"ield" => Kind::Yield,
_ => Kind::Ident,
}

// Non-ASCII characters.
// NB: Must not use `ascii_byte_handler!` macro, as this handler is for non-ASCII chars.
byte_handler!(UNI(lexer) {
lexer.unicode_char_handler()
}

// UTF-8 continuation bytes (0x80 - 0xBF) (i.e. middle of a multi-byte UTF-8 sequence)
// + and byte values which are not legal in UTF-8 strings (0xC0, 0xC1, 0xF5 - 0xFF).
// `handle_byte()` should only be called with 1st byte of a valid UTF-8 character,
// so something has gone wrong if we get here.
// https://datatracker.ietf.org/doc/html/rfc3629
// NB: Must not use `ascii_byte_handler!` macro, as this handler is for non-ASCII bytes.
byte_handler!(UER(_lexer) {
unreachable!();
}
