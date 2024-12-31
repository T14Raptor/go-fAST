package scanner

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func (s *Scanner) identifierUnicodeEscapeSequence() {
	hasEscape = true
	distance := p.chrOffset - startOffset
	p.read()
	if p.chr != 'u' {
		return "", false, fmt.Sprintf("Invalid identifier escape character: %c (%s)", p.chr, string(p.chr))
	}

	p.read()

	var value rune
	if p.chr == '{' {
		p.read()
		value = -1
		for {
			// If we hit '}' before reading any hex digits, break out
			if p.chr == '}' {
				break
			}
			decimal, ok := hex2decimal(byte(p.chr))
			if !ok {
				return "", false, "Invalid Unicode escape sequence"
			}
			if value == -1 {
				value = decimal
			} else {
				value = (value << 4) | decimal
			}
			// Exceeds max rune?
			if value > utf8.MaxRune {
				return "", false, "Invalid Unicode escape sequence"
			}
			p.read()
		}
		if value == -1 {
			return "", false, "Invalid Unicode escape sequence"
		}
	} else {
		// Classic \uXXXX (4 hex digits).
		decimal, ok := hex2decimal(byte(p.chr))
		if !ok {
			return "", false,
				"Invalid identifier escape character: " + string(p.chr)
		}
		value = decimal
		for i := 0; i < 3; i++ {
			p.read()
			decimal, ok = hex2decimal(byte(p.chr))
			if !ok {
				return "", false, "Invalid identifier escape character: " + string(p.chr)
			}
			value = (value << 4) | decimal
		}
	}
	if value == '\\' {
		return "", false, fmt.Sprintf("Invalid identifier escape value: %c (%s)", value, string(value))
	} else if distance == 0 {
		if !isIdentifierStart(value) {
			return "", false, fmt.Sprintf("Invalid identifier escape value: %c (%s)", value, string(value))
		}
	} else if distance > 0 {
		if !isIdentifierPart(value) {
			return "", false, fmt.Sprintf("Invalid identifier escape value: %c (%s)", value, string(value))
		}
	}
	chr = value
}

func (s *Scanner) stringUnicodeEscapeSequence(str *strings.Builder, isValid *bool) {

}
