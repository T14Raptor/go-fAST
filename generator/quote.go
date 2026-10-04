package generator

import (
	"unicode"
	"unicode/utf8"
)

const hexDigits = "0123456789abcdef"

// writeQuoted writes s as a double-quoted JavaScript string literal. Printable
// characters are written as they are; the quote, the backslash, line
// terminators and other control or non-printable characters are escaped.
// strconv.Quote is not a JavaScript quoter: its `\a` reads as "a" and its
// `\U0001f600` doesn't parse.
func (g *GenVisitor) writeQuoted(s string) {
	g.writeByte('"')
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			// A byte that is not UTF-8 has no JavaScript spelling; \xHH is
			// the nearest one.
			g.buf = appendHexEscape(g.buf, 'x', uint32(s[i]), 2)
			i++
			continue
		}
		i += size

		switch r {
		case '"':
			g.buf = append(g.buf, '\\', '"')
		case '\\':
			g.buf = append(g.buf, '\\', '\\')
		case '\n':
			g.buf = append(g.buf, '\\', 'n')
		case '\r':
			g.buf = append(g.buf, '\\', 'r')
		case '\t':
			g.buf = append(g.buf, '\\', 't')
		case '\b':
			g.buf = append(g.buf, '\\', 'b')
		case '\f':
			g.buf = append(g.buf, '\\', 'f')
		case '\v':
			g.buf = append(g.buf, '\\', 'v')
		default:
			switch {
			case r < 0x20 || r == 0x7f:
				g.buf = appendHexEscape(g.buf, 'x', uint32(r), 2)
			case unicode.IsPrint(r):
				g.buf = utf8.AppendRune(g.buf, r)
			case r < 0x10000:
				g.buf = appendHexEscape(g.buf, 'u', uint32(r), 4)
			default:
				// Outside the BMP, as its surrogate pair.
				r -= 0x10000
				g.buf = appendHexEscape(g.buf, 'u', uint32(0xd800+(r>>10)), 4)
				g.buf = appendHexEscape(g.buf, 'u', uint32(0xdc00+(r&0x3ff)), 4)
			}
		}
	}
	g.buf = append(g.buf, '"')
}

// appendHexEscape appends \x or \u (kind) and v in digits hex digits.
func appendHexEscape(buf []byte, kind byte, v uint32, digits int) []byte {
	buf = append(buf, '\\', kind)
	for i := digits - 1; i >= 0; i-- {
		buf = append(buf, hexDigits[(v>>(4*i))&0xf])
	}
	return buf
}
