package scanner

import (
	"github.com/t14raptor/go-fast/token"
	"unsafe"
)

func (s *Scanner) handleLineBreak() token.Token {
	s.Token.OnNewLine = true

	pos := s.src.pos
	base := s.src.base
	end := s.src.len
	for pos < end {
		b := *(*byte)(unsafe.Add(base, pos))
		if b != ' ' && b != '\t' && b != '\r' && b != '\n' {
			break
		}
		pos++
	}
	s.src.pos = pos
	return token.Skip
}
