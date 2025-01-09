package scanner

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
	"unsafe"
)

type Scanner struct {
	token Token

	src *Source
}

func NewScanner(src string) *Scanner {
	return &Scanner{
		src: NewSource(src),
	}
}

func (s *Scanner) Next() Token {
	for {
		s.token.Idx0 = s.src.Offset()

		b, ok := s.PeekByte()
		if !ok {
			s.token.Kind = token.Eof
			break
		}

		if s.token.Kind = byteHandlers[b](s); s.token.Kind != token.Skip {
			break
		}
	}
	s.token.Idx1 = s.src.Offset()
	return s.token
}

type Checkpoint struct {
	pos unsafe.Pointer
	tok Token
	// TODO errors
}

func (s *Scanner) Checkpoint() Checkpoint {
	return Checkpoint{
		pos: s.src.ptr,
		tok: s.token,
	}
}

func (s *Scanner) Rewind(c Checkpoint) {
	s.src.ptr = c.pos
	s.token = c.tok
}

func (s *Scanner) Offset() ast.Idx {
	return s.src.Offset()
}

func (s *Scanner) NextRune() (rune, bool) {
	return s.src.NextRune()
}

func (s *Scanner) NextByte() (byte, bool) {
	return s.src.NextByte()
}

func (s *Scanner) ConsumeRune() rune {
	r, _ := s.src.NextRune()
	return r
}

func (s *Scanner) ConsumeByte() byte {
	return s.src.NextByteUnchecked()
}

func (s *Scanner) PeekRune() (rune, bool) {
	return s.src.PeekRune()
}

func (s *Scanner) PeekByte() (byte, bool) {
	return s.src.PeekByte()
}

func (s *Scanner) AdvanceIfByteEquals(b byte) bool {
	return s.src.AdvanceIfByteEquals(b)
}
