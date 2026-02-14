package scanner

import (
	"strings"
	"unicode/utf8"
	"unsafe"

	"github.com/nukilabs/unicodeid"
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

// Lookup tables for ASCII identifier characters.
// Non-ASCII bytes (>= 128) are always false, branching to the Unicode path.
var asciiStart, asciiContinue [256]bool

func init() {
	for i := 0; i < 128; i++ {
		if i >= 'a' && i <= 'z' || i >= 'A' && i <= 'Z' || i == '$' || i == '_' {
			asciiStart[i] = true
			asciiContinue[i] = true
		}
		if i >= '0' && i <= '9' {
			asciiContinue[i] = true
		}
	}
}

// Fast path for checking “start” of an identifier.
func isIdentifierStart(chr rune) bool {
	// 0) Invalid path
	if chr == -1 {
		return false
	}
	// 1) ASCII path
	if chr < utf8.RuneSelf {
		return asciiStart[chr]
	}

	// 2) Non-ASCII path
	return unicodeid.IsIDStartUnicode(chr)
}

// Fast path for checking “continuation” of an identifier.
func isIdentifierPart(chr rune) bool {
	// 0) Invalid path
	if chr == -1 {
		return false
	}
	// 1) ASCII path
	if chr < utf8.RuneSelf {
		return asciiContinue[chr]
	}

	// 2) Non-ASCII path
	return unicodeid.IsIDContinueUnicode(chr)
}

func (s *Scanner) scanIdentifierTail() string {
	start := s.src.pos
	pos := start + 1 // skip first byte (already identified as identifier start by caller)
	base := s.src.base
	end := s.src.len

	for pos < end {
		b := *(*byte)(unsafe.Add(base, pos))
		if !asciiContinue[b] {
			s.src.pos = pos
			if b >= utf8.RuneSelf {
				return s.scanIdentifierTailUnicode(start)
			}
			if b == '\\' {
				return s.scanIdentifierBackslash(start, false)
			}
			return s.src.FromPositionToCurrent(start)
		}
		pos++
	}

	s.src.pos = pos
	return s.src.FromPositionToCurrent(start)
}

func (s *Scanner) scanIdentifierTailUnicode(start ast.Idx) string {
	c, _ := s.PeekRune()
	if unicodeid.IsIDContinueUnicode(c) {
		s.ConsumeRune()
		return s.scanIdentifierTailAfterUnicode(start)
	}
	return s.src.FromPositionToCurrent(start)
}

func (s *Scanner) scanIdentifierTailAfterUnicode(start ast.Idx) string {
	for {
		c, ok := s.PeekRune()
		if !ok {
			break
		}

		if isIdentifierPart(c) {
			s.ConsumeRune()
		} else if c == '\\' {
			return s.scanIdentifierBackslash(start, false)
		} else {
			break
		}
	}

	return s.src.FromPositionToCurrent(start)
}

func (s *Scanner) identifierBackslashHandler() token.Token {
	str := &strings.Builder{}
	str.Grow(16)

	id := s.scanIdentifierOnBackslash(str, true)
	return token.MatchKeyword(id)
}

func (s *Scanner) scanIdentifierBackslash(startPos ast.Idx, start bool) string {
	soFar := s.src.FromPositionToCurrent(startPos)

	str := &strings.Builder{}
	str.Grow(max(len(soFar)*2, 16))

	str.WriteString(soFar)

	return s.scanIdentifierOnBackslash(str, start)
}

func (s *Scanner) scanIdentifierOnBackslash(str *strings.Builder, start bool) string {
outer:
	for {
		s.ConsumeRune()

		s.identifierUnicodeEscapeSequence(str, start)
		start = false

		chunkStart := s.src.Offset()
		for {
			c, ok := s.PeekRune()
			if ok && isIdentifierPart(c) {
				s.ConsumeRune()
				continue
			}

			chunk := s.src.FromPositionToCurrent(chunkStart)
			str.WriteString(chunk)

			if !ok || c != '\\' {
				// End of identifier or EOF
				break outer
			}

			break
		}
	}

	s.EscapedStr = str.String()
	s.Token.HasEscape = true
	return s.EscapedStr
}
