package scanner

import "github.com/t14raptor/go-fast/token"

func (s *Scanner) readDot() token.Token {
	if s.AdvanceIfByteEquals('.') {
		if s.AdvanceIfByteEquals('.') {
			return token.Ellipsis
		}
		return token.Illegal
	}
	if b, ok := s.PeekByte(); ok && isDecimalDigit(b) {
		return s.decLitAfterDecPoint()
	}
	return token.Period
}
