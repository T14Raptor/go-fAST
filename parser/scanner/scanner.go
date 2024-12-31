package scanner

import (
	"fmt"
	"github.com/t14raptor/go-fast/token"
)

type Scanner struct {
	src *Source

	token Token
}

func NewScanner(src string) *Scanner {
	return &Scanner{
		src: NewSource(src),
	}
}

func (s *Scanner) Next() Token {
	s.token.idx0 = s.src.Offset()
	s.token.kind = s.readNextToken()
	s.token.idx1 = s.src.Offset()
	t := s.token
	s.token = Token{}
	return t
}

func (s *Scanner) readNextToken() token.Token {
	for {
		b, ok := s.PeekByte()
		if !ok {
			return token.Eof
		}

		fmt.Println("hello", string(b))

		if kind := byteHandlers[b](s); kind != token.Skip {
			return kind
		}
	}
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
