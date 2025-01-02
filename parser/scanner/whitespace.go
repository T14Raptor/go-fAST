package scanner

import "github.com/t14raptor/go-fast/token"

func (s *Scanner) handleLineBreak() token.Token {
	s.token.OnNewLine = true

	for {
		b, ok := s.PeekByte()
		if !ok {
			break
		}

		switch b {
		case ' ', '\t', '\r', '\n':
			s.ConsumeByte()
			continue
		}

		break
	}

	return token.Skip
}
