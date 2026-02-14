package scanner

import (
	"unicode/utf8"
	"unsafe"

	"github.com/t14raptor/go-fast/token"
)

func (s *Scanner) readZero() token.Token {
	b, ok := s.PeekByte()
	if !ok {
		return s.checkAfterNumericLiteral(token.Number)
	}

	switch b {
	case 'b', 'B':
		return s.readNonDecimal(2)
	case 'o', 'O':
		return s.readNonDecimal(8)
	case 'x', 'X':
		return s.readNonDecimal(16)
	case 'e', 'E':
		s.ConsumeByte()
		s.readDecExp()
		return token.Number
	case '.':
		s.ConsumeByte()
		return s.decLitAfterDecPointAfterDigits()
	case 'n':
		s.ConsumeByte()
		return s.checkAfterNumericLiteral(token.Number)
	}

	// legacy
	return s.readLegacyOctal()
}

func (s *Scanner) decimalLiteralAfterFirstDigit() token.Token {
	s.decimalDigitsAfterFirstDigit()
	if s.AdvanceIfByteEquals('.') {
		return s.decLitAfterDecPointAfterDigits()
	} else if s.AdvanceIfByteEquals('n') {
		return s.checkAfterNumericLiteral(token.Number)
	}

	return s.checkAfterNumericLiteral(token.Number)
}

func (s *Scanner) readNonDecimal(base int) token.Token {
	s.ConsumeRune()

	if b, ok := s.PeekByte(); ok && digitValue(b) < base {
		s.ConsumeByte()
	} else {
		s.unexpectedErr()
		return token.Undetermined
	}

	for {
		b, ok := s.PeekByte()
		if !ok {
			break
		}

		if b == '_' {
			s.ConsumeByte()

			if b, ok := s.PeekByte(); ok && digitValue(b) < base {
				s.ConsumeByte()
			} else {
				s.unexpectedErr()
				return token.Undetermined
			}
		} else if b, ok := s.PeekByte(); ok && digitValue(b) < base {
			s.ConsumeByte()
		} else {
			break
		}
	}

	s.AdvanceIfByteEquals('n')
	return s.checkAfterNumericLiteral(token.Number)
}

func (s *Scanner) readLegacyOctal() token.Token {
	for {
		b, ok := s.PeekByte()
		if !ok {
			break
		}

		switch b {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			s.ConsumeByte()
			continue
		}
		break
	}

	if b, ok := s.PeekByte(); ok {
		switch b {
		case '.':
			s.ConsumeByte()
			return s.decLitAfterDecPointAfterDigits()
		case 'e':
			s.ConsumeByte()
			s.readDecExp()
		}
	}
	return s.checkAfterNumericLiteral(token.Number)
}

func (s *Scanner) readDecExp() {
	if b, ok := s.PeekByte(); ok {
		switch b {
		case '-', '+':
			s.ConsumeByte()
		}
	}
	s.readDecimalDigits()
}

func (s *Scanner) readDecimalDigits() {
	if b, ok := s.PeekByte(); ok && isDecimalDigit(b) {
		s.ConsumeByte()
	} else {
		s.unexpectedErr()
		return
	}
	s.decimalDigitsAfterFirstDigit()
}

func (s *Scanner) decimalDigitsAfterFirstDigit() {
	pos := s.src.pos
	base := s.src.base
	end := s.src.len

	// Fast path: scan pure digit runs with direct memory access.
	for pos < end {
		b := *(*byte)(unsafe.Add(base, pos))
		if b >= '0' && b <= '9' {
			pos++
			continue
		}
		if b == '_' {
			// Numeric separator — must be followed by a digit.
			s.src.pos = pos + 1
			if nb, ok := s.PeekByte(); ok && isDecimalDigit(nb) {
				s.ConsumeByte()
				pos = s.src.pos
				continue
			}
			// Trailing numeric separator
			s.unexpectedErr()
			return
		}
		break
	}

	s.src.pos = pos
}

func (s *Scanner) decLitAfterDecPoint() token.Token {
	s.readDecimalDigits()
	s.optionalExp()
	return s.checkAfterNumericLiteral(token.Number)
}

func (s *Scanner) decLitAfterDecPointAfterDigits() token.Token {
	s.optionalDecDigits()
	s.optionalExp()
	return s.checkAfterNumericLiteral(token.Number)
}

func (s *Scanner) optionalDecDigits() {
	if b, ok := s.PeekByte(); ok && isDecimalDigit(b) {
		s.ConsumeByte()
		s.decimalDigitsAfterFirstDigit()
	}
}

func (s *Scanner) optionalExp() {
	b, ok := s.PeekByte()
	if ok && (b == 'e' || b == 'E') {
		s.ConsumeByte()
		s.readDecExp()
	}
}

func (s *Scanner) checkAfterNumericLiteral(kind token.Token) token.Token {
	switch b, ok := s.PeekByte(); {
	case ok && b < utf8.RuneSelf:
		if !asciiContinue[b] {
			return kind
		}
	case ok:
		c, _ := s.PeekRune()
		if !isIdentifierStart(c) {
			return kind
		}
	default:
		return kind
	}

	offset := s.src.Offset()
	s.ConsumeRune()
	for {
		if c, ok := s.PeekRune(); ok && isIdentifierStart(c) {
			s.ConsumeRune()
		} else {
			break
		}
	}
	s.error(invalidNumberEnd(offset, s.src.Offset()))
	return token.Undetermined
}

func isDecimalDigit(chr byte) bool {
	return '0' <= chr && chr <= '9'
}

func digitValue(chr byte) int {
	switch {
	case '0' <= chr && chr <= '9':
		return int(chr - '0')
	case 'a' <= chr && chr <= 'f':
		return int(chr - 'a' + 10)
	case 'A' <= chr && chr <= 'F':
		return int(chr - 'A' + 10)
	}
	return 16 // Larger than any legal digit value
}
