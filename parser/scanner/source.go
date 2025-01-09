package scanner

import (
	"github.com/t14raptor/go-fast/ast"
	"unicode/utf8"
	"unsafe"
)

type Source struct {
	ptr unsafe.Pointer

	start, end unsafe.Pointer
}

func NewSource(src string) *Source {
	s := &Source{
		start: unsafe.Pointer(unsafe.StringData(src)),
	}
	s.end = unsafe.Add(s.start, len(src)-1)
	s.ptr = s.start
	return s
}

func (s *Source) Slice(from, to ast.Idx) string {
	return newString(unsafe.Add(s.start, from), uintptr(to)-uintptr(from))
}

func (s *Source) EOF() bool {
	return s.ptr == s.end
}

func (s *Source) Offset() ast.Idx {
	return ast.Idx(uintptr(s.ptr) - uintptr(s.start))
}

func (s *Source) NextRune() (rune, bool) {
	b, ok := s.PeekByte()
	if !ok {
		return 0, false
	}
	if b <= utf8.RuneSelf {
		s.ptr = unsafe.Add(s.ptr, 1)
		return rune(b), true
	}

	str := newString(s.ptr, uintptr(s.end)-uintptr(s.ptr))
	var chr rune
	for _, chr = range str {
		break
	}
	s.ptr = unsafe.Add(s.ptr, utf8.RuneLen(chr))
	return chr, true
}

func (s *Source) PeekRune() (rune, bool) {
	b, ok := s.PeekByte()
	if !ok {
		return 0, false
	}
	if b <= utf8.RuneSelf {
		return rune(b), true
	}

	str := newString(s.ptr, uintptr(s.end)-uintptr(s.ptr))
	var chr rune
	for _, chr = range str {
		break
	}
	return chr, true
}

func (s *Source) NextByte() (byte, bool) {
	if s.EOF() {
		return 0, false
	}
	return s.NextByteUnchecked(), true
}

func (s *Source) NextByteUnchecked() byte {
	b := s.PeekByteUnchecked()
	s.ptr = unsafe.Add(s.ptr, 1)
	return b
}

func (s *Source) PeekByte() (byte, bool) {
	if s.EOF() {
		return 0, false
	}
	return s.PeekByteUnchecked(), true
}

func (s *Source) PeekTwoBytes() ([2]byte, bool) {
	if uintptr(s.end)-uintptr(s.ptr) >= 2 {
		return *(*[2]byte)(s.ptr), true
	}
	return [2]byte{}, false
}

func (s *Source) PeekByteUnchecked() byte {
	return *(*byte)(s.ptr)
}

func (s *Source) AdvanceIfByteEquals(b byte) (matched bool) {
	nextB, ok := s.PeekByte()
	if ok && nextB == b {
		s.ptr = unsafe.Add(s.ptr, 1)
		return true
	}
	return false
}

func (s *Source) FromPositionToCurrent(pos ast.Idx) string {
	p := unsafe.Add(s.start, pos)
	return newString(p, uintptr(s.ptr)-uintptr(p))
}

type unsafeString struct {
	Data unsafe.Pointer
	Len  int
}

func newString(data unsafe.Pointer, l uintptr) string {
	return *(*string)(unsafe.Pointer(&unsafeString{Data: data, Len: int(l)}))
}
