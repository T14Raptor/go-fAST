package scanner

import (
	"github.com/t14raptor/go-fast/ast"
	"strings"
)

func isLineTerminator(chr rune) bool {
	switch chr {
	case '\u000a', '\u000d', '\u2028', '\u2029':
		return true
	}
	return false
}

// Irregular line breaks - '\u{2028}' (LS) and '\u{2029}' (PS)
// These are 3-byte UTF-8 sequences starting with 0xE2.
const lsOrPsFirst byte = 0xE2

var lsBytes2And3 = [2]byte{0x80, 0xA8}
var psBytes2And3 = [2]byte{0x80, 0xA9}

// Matches: '\r', '\n', 0xE2 (first byte of LS/PS).
var lineBreakTable [256]bool

// Matches: '*', '\r', '\n', 0xE2 (first byte of LS/PS).
var multiLineCommentTable [256]bool

func init() {
	for _, b := range []byte{'\r', '\n', lsOrPsFirst} {
		lineBreakTable[b] = true
	}
	for _, b := range []byte{'*', '\r', '\n', lsOrPsFirst} {
		multiLineCommentTable[b] = true
	}
}

// skipSingleLineComment skips a single-line comment (// already consumed).
// Does NOT consume the line terminator.
func (s *Scanner) skipSingleLineComment() {
	for {
		b, ok := s.PeekByte()
		if !ok {
			// EOF - end of comment
			return
		}

		if !lineBreakTable[b] {
			// Regular character - advance past it.
			s.src.NextByteUnchecked()
			continue
		}

		if b != lsOrPsFirst {
			// '\r' or '\n' - end of comment (don't consume the line break)
			return
		}

		// 0xE2: Could be first byte of LS/PS, or some other Unicode char.
		// Consume the 0xE2 first, then check next 2 bytes.
		s.ConsumeByte()
		twoMore, ok := s.src.PeekTwoBytes()
		if !ok {
			// Not enough bytes - treat as regular char, continue
			continue
		}
		if twoMore == lsBytes2And3 || twoMore == psBytes2And3 {
			// LS or PS - end of comment.
			// Back up to the start of the 3-byte LS/PS char so the main loop handles it.
			s.src.SetPosition(s.src.Offset() - 1)
			return
		}
		// Some other Unicode char starting with 0xE2 (always 3 bytes).
		// Skip remaining 2 bytes.
		s.ConsumeByte()
		s.ConsumeByte()
	}
}

// skipMultiLineComment skips a multi-line comment (/* already consumed).
// Sets s.Token.OnNewLine if the comment contains a line terminator.
// After finding a line break, switches to a faster path that only looks for `*/`.
func (s *Scanner) skipMultiLineComment() {
	for {
		b, ok := s.PeekByte()
		if !ok {
			// Unterminated multi-line comment
			s.error(unterminatedMultiLineComment(s.unterminatedRange()))
			return
		}

		if !multiLineCommentTable[b] {
			// Regular character - advance past it
			s.src.NextByteUnchecked()
			continue
		}

		switch b {
		case '*':
			s.ConsumeByte()
			if s.AdvanceIfByteEquals('/') {
				// Found */ - end of comment
				return
			}
		case lsOrPsFirst:
			// 0xE2: Could be first byte of LS/PS or other Unicode char.
			s.ConsumeByte()
			twoMore, ok := s.src.PeekTwoBytes()
			if !ok {
				continue
			}
			if twoMore == lsBytes2And3 || twoMore == psBytes2And3 {
				// LS or PS line terminator
				s.Token.OnNewLine = true
				s.ConsumeByte()
				s.ConsumeByte()
				// Ideally we'd switch to the fast path here, but irregular
				// line breaks are rare, so just continue the main loop.
			} else {
				// Other Unicode char, skip remaining 2 bytes of 3-byte sequence
				s.ConsumeByte()
				s.ConsumeByte()
			}
		default:
			// '\r' or '\n' - regular line break
			s.Token.OnNewLine = true
			s.ConsumeByte()
			// Switch to faster search that only looks for `*/`.
			s.skipMultiLineCommentAfterLineBreak()
			return
		}
	}
}

// skipMultiLineCommentAfterLineBreak is the fast path for multi-line comment scanning
// after a line break has been found. Only needs to search for `*/`.
func (s *Scanner) skipMultiLineCommentAfterLineBreak() {
	remaining := s.src.Slice(s.src.Offset(), s.src.EndOffset())
	idx := strings.Index(remaining, "*/")
	if idx >= 0 {
		// Found `*/`, advance past it
		s.src.SetPosition(s.src.Offset() + ast.Idx(idx) + 2)
	} else {
		// Unterminated comment - advance to end
		s.src.SetPosition(s.src.EndOffset())
	}
}
