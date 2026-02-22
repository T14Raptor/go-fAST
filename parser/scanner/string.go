package scanner

import (
	"strings"
	"unsafe"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

// String end match tables indexed by byte value; true means "this byte ends the fast scan".
var (
	doubleQuoteEnd [256]bool
	singleQuoteEnd [256]bool
)

func init() {
	for _, b := range []byte{'"', '\r', '\n', '\\'} {
		doubleQuoteEnd[b] = true
	}
	for _, b := range []byte{'\'', '\r', '\n', '\\'} {
		singleQuoteEnd[b] = true
	}
}

func (s *Scanner) scanStringLiteralDoubleQuote() token.Token {
	return s.scanStringLiteral('"', &doubleQuoteEnd)
}

func (s *Scanner) scanStringLiteralSingleQuote() token.Token {
	return s.scanStringLiteral('\'', &singleQuoteEnd)
}

// scanStringLiteral reads a string literal delimited by delim.
// If a backslash is found, it switches to the escaped path.
func (s *Scanner) scanStringLiteral(delim byte, table *[256]bool) token.Token {
	// Skip opening quote
	s.src.pos++
	afterOpen := s.src.pos

	// Fast path: scan bytes directly until we hit a table match.
	// Uses local pos/base/end to avoid per-iteration function call overhead.
	pos := s.src.pos
	base := s.src.base
	end := s.src.len

	for pos < end {
		b := *(*byte)(unsafe.Add(base, pos))
		if table[b] {
			s.src.pos = pos
			switch b {
			case delim:
				// End of string found — advance past closing quote.
				s.src.pos = pos + 1
				return token.String
			case '\\':
				// Escape found — switch to escaped path.
				return s.scanStringLiteralEscaped(delim, table, afterOpen)
			default:
				// \r or \n — unterminated string.
				s.error(unterminatedString(s.unterminatedRange()))
				return token.Undetermined
			}
		}
		// Only ASCII bytes are in the table, so skipping non-matching bytes
		// (including UTF-8 continuation bytes) is safe.
		pos++
	}

	// EOF — unterminated string.
	s.src.pos = pos
	s.error(unterminatedString(s.unterminatedRange()))
	return token.Undetermined
}

// scanStringLiteralEscaped handles the string literal after finding the first backslash.
// Builds the unescaped string in a strings.Builder.
func (s *Scanner) scanStringLiteralEscaped(delim byte, table *[256]bool, afterOpen ast.Idx) token.Token {
	soFar := s.src.FromPositionToCurrent(afterOpen)
	str := &strings.Builder{}
	str.Grow(max(len(soFar)*2, 16))
	str.WriteString(soFar)

outer:
	for {
		// Consume the backslash
		escapeStart := s.src.Offset()
		s.ConsumeByte()

		isValid := true
		s.readStringEscapeSequence(str, false, &isValid)
		if !isValid {
			s.error(invalidEscapeSequence(escapeStart, s.src.Offset()))
		}

		// Consume bytes until we reach end of string, line break, or another escape
		chunkStart := s.src.Offset()
		for {
			b, ok := s.PeekByte()
			if !ok {
				// EOF - unterminated string
				s.error(unterminatedString(s.unterminatedRange()))
				return token.Undetermined
			}

			switch {
			case !table[b]:
				// Regular string character - advance past it.
				// Only ASCII bytes are in the table, safe for multi-byte chars.
				s.src.NextByteUnchecked()
				continue
			case b == delim:
				// End of string found. Push last chunk to str.
				str.WriteString(s.src.FromPositionToCurrent(chunkStart))
				// Consume closing quote
				s.src.NextByteUnchecked()
				break outer
			case b == '\\':
				// Another escape found. Push last chunk to str, and loop back.
				str.WriteString(s.src.FromPositionToCurrent(chunkStart))
				continue outer
			default:
				// \r or \n - unterminated string
				s.error(unterminatedString(s.unterminatedRange()))
				return token.Undetermined
			}
		}
	}

	s.EscapedStr = str.String()
	s.Token.HasEscape = true
	return token.String
}
