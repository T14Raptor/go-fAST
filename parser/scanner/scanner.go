package scanner

import (
	"github.com/t14raptor/go-fast/token"
	"unsafe"
)

type Scanner struct {
	src *Source

	token token.Token
}

func NewScanner(src string) *Scanner {
	return &Scanner{
		src: NewSource(src),
	}
}

func (s *Scanner) EOF() bool {
	return s.ptr == s.end
}

func (s *Scanner) End() *byte {
	return s.end
}

func (s *Scanner)

func (s *Scanner) NextByte() (byte, bool) {
	if s.EOF() {
		return 0, false
	}
	return s.NextByteUnchecked(), true
}

func (s *Scanner) NextByteUnchecked() byte {
	b := s.PeekByteUnchecked()
	s.add()
	return b
}

func (s *Scanner) PeekByte() (byte, bool) {
	if s.EOF() {
		return 0, false
	}
	return s.PeekByteUnchecked(), true
}

func (s *Scanner) PeekByteUnchecked() byte {
	return *s.ptr
}

func (s *Scanner) AdvanceIfByteEquals(b byte) (matched bool) {
	nextB, ok := s.PeekByte()
	if !ok {
		return false
	}
	if matched = nextB == b; matched {
		s.add()
	}
	return matched
}

func (s *Scanner) add() {
	s.ptr = (*byte)(unsafe.Add(unsafe.Pointer(s.ptr), 1))
}
