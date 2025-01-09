package scanner

import (
	"github.com/t14raptor/go-fast/token"
	"unicode"
)

func (s *Scanner) ParseRegExp() (string, string, string) {
	offset := s.Offset() - 1 // Opening slash already gotten
	if s.token.Kind == token.QuotientAssign {
		offset -= 1 // =
	}

	var (
		inEscape    bool
		inCharClass bool
		chr         rune
	)
	for {
		chr, ok := s.NextRune()
		if !ok || isLineTerminator(chr) {
			//p.error(errUnexpectedEndOfInput)
			return "", "", ""
		}

		if inEscape {
			inEscape = false
		} else if chr == '/' && !inCharClass {
			break
		} else if chr == '[' {
			inCharClass = true
		} else if chr == '\\' {
			inEscape = true
		} else if chr == ']' {
			inCharClass = false
		}
	}

	pattern := s.src.FromPositionToCurrent(offset + 1)

	flags := ""
	if !isLineTerminator(chr) && !isLineWhiteSpace(chr) {
		s.Next()

		if s.token.Kind == token.Identifier { // gim
			flags = s.token.String(s)

			s.Next()
		}
	} else {
		s.NextRune()
	}

	return pattern, flags, s.src.FromPositionToCurrent(offset)
}

func isLineWhiteSpace(chr rune) bool {
	switch chr {
	case '\u0009', '\u000b', '\u000c', '\u0020', '\u00a0', '\ufeff':
		return true
	case '\u000a', '\u000d', '\u2028', '\u2029':
		return false
	case '\u0085':
		return false
	}
	return unicode.IsSpace(chr)
}
