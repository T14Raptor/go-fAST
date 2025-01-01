package scanner

import (
	"fmt"
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
	"slices"
	"strings"
)

func (s *Scanner) scanStringLiteralDoubleQuote() token.Token {
	return s.scanStringLiteral('"', []byte{'"', '\r', '\n', '\\'})
}

func (s *Scanner) scanStringLiteralSingleQuote() token.Token {
	return s.scanStringLiteral('\'', []byte{'\'', '\r', '\n', '\\'})
}

func (s *Scanner) scanStringLiteral(delim byte, table []byte) token.Token {
	s.ConsumeByte()
	afterOpen := s.src.Offset()

	var b byte
	for {
		var ok bool
		b, ok = s.PeekByte()
		if !ok {
			return token.Undetermined
		}

		if slices.Contains(table, b) {
			break
		}

		s.ConsumeByte()
	}

	switch b {
	case delim:
		s.NextByte()
		return token.String
	case '\\':
		return s.scanStringLiteralEscaped(delim, table, afterOpen)
	default:
		s.ConsumeRune()
		return token.Undetermined
	}
}

func (s *Scanner) scanStringLiteralEscaped(delim byte, table []byte, afterOpen ast.Idx) token.Token {
	soFar := s.src.FromPositionToCurrent(afterOpen)
	str := &strings.Builder{}
	str.Grow(max(len(soFar)*2, 16))

	str.WriteString(soFar)

outer:
	for {
		escapeStartOffset := s.src.Offset()
		s.ConsumeRune()

		isValid := true
		s.readStringEscapeSequence(str, &isValid)
		if !isValid {
			_ = escapeStartOffset
			fmt.Println("invalid string escape sequence")
			// TODO: report error
		}

		chunkStart := s.src.Offset()
		for {
			b, ok := s.PeekByte()
			if !ok {
				break
			}

			switch {
			case !slices.Contains(table, b):
				s.src.NextByteUnchecked()
				continue
			case b == delim:
				str.WriteString(s.src.FromPositionToCurrent(chunkStart))

				s.src.NextByteUnchecked()
				break outer
			case b == '\\':
				str.WriteString(s.src.FromPositionToCurrent(chunkStart))

				continue outer
			default:
				s.ConsumeRune()
				// TODO: report error
				return token.Undetermined
			}
		}
		return token.Undetermined
	}

	escapedString := str.String()
	s.token.escaped = &escapedString
	return token.String
}
