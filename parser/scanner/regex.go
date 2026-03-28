package scanner

import (
	"github.com/t14raptor/go-fast/token"
	"unicode"
)

// ParseRegExp re-tokenizes the current `/` or `/=` as a regular expression.
func (s *Scanner) ParseRegExp() (string, string, string) {
	offset := s.Offset() - 1 // Opening slash already gotten
	if s.Token.Kind == token.QuotientAssign {
		offset -= 1 // =
	}

	var (
		inEscape    bool
		inCharClass bool
	)
	for {
		chr, ok := s.NextRune()
		if !ok {
			s.error(unterminatedRegExp(s.Token.Idx0, s.src.Offset()))
			break
		}
		if isLineTerminator(chr) {
			s.error(unterminatedRegExp(s.Token.Idx0, s.src.Offset()))
			break
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

	// Pattern is between the slashes (exclusive of closing /)
	// Current position is after closing /, so subtract 1
	pattern := s.src.Slice(offset+1, s.src.Offset()-1)

	// Parse flags with duplicate/invalid flag detection
	flags := ""
	flagStart := s.src.Offset()
	var seenFlags [256]bool
	for {
		c, ok := s.PeekByte()
		if !ok || !isIdentFlagByte(c) {
			break
		}
		s.ConsumeByte()

		// Check for valid regex flags: d, g, i, m, s, u, v, y
		if !isValidRegExpFlag(c) {
			flagOffset := s.src.Offset()
			s.error(regExpFlag(c, flagOffset-1, flagOffset))
		} else if seenFlags[c] {
			flagOffset := s.src.Offset()
			s.error(regExpFlagTwice(c, flagOffset-1, flagOffset))
		}
		seenFlags[c] = true
	}
	if s.src.Offset() > flagStart {
		flags = s.src.FromPositionToCurrent(flagStart)
	}

	literal := s.src.Slice(offset, s.src.Offset())
	return pattern, flags, literal
}

// isIdentFlagByte returns true for bytes that could be regex flag characters.
func isIdentFlagByte(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '$'
}

// isValidRegExpFlag returns true for valid ECMAScript regex flags.
func isValidRegExpFlag(c byte) bool {
	switch c {
	case 'd', 'g', 'i', 'm', 's', 'u', 'v', 'y':
		return true
	}
	return false
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
