package scanner

import (
	"github.com/nukilabs/unicodeid"
	"github.com/t14raptor/go-fast/token"
	"unicode"
	"unsafe"
)

func (s *Scanner) Next() {
	s.Token.HasEscape = false
	s.Token.OnNewLine = false

	for {
		s.Token.Idx0 = s.src.pos

		if s.src.pos >= s.src.len {
			s.Token.Kind = token.Eof
			break
		}

		b := *(*byte)(unsafe.Add(s.src.base, s.src.pos))

		switch b {
		// ---- Whitespace ----
		case '\t', ' ':
			pos := s.src.pos + 1
			base := s.src.base
			end := s.src.len
			for pos < end {
				c := *(*byte)(unsafe.Add(base, pos))
				if c != ' ' && c != '\t' {
					break
				}
				pos++
			}
			s.src.pos = pos
			continue

		case '\n', '\r':
			s.ConsumeByte()
			s.handleLineBreak()
			continue

		case 0x0B, 0x0C:
			s.ConsumeByte()
			continue

		// ---- Single-character punctuation / delimiters ----
		case '(':
			s.ConsumeByte()
			s.Token.Kind = token.LeftParenthesis
		case ')':
			s.ConsumeByte()
			s.Token.Kind = token.RightParenthesis
		case ',':
			s.ConsumeByte()
			s.Token.Kind = token.Comma
		case ':':
			s.ConsumeByte()
			s.Token.Kind = token.Colon
		case ';':
			s.ConsumeByte()
			s.Token.Kind = token.Semicolon
		case '[':
			s.ConsumeByte()
			s.Token.Kind = token.LeftBracket
		case ']':
			s.ConsumeByte()
			s.Token.Kind = token.RightBracket
		case '{':
			s.ConsumeByte()
			s.Token.Kind = token.LeftBrace
		case '}':
			s.ConsumeByte()
			s.Token.Kind = token.RightBrace
		case '~':
			s.ConsumeByte()
			s.Token.Kind = token.BitwiseNot

		// ---- Operators / multi-character punctuation ----
		case '!':
			s.ConsumeByte()
			if s.AdvanceIfByteEquals('=') {
				if s.AdvanceIfByteEquals('=') {
					s.Token.Kind = token.StrictNotEqual
				} else {
					s.Token.Kind = token.NotEqual
				}
			} else {
				s.Token.Kind = token.Not
			}

		case '%':
			s.ConsumeByte()
			if s.AdvanceIfByteEquals('=') {
				s.Token.Kind = token.RemainderAssign
			} else {
				s.Token.Kind = token.Remainder
			}

		case '&':
			s.ConsumeByte()
			if s.AdvanceIfByteEquals('&') {
				if s.AdvanceIfByteEquals('=') {
					s.Token.Kind = token.LogicalAndAssign
				} else {
					s.Token.Kind = token.LogicalAnd
				}
			} else if s.AdvanceIfByteEquals('=') {
				s.Token.Kind = token.AndAssign
			} else {
				s.Token.Kind = token.And
			}

		case '*':
			s.ConsumeByte()
			if s.AdvanceIfByteEquals('*') {
				if s.AdvanceIfByteEquals('=') {
					s.Token.Kind = token.ExponentAssign
				} else {
					s.Token.Kind = token.Exponent
				}
			} else if s.AdvanceIfByteEquals('=') {
				s.Token.Kind = token.MultiplyAssign
			} else {
				s.Token.Kind = token.Multiply
			}

		case '+':
			s.ConsumeByte()
			if s.AdvanceIfByteEquals('+') {
				s.Token.Kind = token.Increment
			} else if s.AdvanceIfByteEquals('=') {
				s.Token.Kind = token.AddAssign
			} else {
				s.Token.Kind = token.Plus
			}

		case '-':
			s.ConsumeByte()
			if s.AdvanceIfByteEquals('-') {
				s.Token.Kind = token.Decrement
			} else if s.AdvanceIfByteEquals('=') {
				s.Token.Kind = token.SubtractAssign
			} else {
				s.Token.Kind = token.Minus
			}

		case '.':
			s.ConsumeByte()
			s.Token.Kind = s.readDot()

		case '/':
			s.ConsumeByte()
			b2, ok := s.PeekByte()
			if ok {
				switch b2 {
				case '/':
					s.ConsumeByte()
					s.skipSingleLineComment()
					continue
				case '*':
					s.ConsumeByte()
					s.skipMultiLineComment()
					continue
				}
			}
			if s.AdvanceIfByteEquals('=') {
				s.Token.Kind = token.QuotientAssign
			} else {
				s.Token.Kind = token.Slash
			}

		case '<':
			s.ConsumeByte()
			if s.AdvanceIfByteEquals('<') {
				if s.AdvanceIfByteEquals('=') {
					s.Token.Kind = token.ShiftLeftAssign
				} else {
					s.Token.Kind = token.ShiftLeft
				}
			} else if s.AdvanceIfByteEquals('=') {
				s.Token.Kind = token.LessOrEqual
			} else {
				s.Token.Kind = token.Less
			}

		case '=':
			s.ConsumeByte()
			if s.AdvanceIfByteEquals('=') {
				if s.AdvanceIfByteEquals('=') {
					s.Token.Kind = token.StrictEqual
				} else {
					s.Token.Kind = token.Equal
				}
			} else if s.AdvanceIfByteEquals('>') {
				s.Token.Kind = token.Arrow
			} else {
				s.Token.Kind = token.Assign
			}

		case '>':
			s.ConsumeByte()
			if s.AdvanceIfByteEquals('=') {
				s.Token.Kind = token.GreaterOrEqual
			} else if s.AdvanceIfByteEquals('>') {
				if s.AdvanceIfByteEquals('=') {
					s.Token.Kind = token.ShiftRightAssign
				} else if s.AdvanceIfByteEquals('>') {
					if s.AdvanceIfByteEquals('=') {
						s.Token.Kind = token.UnsignedShiftRightAssign
					} else {
						s.Token.Kind = token.UnsignedShiftRight
					}
				} else {
					s.Token.Kind = token.ShiftRight
				}
			} else {
				s.Token.Kind = token.Greater
			}

		case '?':
			s.ConsumeByte()
			next2Bytes, ok := s.src.PeekTwoBytes()
			if ok {
				switch next2Bytes[0] {
				case '?':
					if next2Bytes[1] == '=' {
						s.ConsumeByte()
						s.ConsumeByte()
						s.Token.Kind = token.CoalesceAssign
					} else {
						s.ConsumeByte()
						s.Token.Kind = token.Coalesce
					}
				case '.':
					if next2Bytes[1] < '0' || next2Bytes[1] > '9' {
						s.ConsumeByte()
						s.Token.Kind = token.QuestionDot
					} else {
						s.Token.Kind = token.QuestionMark
					}
				default:
					s.Token.Kind = token.QuestionMark
				}
			} else {
				nextByte, ok := s.PeekByte()
				if !ok {
					s.Token.Kind = token.QuestionMark
				} else {
					switch nextByte {
					case '?':
						s.ConsumeByte()
						s.Token.Kind = token.Coalesce
					case '.':
						s.ConsumeByte()
						s.Token.Kind = token.QuestionDot
					default:
						s.Token.Kind = token.QuestionMark
					}
				}
			}

		case '^':
			s.ConsumeByte()
			if s.AdvanceIfByteEquals('=') {
				s.Token.Kind = token.ExclusiveOrAssign
			} else {
				s.Token.Kind = token.ExclusiveOr
			}

		case '|':
			s.ConsumeByte()
			if s.AdvanceIfByteEquals('|') {
				if s.AdvanceIfByteEquals('=') {
					s.Token.Kind = token.LogicalOrAssign
				} else {
					s.Token.Kind = token.LogicalOr
				}
			} else if s.AdvanceIfByteEquals('=') {
				s.Token.Kind = token.OrAssign
			} else {
				s.Token.Kind = token.Or
			}

		// ---- String / template literals ----

		case '"':
			s.Token.Kind = s.scanStringLiteralDoubleQuote()

		case '\'':
			s.Token.Kind = s.scanStringLiteralSingleQuote()

		case '`':
			s.ConsumeByte()
			s.Token.Kind = s.ReadTemplateLiteral(token.TemplateHead, token.NoSubstitutionTemplate)

		// ---- Special ----

		case '#':
			if s.src.Offset() == 0 {
				afterHash := s.src.Offset() + 1
				if afterHash < s.src.EndOffset() && s.src.ReadPosition(afterHash) == '!' {
					s.ConsumeByte()
					s.ConsumeByte()
					s.skipSingleLineComment()
					continue
				}
			}
			s.scanIdentifierTail()
			s.Token.Kind = token.PrivateIdentifier

		case '\\':
			s.Token.Kind = s.identifierBackslashHandler()

		// ---- Numeric literals ----

		case '0':
			s.ConsumeByte()
			s.Token.Kind = s.readZero()

		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			s.ConsumeByte()
			s.Token.Kind = s.decimalLiteralAfterFirstDigit()

		// ---- Identifier start: keyword-leading lowercase letters ----

		case 'a':
			switch s.scanIdentifierTail() {
			case "await":
				s.Token.Kind = token.Await
			case "async":
				s.Token.Kind = token.Async
			default:
				s.Token.Kind = token.Identifier
			}

		case 'b':
			switch s.scanIdentifierTail() {
			case "break":
				s.Token.Kind = token.Break
			default:
				s.Token.Kind = token.Identifier
			}

		case 'c':
			switch s.scanIdentifierTail() {
			case "case":
				s.Token.Kind = token.Case
			case "catch":
				s.Token.Kind = token.Catch
			case "class":
				s.Token.Kind = token.Class
			case "const":
				s.Token.Kind = token.Const
			case "continue":
				s.Token.Kind = token.Continue
			default:
				s.Token.Kind = token.Identifier
			}

		case 'd':
			switch s.scanIdentifierTail() {
			case "debugger":
				s.Token.Kind = token.Debugger
			case "default":
				s.Token.Kind = token.Default
			case "delete":
				s.Token.Kind = token.Delete
			case "do":
				s.Token.Kind = token.Do
			default:
				s.Token.Kind = token.Identifier
			}

		case 'e':
			switch s.scanIdentifierTail() {
			case "else":
				s.Token.Kind = token.Else
			case "enum":
				s.Token.Kind = token.Keyword
			case "export":
				s.Token.Kind = token.Keyword
			case "extends":
				s.Token.Kind = token.Extends
			default:
				s.Token.Kind = token.Identifier
			}

		case 'f':
			switch s.scanIdentifierTail() {
			case "false":
				s.Token.Kind = token.Boolean
			case "finally":
				s.Token.Kind = token.Finally
			case "for":
				s.Token.Kind = token.For
			case "function":
				s.Token.Kind = token.Function
			default:
				s.Token.Kind = token.Identifier
			}

		case 'i':
			switch s.scanIdentifierTail() {
			case "if":
				s.Token.Kind = token.If
			case "import":
				s.Token.Kind = token.Keyword
			case "in":
				s.Token.Kind = token.In
			case "instanceof":
				s.Token.Kind = token.InstanceOf
			default:
				s.Token.Kind = token.Identifier
			}

		case 'l':
			switch s.scanIdentifierTail() {
			case "let":
				s.Token.Kind = token.Let
			default:
				s.Token.Kind = token.Identifier
			}

		case 'n':
			switch s.scanIdentifierTail() {
			case "new":
				s.Token.Kind = token.New
			case "null":
				s.Token.Kind = token.Null
			default:
				s.Token.Kind = token.Identifier
			}

		case 'o':
			switch s.scanIdentifierTail() {
			case "of":
				s.Token.Kind = token.Of
			default:
				s.Token.Kind = token.Identifier
			}

		case 'r':
			switch s.scanIdentifierTail() {
			case "return":
				s.Token.Kind = token.Return
			default:
				s.Token.Kind = token.Identifier
			}

		case 's':
			switch s.scanIdentifierTail() {
			case "super":
				s.Token.Kind = token.Super
			case "static":
				s.Token.Kind = token.Static
			case "switch":
				s.Token.Kind = token.Switch
			default:
				s.Token.Kind = token.Identifier
			}

		case 't':
			switch s.scanIdentifierTail() {
			case "this":
				s.Token.Kind = token.This
			case "throw":
				s.Token.Kind = token.Throw
			case "true":
				s.Token.Kind = token.Boolean
			case "typeof":
				s.Token.Kind = token.Typeof
			case "try":
				s.Token.Kind = token.Try
			default:
				s.Token.Kind = token.Identifier
			}

		case 'v':
			switch s.scanIdentifierTail() {
			case "var":
				s.Token.Kind = token.Var
			case "void":
				s.Token.Kind = token.Void
			default:
				s.Token.Kind = token.Identifier
			}

		case 'w':
			switch s.scanIdentifierTail() {
			case "while":
				s.Token.Kind = token.While
			case "with":
				s.Token.Kind = token.With
			default:
				s.Token.Kind = token.Identifier
			}

		case 'y':
			switch s.scanIdentifierTail() {
			case "yield":
				s.Token.Kind = token.Yield
			default:
				s.Token.Kind = token.Identifier
			}

		// ---- Identifier start: uppercase A-Z, non-keyword lowercase, _, $ ----

		case '$', '_',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'g', 'h', 'j', 'k', 'm', 'p', 'q', 'u', 'x', 'z':
			s.scanIdentifierTail()
			s.Token.Kind = token.Identifier

		// ---- Invalid ASCII ----

		case 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
			0x0E, 0x0F,
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
			0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F,
			'@', 0x7F:
			c := s.ConsumeRune()
			s.error(invalidCharacter(c, s.Token.Idx0, s.src.Offset()))
			s.Token.Kind = token.Undetermined

		// ---- Non-ASCII: valid UTF-8 leading bytes ----

		case 0xC2, 0xC3, 0xC4, 0xC5, 0xC6, 0xC7, 0xC8, 0xC9, 0xCA, 0xCB, 0xCC, 0xCD, 0xCE, 0xCF,
			0xD0, 0xD1, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7, 0xD8, 0xD9, 0xDA, 0xDB, 0xDC, 0xDD, 0xDE, 0xDF,
			0xE0, 0xE1, 0xE2, 0xE3, 0xE4, 0xE5, 0xE6, 0xE7, 0xE8, 0xE9, 0xEA, 0xEB, 0xEC, 0xED, 0xEE, 0xEF,
			0xF0, 0xF1, 0xF2, 0xF3, 0xF4:
			switch c, _ := s.PeekRune(); {
			case unicodeid.IsIDStartUnicode(c):
				s.scanIdentifierTailAfterUnicode(s.src.Offset())
				s.Token.Kind = token.Identifier
			case unicode.IsSpace(c):
				s.ConsumeRune()
				continue
			case isLineTerminator(c):
				s.ConsumeRune()
				s.Token.OnNewLine = true
				continue
			default:
				start := s.src.Offset()
				s.ConsumeRune()
				s.error(invalidCharacter(c, start, s.src.Offset()))
				s.Token.Kind = token.Undetermined
			}
		}
		break
	}
	s.Token.Idx1 = s.src.pos
}
