package scanner

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
	"strings"
)

// Matches: '$', '`', '\r', '\\'.
var templateLiteralEnd [256]bool

func init() {
	for _, b := range []byte{'$', '`', '\r', '\\'} {
		templateLiteralEnd[b] = true
	}
}

// ReadTemplateLiteral scans the body of a template literal.
// The opening delimiter (` or }) must already have been consumed by the caller.
// sub is the Token to return when encountering ${, tail is for closing `.
func (s *Scanner) ReadTemplateLiteral(sub, tail token.Token) token.Token {
	ret := sub
	contentStart := s.src.Offset()

	for {
		b, ok := s.PeekByte()
		if !ok {
			// Unterminated template literal
			s.error(unterminatedTemplateLiteral(s.unterminatedRange()))
			return token.Undetermined
		}

		if !templateLiteralEnd[b] {
			s.src.NextByteUnchecked()
			continue
		}

		switch b {
		case '$':
			afterDollar := s.src.Offset() + 1
			if afterDollar < s.src.EndOffset() && s.src.ReadPosition(afterDollar) == '{' {
				// Skip `${` and stop
				s.ConsumeByte() // $
				s.ConsumeByte() // {
				return ret      // = sub
			}
			// Not `${`, continue scanning
			s.ConsumeByte()
		case '`':
			// Skip '`' and stop
			s.ConsumeByte()
			ret = tail
			return ret
		case '\r':
			// Switch to escaped path for \r normalization.
			return s.templateLiteralCarriageReturn(contentStart, sub, tail)
		default: // '\\'
			// Switch to escaped path for escape sequence processing.
			return s.templateLiteralBackslash(contentStart, sub, tail)
		}
	}
}

// templateLiteralCarriageReturn handles a \r in a template literal by switching to
// the escaped (string builder) path. Normalizes \r and \r\n to \n.
func (s *Scanner) templateLiteralCarriageReturn(contentStart ast.Idx, sub, tail token.Token) token.Token {
	// Create string builder with content up to before \r
	str := s.createTemplateLiteralString(contentStart)

	// Consume \r
	s.ConsumeByte()

	// Set chunk start to after \r
	chunkStart := s.src.Offset()

	if s.AdvanceIfByteEquals('\n') {
		// \r\n: The \n will be the first byte of the next chunk, so it gets included
		// in str when that chunk is pushed. Don't push \n here.
		// chunkStart stays before the \n.
		// (AdvanceIfByteEquals already consumed \n, back up chunkStart to include it)
		chunkStart = s.src.Offset() - 1
	} else {
		// Lone \r: convert to \n by pushing \n to str.
		// chunkStart is after \r, so the \r won't be in the next chunk.
		str.WriteByte('\n')
	}

	isValid := true
	return s.templateLiteralEscaped(str, chunkStart, &isValid, sub, tail)
}

// templateLiteralBackslash handles a backslash escape in a template literal
// by switching to the escaped (string builder) path.
func (s *Scanner) templateLiteralBackslash(contentStart ast.Idx, sub, tail token.Token) token.Token {
	// Create string builder with content up to before backslash
	str := s.createTemplateLiteralString(contentStart)

	// Consume the backslash and process escape sequence (inTemplate=true)
	s.ConsumeByte()
	isValid := true
	s.readStringEscapeSequence(str, true, &isValid)

	chunkStart := s.src.Offset()
	return s.templateLiteralEscaped(str, chunkStart, &isValid, sub, tail)
}

// createTemplateLiteralString creates a string builder initialized with the
// content from contentStart to the current scanner position.
func (s *Scanner) createTemplateLiteralString(contentStart ast.Idx) *strings.Builder {
	soFar := s.src.FromPositionToCurrent(contentStart)
	b := &strings.Builder{}
	b.Grow(max(len(soFar)*2, 16))
	b.WriteString(soFar)
	return b
}

// templateLiteralEscaped continues scanning a template literal using a
// string builder for escape-processed content. chunkStart tracks where the
// current unescaped chunk begins in the source.
func (s *Scanner) templateLiteralEscaped(str *strings.Builder, chunkStart ast.Idx, isValid *bool, sub, tail token.Token) token.Token {
	ret := sub

	for {
		b, ok := s.PeekByte()
		if !ok {
			// Unterminated template literal - save what we have
			s.error(unterminatedTemplateLiteral(s.unterminatedRange()))
			str.WriteString(s.src.FromPositionToCurrent(chunkStart))
			s.EscapedStr = str.String()
			s.Token.HasEscape = true
			return token.Undetermined
		}

		if !templateLiteralEnd[b] {
			s.src.NextByteUnchecked()
			continue
		}

		switch b {
		case '$':
			afterDollar := s.src.Offset() + 1
			if afterDollar < s.src.EndOffset() && s.src.ReadPosition(afterDollar) == '{' {
				// Add last chunk to str
				str.WriteString(s.src.FromPositionToCurrent(chunkStart))
				s.ConsumeByte() // $
				s.ConsumeByte() // {
				s.EscapedStr = str.String()
				s.Token.HasEscape = true
				return ret // = sub
			}
			// Not `${`, continue scanning
			s.ConsumeByte()

		case '`':
			// End of template literal
			str.WriteString(s.src.FromPositionToCurrent(chunkStart))
			s.ConsumeByte()
			ret = tail
			s.EscapedStr = str.String()
			s.Token.HasEscape = true
			return ret

		case '\r':
			// Add chunk up to before \r to str
			str.WriteString(s.src.FromPositionToCurrent(chunkStart))
			s.ConsumeByte() // consume \r

			// Set next chunk to start after \r
			chunkStart = s.src.Offset()

			if s.AdvanceIfByteEquals('\n') {
				// \r\n: include the \n in next chunk.
				// The \n will be the first char of the chunk, getting the correct
				// line break representation. Adjust chunkStart to include it.
				chunkStart = s.src.Offset() - 1
			} else {
				// Lone \r, convert to \n
				str.WriteByte('\n')
				// chunkStart already set after \r
			}

		default: // '\\'
			// Add chunk up to before backslash to str
			str.WriteString(s.src.FromPositionToCurrent(chunkStart))
			// Consume backslash
			s.ConsumeByte()
			// Process escape sequence with inTemplate=true
			s.readStringEscapeSequence(str, true, isValid)
			// Start next chunk after escape sequence
			chunkStart = s.src.Offset()
		}
	}
}
