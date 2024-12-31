package scanner

import (
	"github.com/t14raptor/go-fast/token"
	"unicode/utf8"
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
		// TODO error
		return token.Undetermined
	}

	for {
		b, ok := s.PeekByte()
		if !ok {
			break
		}

		if b == '_' {
			s.ConsumeByte()

			// TODO
			if b, ok := s.PeekByte(); ok && digitValue(b) < base {
				s.ConsumeByte()
			} else {
				// TODO error
				return token.Undetermined
			}
		} else if b, ok := s.PeekByte(); ok && digitValue(b) < base {
			s.ConsumeByte()
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
		}
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
		// TODO error
		return
	}
	s.decimalDigitsAfterFirstDigit()
}

func (s *Scanner) decimalDigitsAfterFirstDigit() {
	for {
		b, ok := s.PeekByte()
		if !ok {
			break
		}

		switch b {
		case '_':
			s.ConsumeRune()

			if b, ok := s.PeekByte(); ok && isDecimalDigit(b) {
				s.ConsumeByte()
			} else {
				// TODO error
				return
			}
			// TODO
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			s.ConsumeRune()
		default:
			break
		}
	}
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

	s.ConsumeRune()
	for {
		if c, ok := s.PeekRune(); ok && isIdentifierStart(c) {
			s.ConsumeRune()
		} else {
			break
		}
	}
	// TODO report error
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
