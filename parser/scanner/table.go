package scanner

import (
	"github.com/nukilabs/unicodeid"
	"github.com/t14raptor/go-fast/token"
	"unicode"
)

func (s *Scanner) handleByte(b byte) token.Token {
	switch b {
	// ---- Whitespace ----
	case '\t', ' ': // 0x09, 0x20
		s.ConsumeByte()
		return token.Skip

	case '\n', '\r': // 0x0A, 0x0D
		s.ConsumeByte()
		return s.handleLineBreak()

	case 0x0B, 0x0C: // VT, FF — irregular whitespace
		s.ConsumeByte()
		return token.Skip

	// ---- Single-character punctuation / delimiters ----
	case '(':
		s.ConsumeByte()
		return token.LeftParenthesis
	case ')':
		s.ConsumeByte()
		return token.RightParenthesis
	case ',':
		s.ConsumeByte()
		return token.Comma
	case ':':
		s.ConsumeByte()
		return token.Colon
	case ';':
		s.ConsumeByte()
		return token.Semicolon
	case '[':
		s.ConsumeByte()
		return token.LeftBracket
	case ']':
		s.ConsumeByte()
		return token.RightBracket
	case '{':
		s.ConsumeByte()
		return token.LeftBrace
	case '}':
		s.ConsumeByte()
		return token.RightBrace
	case '~':
		s.ConsumeByte()
		return token.BitwiseNot

	// ---- Operators / multi-character punctuation ----
	case '!':
		s.ConsumeByte()
		if s.AdvanceIfByteEquals('=') {
			if s.AdvanceIfByteEquals('=') {
				return token.StrictNotEqual
			}
			return token.NotEqual
		}
		return token.Not

	case '%':
		s.ConsumeByte()
		if s.AdvanceIfByteEquals('=') {
			return token.RemainderAssign
		}
		return token.Remainder

	case '&':
		s.ConsumeByte()
		if s.AdvanceIfByteEquals('&') {
			if s.AdvanceIfByteEquals('=') {
				return token.LogicalAndAssign
			}
			return token.LogicalAnd
		} else if s.AdvanceIfByteEquals('=') {
			return token.AndAssign
		}
		return token.And

	case '*':
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

	case '+':
		s.ConsumeByte()
		if s.AdvanceIfByteEquals('+') {
			return token.Increment
		} else if s.AdvanceIfByteEquals('=') {
			return token.AddAssign
		}
		return token.Plus

	case '-':
		s.ConsumeByte()
		if s.AdvanceIfByteEquals('-') {
			return token.Decrement
		} else if s.AdvanceIfByteEquals('=') {
			return token.SubtractAssign
		}
		return token.Minus

	case '.':
		s.ConsumeByte()
		return s.readDot()

	case '/':
		s.ConsumeByte()
		b2, ok := s.PeekByte()
		if ok {
			switch b2 {
			case '/':
				s.ConsumeByte()
				s.skipSingleLineComment()
				return token.Skip
			case '*':
				s.ConsumeByte()
				s.skipMultiLineComment()
				return token.Skip
			}
		}
		if s.AdvanceIfByteEquals('=') {
			return token.QuotientAssign
		}
		return token.Slash

	case '<':
		s.ConsumeByte()
		if s.AdvanceIfByteEquals('<') {
			if s.AdvanceIfByteEquals('=') {
				return token.ShiftLeftAssign
			}
			return token.ShiftLeft
		} else if s.AdvanceIfByteEquals('=') {
			return token.LessOrEqual
		}
		return token.Less

	case '=':
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

	case '>':
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

	case '?':
		s.ConsumeByte()
		next2Bytes, ok := s.src.PeekTwoBytes()
		if ok {
			switch next2Bytes[0] {
			case '?':
				if next2Bytes[1] == '=' {
					s.ConsumeByte()
					s.ConsumeByte()
					return token.CoalesceAssign
				}
				s.ConsumeByte()
				return token.Coalesce
			case '.':
				if !unicode.IsDigit(rune(next2Bytes[1])) {
					s.ConsumeByte()
					return token.QuestionDot
				}
			}
			return token.QuestionMark
		}
		nextByte, ok := s.PeekByte()
		if !ok {
			return token.QuestionMark
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

	case '^':
		s.ConsumeByte()
		if s.AdvanceIfByteEquals('=') {
			return token.ExclusiveOrAssign
		}
		return token.ExclusiveOr

	case '|':
		s.ConsumeByte()
		if s.AdvanceIfByteEquals('|') {
			if s.AdvanceIfByteEquals('=') {
				return token.LogicalOrAssign
			}
			return token.LogicalOr
		} else if s.AdvanceIfByteEquals('=') {
			return token.OrAssign
		}
		return token.Or

	// ---- String / template literals ----

	case '"':
		return s.scanStringLiteralDoubleQuote()

	case '\'':
		return s.scanStringLiteralSingleQuote()

	case '`':
		s.ConsumeByte()
		return s.ReadTemplateLiteral(token.TemplateHead, token.NoSubstitutionTemplate)

	// ---- Special ----

	case '#':
		// Check for shebang (#!) at the very start of the file
		if s.src.Offset() == 0 {
			afterHash := s.src.Offset() + 1
			if afterHash < s.src.EndOffset() && s.src.ReadPosition(afterHash) == '!' {
				s.ConsumeByte() // #
				s.ConsumeByte() // !
				s.skipSingleLineComment()
				return token.Skip
			}
		}
		s.scanIdentifierTail()
		return token.PrivateIdentifier

	case '\\':
		return s.identifierBackslashHandler()

	// ---- Numeric literals ----

	case '0':
		s.ConsumeByte()
		return s.readZero()

	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		s.ConsumeByte()
		return s.decimalLiteralAfterFirstDigit()

	// ---- Identifier start: keyword-leading lowercase letters ----

	case 'a':
		switch s.scanIdentifierTail() {
		case "await":
			return token.Await
		case "async":
			return token.Async
		}
		return token.Identifier

	case 'b':
		switch s.scanIdentifierTail() {
		case "break":
			return token.Break
		}
		return token.Identifier

	case 'c':
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

	case 'd':
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

	case 'e':
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

	case 'f':
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

	case 'i':
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

	case 'l':
		switch s.scanIdentifierTail() {
		case "let":
			return token.Let
		}
		return token.Identifier

	case 'n':
		switch s.scanIdentifierTail() {
		case "new":
			return token.New
		case "null":
			return token.Null
		}
		return token.Identifier

	case 'o':
		switch s.scanIdentifierTail() {
		case "of":
			return token.Of
		}
		return token.Identifier

	case 'r':
		switch s.scanIdentifierTail() {
		case "return":
			return token.Return
		}
		return token.Identifier

	case 's':
		switch s.scanIdentifierTail() {
		case "super":
			return token.Super
		case "static":
			return token.Static
		case "switch":
			return token.Switch
		}
		return token.Identifier

	case 't':
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

	case 'v':
		switch s.scanIdentifierTail() {
		case "var":
			return token.Var
		case "void":
			return token.Void
		}
		return token.Identifier

	case 'w':
		switch s.scanIdentifierTail() {
		case "while":
			return token.While
		case "with":
			return token.With
		}
		return token.Identifier

	case 'y':
		switch s.scanIdentifierTail() {
		case "yield":
			return token.Yield
		}
		return token.Identifier

	// ---- Identifier start: uppercase A-Z, non-keyword lowercase, _, $ ----

	case '$', '_',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
		'g', 'h', 'j', 'k', 'm', 'p', 'q', 'u', 'x', 'z':
		s.scanIdentifierTail()
		return token.Identifier

	// ---- Invalid ASCII ----

	case 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x0E, 0x0F,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
		0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F,
		'@', 0x7F:
		c := s.ConsumeRune()
		s.error(invalidCharacter(c, s.Token.Idx0, s.src.Offset()))
		return token.Undetermined

	// ---- Non-ASCII: valid UTF-8 leading bytes ----

	case 0xC2, 0xC3, 0xC4, 0xC5, 0xC6, 0xC7, 0xC8, 0xC9, 0xCA, 0xCB, 0xCC, 0xCD, 0xCE, 0xCF,
		0xD0, 0xD1, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7, 0xD8, 0xD9, 0xDA, 0xDB, 0xDC, 0xDD, 0xDE, 0xDF,
		0xE0, 0xE1, 0xE2, 0xE3, 0xE4, 0xE5, 0xE6, 0xE7, 0xE8, 0xE9, 0xEA, 0xEB, 0xEC, 0xED, 0xEE, 0xEF,
		0xF0, 0xF1, 0xF2, 0xF3, 0xF4:
		switch c, _ := s.PeekRune(); {
		case unicodeid.IsIDStartUnicode(c):
			s.scanIdentifierTailAfterUnicode(s.src.Offset())
			return token.Identifier
		case unicode.IsSpace(c):
			s.ConsumeRune()
			return token.Skip
		case isLineTerminator(c):
			s.ConsumeRune()
			s.Token.OnNewLine = true
			return token.Skip
		default:
			start := s.src.Offset()
			s.ConsumeRune()
			s.error(invalidCharacter(c, start, s.src.Offset()))
			return token.Undetermined
		}

	default:
		panic("unreachable")
	}
}
