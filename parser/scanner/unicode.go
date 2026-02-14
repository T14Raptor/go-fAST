package scanner

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

func (s *Scanner) identifierUnicodeEscapeSequence(str *strings.Builder, checkIdentifierStart bool) {
	start := s.src.Offset()

	b, ok := s.PeekByte()
	if !ok || b != 'u' {
		s.error(invalidUnicodeEscapeSequence(start, s.src.Offset()))
		return
	}
	s.ConsumeByte()

	b, _ = s.PeekByte()
	var value rune
	if b == '{' {
		value = s.unicodeCodePoint()
	} else {
		value = s.surrogatePair()
	}
	if value == '\\' || value == -1 {
		s.error(invalidUnicodeEscapeSequence(start, s.src.Offset()))
		return
	}

	if checkIdentifierStart {
		if !isIdentifierStart(value) {
			s.error(invalidUnicodeEscapeSequence(start, s.src.Offset()))
			return
		}
	} else if !isIdentifierPart(value) {
		s.error(invalidUnicodeEscapeSequence(start, s.src.Offset()))
		return
	}

	str.WriteRune(value)
}

func (s *Scanner) stringUnicodeEscapeSequence(str *strings.Builder, isValid *bool) {
	b, ok := s.PeekByte()
	if !ok {
		*isValid = false
		return
	}

	var value rune
	if b == '{' {
		value = s.unicodeCodePoint()
	} else {
		value = s.surrogatePair()
	}

	if value == -1 {
		*isValid = false
		return
	}

	str.WriteRune(value)
}

func (s *Scanner) unicodeCodePoint() rune {
	if !s.AdvanceIfByteEquals('{') {
		return -1
	}

	val := s.codePoint()
	if val == -1 {
		return -1
	}

	if !s.AdvanceIfByteEquals('}') {
		return -1
	}
	return val
}

func (s *Scanner) hexFourDigits() (val rune) {
	for i := 0; i < 4; i++ {
		next, ok := s.hexDigit()
		if !ok {
			return -1
		}
		val = (val << 4) | next
	}
	return val
}

// hexDigit peeks the next byte and, if it's a hex digit, consumes it and returns its value.
// If not a hex digit (or at EOF), returns (0, false) without consuming.
func (s *Scanner) hexDigit() (rune, bool) {
	b, ok := s.PeekByte()
	if !ok {
		return 0, false
	}

	// `b | 32` folds uppercase to lowercase in one branch.
	var value byte
	if b >= '0' && b <= '9' {
		value = b - '0'
	} else {
		lower := b | 32
		if lower >= 'a' && lower <= 'f' {
			value = lower - 'a' + 10
		} else {
			return 0, false
		}
	}

	// Only consume after confirming it's a valid hex digit.
	s.ConsumeByte()
	return rune(value), true
}

func (s *Scanner) codePoint() rune {
	val, ok := s.hexDigit()
	if !ok {
		return -1
	}
	for {
		next, ok := s.hexDigit()
		if !ok {
			break
		}
		val = (val << 4) | next
		if val > utf8.MaxRune {
			return -1
		}
	}
	return val
}

func (s *Scanner) surrogatePair() rune {
	high := s.hexFourDigits()
	if !utf16.IsSurrogate(high) {
		return high
	} else if b, ok := s.src.PeekTwoBytes(); !ok || b != [2]byte{'\\', 'u'} {
		return high
	}

	s.ConsumeByte()
	s.ConsumeByte()

	low := s.hexFourDigits()
	if !utf16.IsSurrogate(low) {
		// invalid
		return -1
	}

	return utf16.DecodeRune(high, low)
}

// readStringEscapeSequence processes one escape sequence after the backslash has been consumed.
//
// When inTemplate is false (string literals):
//   - Legacy octal escapes (\0-\7) are valid and produce the octal value
//   - \8 and \9 are identity escapes (NonOctalDecimalEscapeSequence)
//
// When inTemplate is true (template literals):
//   - \0 followed by a digit is invalid
//   - \1-\9 are all invalid
func (s *Scanner) readStringEscapeSequence(str *strings.Builder, inTemplate bool, isValid *bool) {
	chr, ok := s.NextRune()
	if !ok {
		*isValid = false
		return
	}

	switch chr {
	// \ LineTerminatorSequence - line continuation produces nothing
	case '\u000a', '\u2028', '\u2029':
		// LF, LS, PS
	case '\u000d':
		// CR - also consume LF if present (\r\n line continuation)
		s.AdvanceIfByteEquals('\n')

	// SingleEscapeCharacter :: one of ' " \ b f n r t v
	case '\'', '"', '\\':
		str.WriteRune(chr)
	case 'b':
		str.WriteRune('\u0008')
	case 'f':
		str.WriteRune('\u000c')
	case 'n':
		str.WriteRune('\u000a')
	case 'r':
		str.WriteRune('\u000d')
	case 't':
		str.WriteRune('\u0009')
	case 'v':
		str.WriteRune('\u000b')

	// HexEscapeSequence
	case 'x':
		d1, ok1 := s.hexDigit()
		if !ok1 {
			*isValid = false
			return
		}
		d2, ok2 := s.hexDigit()
		if !ok2 {
			*isValid = false
			return
		}
		str.WriteRune((d1 << 4) | d2)

	// UnicodeEscapeSequence
	case 'u':
		s.stringUnicodeEscapeSequence(str, isValid)

	// \0 [lookahead ∉ DecimalDigit]
	case '0':
		if b, ok := s.PeekByte(); ok && b >= '0' && b <= '9' {
			if inTemplate {
				// In templates: \0 followed by digit is invalid
				s.ConsumeByte()
				*isValid = false
			} else {
				// In strings: LegacyOctalEscapeSequence starting with 0
				s.readLegacyOctal()
			}
		} else {
			str.WriteRune('\u0000')
		}

	case '1', '2', '3', '4', '5', '6', '7':
		if inTemplate {
			// In templates: \1-\7 are invalid escape sequences
			*isValid = false
		} else {
			// In strings: LegacyOctalEscapeSequence
			s.readLegacyOctal()
		}

	case '8', '9':
		if inTemplate {
			// In templates: \8 and \9 are invalid
			*isValid = false
		} else {
			// In strings: NonOctalDecimalEscapeSequence - just the character itself
			str.WriteRune(chr)
		}

	default:
		// Identity escape - the character itself
		str.WriteRune(chr)
	}
}
