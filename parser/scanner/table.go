package scanner

import (
	"github.com/nukilabs/unicodeid"
	"github.com/t14raptor/go-fast/token"
	"unicode"
)

func (s *Scanner) handleByte(c byte) token.Token {
	switch c {
	case 0, 1, 2, 3, 4, 5, 6, 7, 8, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23,
		24, 25, 26, 27, 28, 29, 30, 31, 64: // `\0` `\1` etc
		s.ConsumeByte()
		// p.errorUnexpected(p.chr) TODO
		return token.Undetermined
	case 9, 32: // <SPACE> <TAB> Normal Whitespace
		s.ConsumeByte()
		return token.Skip
	case 10, 13: // '\r' '\n'
		s.ConsumeByte()
		return s.handleLineBreak()
	case 11, 12: // <VT> <FF> Irregular Whitespace
		s.ConsumeByte()
		return token.Skip
	case 33: // '!'
		s.ConsumeByte()
		if s.AdvanceIfByteEquals('=') {
			if s.AdvanceIfByteEquals('=') {
				return token.StrictNotEqual
			}
			return token.NotEqual
		}
		return token.Not
	case 34: // '"'
		return s.scanStringLiteralDoubleQuote()
	case 35: // '#'
		// possible shebang or private identifier
		s.scanIdentifierTail()
		return token.PrivateIdentifier
	case 36: // '$'
		s.scanIdentifierTail()
		return token.Identifier
	case 37: // '%'
		s.ConsumeByte()
		if s.AdvanceIfByteEquals('=') {
			return token.RemainderAssign
		}
		return token.Remainder
	case 38: // '&'
		s.ConsumeByte()
		if s.AdvanceIfByteEquals('&') {
			// Could also check if next is '='
			if s.AdvanceIfByteEquals('=') {
				// e.g. token.LogicalAndAssign if you had that
				// but your code does “TODO” so we’ll just return token.LogicalAnd
			}
			return token.LogicalAnd
		} else if s.AdvanceIfByteEquals('=') {
			return token.AndAssign
		}
		return token.And
	case 39: // "'"
		return s.scanStringLiteralSingleQuote()
	case 40: // '('
		s.ConsumeByte()
		return token.LeftParenthesis
	case 41: // ')'
		s.ConsumeByte()
		return token.RightParenthesis
	case 42: // '*'
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
	case 43: // '+'
		s.ConsumeByte()
		if s.AdvanceIfByteEquals('+') {
			return token.Increment
		} else if s.AdvanceIfByteEquals('=') {
			return token.AddAssign
		}
		return token.Plus
	case 44: // ','
		s.ConsumeByte()
		return token.Comma
	case 45: // '-'
		s.ConsumeByte()
		if s.AdvanceIfByteEquals('-') {
			return token.Decrement
		} else if s.AdvanceIfByteEquals('=') {
			return token.SubtractAssign
		}
		return token.Minus
	case 46: // '.'
		s.ConsumeByte()
		return s.readDot()
	case 47: // '/'
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
		if s.AdvanceIfByteEquals('=') {
			return token.QuotientAssign
		}
		return token.Slash
	case 48: // '0'
		s.ConsumeByte()
		return s.readZero()
	case 49, 50, 51, 52, 53, 54, 55, 56, 57: // '1'..'9'
		s.ConsumeByte()
		return s.decimalLiteralAfterFirstDigit()
	case 58: // ':'
		s.ConsumeByte()
		return token.Colon
	case 59: // ';'
		s.ConsumeByte()
		return token.Semicolon
	case 60: // '<'
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
	case 61: // '='
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
	case 62: // '>'
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
	case 63: // '?'
		s.ConsumeByte()
		next2, ok := s.src.PeekTwoBytes()
		if ok {
			switch next2[0] {
			case '?':
				if next2[1] == '=' {
					s.ConsumeByte()
					s.ConsumeByte()
					return token.Coalesce
				}
				s.ConsumeByte()
				return token.Coalesce
			case '.':
				// `?.`
				if !unicode.IsDigit(rune(next2[1])) {
					s.ConsumeByte()
					return token.QuestionDot
				}
			}
			return token.QuestionMark
		}
		// Only 1 byte left or EOF
		nb, ok := s.PeekByte()
		if !ok {
			return token.Eof
		}
		switch nb {
		case '?':
			s.ConsumeByte()
			return token.Coalesce
		case '.':
			s.ConsumeByte()
			return token.QuestionDot
		}
		return token.QuestionMark
	case 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 95, 103, 104, 106, 107, 109, 112, 113, 117, 120: // `A..=Z`, `a..=z`
		s.scanIdentifierTail()
		return token.Identifier
	case 91: // '['
		s.ConsumeByte()
		return token.LeftBracket
	case 92: // '\\'
		return s.identifierBackslashHandler()
	case 93: // ']'
		s.ConsumeByte()
		return token.RightBracket
	case 94: // '^'
		s.ConsumeByte()
		if s.AdvanceIfByteEquals('=') {
			return token.ExclusiveOrAssign
		}
		return token.ExclusiveOr
	case 96: // '`'
		s.ConsumeByte()
		return token.Backtick
	case 97: // 'a'
		switch s.scanIdentifierTail() {
		case "await":
			return token.Await
		case "async":
			return token.Async
		}
		return token.Identifier
	case 98: // 'b'
		switch s.scanIdentifierTail() {
		case "break":
			return token.Break
		case "boolean":
			return token.Boolean
		}
		return token.Identifier
	case 99: // 'c'
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
	case 100: // 'd'
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
	case 101: // 'e'
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
	case 102: // 'f'
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
	case 105: // 'i'
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
	case 108: // 'l'
		switch s.scanIdentifierTail() {
		case "let":
			return token.Let
		}
		return token.Identifier
	case 110: // 'n'
		switch s.scanIdentifierTail() {
		case "new":
			return token.New
		case "null":
			return token.Null
		case "number":
			return token.Number
		}
		return token.Identifier
	case 111: // 'o'
		switch s.scanIdentifierTail() {
		case "of":
			return token.Of
		}
		return token.Identifier
	case 114: // 'r'
		switch s.scanIdentifierTail() {
		case "return":
			return token.Return
		}
		return token.Identifier
	case 115: // 's'
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
	case 116: // 't'
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
	case 118: // 'v'
		switch s.scanIdentifierTail() {
		case "var":
			return token.Var
		case "void":
			return token.Void
		}
		return token.Identifier
	case 119: // 'w'
		switch s.scanIdentifierTail() {
		case "while":
			return token.While
		case "with":
			return token.With
		}
		return token.Identifier
	case 121: // 'y'
		switch s.scanIdentifierTail() {
		case "yield":
			return token.Yield
		}
		return token.Identifier
	case 122: // 'z'
		s.scanIdentifierTail()
		return token.Identifier
	case 123: // '{'
		s.ConsumeByte()
		return token.LeftBrace
	case 124: // '|'
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
	case 125: // '}'
		s.ConsumeByte()
		return token.RightBrace
	case 126: // '~'
		s.ConsumeByte()
		return token.BitwiseNot
	case 127:
		// err (DEL)
		s.ConsumeByte()
		return token.Undetermined
	case 128, 129, 130, 131, 132, 133, 134, 135,
		136, 137, 138, 139, 140, 141, 142, 143,
		144, 145, 146, 147, 148, 149, 150, 151,
		152, 153, 154, 155, 156, 157, 158, 159,
		160, 161, 162, 163, 164, 165, 166, 167,
		168, 169, 170, 171, 172, 173, 174, 175,
		176, 177, 178, 179, 180, 181, 182, 183,
		184, 185, 186, 187, 188, 189, 190, 191:
		panic("unreachable (invalid UTF-8 continuation or unexpected high-bit byte)")
	case 192, 193, 194, 195, 196, 197, 198, 199,
		200, 201, 202, 203, 204, 205, 206, 207,
		208, 209, 210, 211, 212, 213, 214, 215,
		216, 217, 218, 219, 220, 221, 222, 223,
		224, 225, 226, 227, 228, 229, 230, 231,
		232, 233, 234, 235, 236, 237, 238, 239:
		switch chr, _ := s.PeekRune(); {
		case unicodeid.IsIDStartUnicode(chr):
			s.scanIdentifierTailAfterUnicode(s.src.Offset())
			return token.Identifier
		case unicode.IsSpace(chr):
			s.ConsumeRune()
			return token.Skip
		case isLineTerminator(chr):
			s.ConsumeRune()
			s.token.OnNewLine = true
			return token.Skip
		default:
			s.ConsumeRune()
			// p.errorUnexpected(chr) TODO
			return token.Undetermined
		}
	case 240, 241, 242, 243, 244, 245, 246, 247,
		248, 249, 250, 251, 252, 253, 254, 255:
		panic("unreachable")
	}
	panic("unreachable")
}
