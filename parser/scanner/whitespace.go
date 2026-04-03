package scanner

import (
	"unsafe"
)

func (s *Scanner) handleLineBreak() {
	s.Token.OnNewLine = true

	pos := s.src.pos
	base := s.src.base
	end := s.src.len
	for pos < end {
		b := *(*byte)(unsafe.Add(base, pos))

		switch b {
		case ' ', '\t', '\r', '\n':
			pos++
			continue
		}
		break
	}
	s.src.pos = pos
}
