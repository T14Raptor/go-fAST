package scanner

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

func (s *Scanner) identifierUnicodeEscapeSequence(str *strings.Builder, checkIdentifierStart bool) {
	b, ok := s.PeekByte()
	if !ok || b != 'u' {
		// TODO report error
		return
	}
	s.ConsumeByte()

	b, ok = s.PeekByte()
	var value rune
	if b == '{' {
		value = s.codePoint()
	} else {
		value = s.surrogatePair()
	}
	if value == '\\' {
		// TODO error fmt.Sprintf("Invalid identifier escape value: %c (%s)", value, string(value))
		return
	}

	if checkIdentifierStart {
		if !isIdentifierStart(value) {
			// TODO error fmt.Sprintf("Invalid identifier escape value: %c (%s)", value, string(value))
			return
		}
	} else if !isIdentifierPart(value) {
		// TODO error fmt.Sprintf("Invalid identifier escape value: %c (%s)", value, string(value))
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
		value = s.codePoint()
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
	if val != -1 {
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

func (s *Scanner) hexDigit() (rune, bool) {
	chr, ok := s.NextRune()
	if !ok {
		return 0, false
	}
	switch {
	case '0' <= chr && chr <= '9':
		return chr - '0', true
	case 'a' <= chr && chr <= 'f':
		return chr - 'a' + 10, true
	case 'A' <= chr && chr <= 'F':
		return chr - 'A' + 10, true
	}
	return 0, false
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

func (s *Scanner) readStringEscapeSequence(str *strings.Builder, isValid *bool) {
	chr, ok := s.NextRune()
	if !ok {
		// TODO
		return
	}

	switch chr {
	case '\u000a', '\u2028', '\u2029':
	case '\u000d':
		s.AdvanceIfByteEquals('\n')
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
	case 'x':
		if d, ok := s.hexDigit(); ok {
			if nextD, ok := s.hexDigit(); ok {
				val := d<<4 | nextD
				str.WriteRune(val)
			}
		}
	case 'u':
		s.stringUnicodeEscapeSequence(str, isValid)
	case '0':
		// TODO
	}
}
