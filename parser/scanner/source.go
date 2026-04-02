package scanner

import (
	"github.com/t14raptor/go-fast/ast"
	"unicode/utf8"
	"unsafe"
)

type Source struct {
	base unsafe.Pointer
	pos  ast.Idx
	len  ast.Idx
}

func NewSource(src string) Source {
	return Source{
		base: unsafe.Pointer(unsafe.StringData(src)),
		pos:  0,
		len:  ast.Idx(len(src)),
	}
}

func (s *Source) EOF() bool {
	return s.pos >= s.len
}

func (s *Source) Offset() ast.Idx {
	return s.pos
}

func (s *Source) EndOffset() ast.Idx {
	return s.len
}

func (s *Source) SetPosition(pos ast.Idx) {
	s.pos = pos
}

func (s *Source) ReadPosition(pos ast.Idx) byte {
	return *(*byte)(unsafe.Add(s.base, pos))
}

func (s *Source) NextRune() (rune, bool) {
	b, ok := s.PeekByte()
	if !ok {
		return 0, false
	}
	if b < utf8.RuneSelf {
		s.pos++
		return rune(b), true
	}

	str := newString(unsafe.Add(s.base, s.pos), uintptr(s.len-s.pos))
	var chr rune
	for _, chr = range str {
		break
	}
	s.pos += ast.Idx(utf8.RuneLen(chr))
	return chr, true
}

func (s *Source) PeekRune() (rune, bool) {
	b, ok := s.PeekByte()
	if !ok {
		return 0, false
	}
	if b < utf8.RuneSelf {
		return rune(b), true
	}

	str := newString(unsafe.Add(s.base, s.pos), uintptr(s.len-s.pos))
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
	b := *(*byte)(unsafe.Add(s.base, s.pos))
	s.pos++
	return b
}

func (s *Source) PeekByte() (byte, bool) {
	if s.EOF() {
		return 0, false
	}
	return s.PeekByteUnchecked(), true
}

func (s *Source) PeekTwoBytes() ([2]byte, bool) {
	if s.len-s.pos >= 2 {
		return *(*[2]byte)(unsafe.Add(s.base, s.pos)), true
	}
	return [2]byte{}, false
}

func (s *Source) PeekByteUnchecked() byte {
	return *(*byte)(unsafe.Add(s.base, s.pos))
}

func (s *Source) AdvanceIfByteEquals(b byte) (matched bool) {
	nextB, ok := s.PeekByte()
	if ok && nextB == b {
		s.pos++
		return true
	}
	return false
}

func (s *Source) FromPositionToCurrent(pos ast.Idx) string {
	return newString(unsafe.Add(s.base, pos), uintptr(s.pos-pos))
}

func (s *Source) Slice(from, to ast.Idx) string {
	return newString(unsafe.Add(s.base, from), uintptr(to-from))
}

type unsafeString struct {
	Data unsafe.Pointer
	Len  int
}

func newString(data unsafe.Pointer, l uintptr) string {
	return *(*string)(unsafe.Pointer(&unsafeString{Data: data, Len: int(l)}))
}
