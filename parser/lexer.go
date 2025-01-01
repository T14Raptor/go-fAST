package parser

import (
	"errors"
	"fmt"
	"github.com/t14raptor/go-fast/parser/scanner"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

var asciiStart, asciiContinue [128]bool

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

func isDecimalDigit(chr rune) bool {
	return '0' <= chr && chr <= '9'
}

func digitValue(chr rune) int {
	switch {
	case '0' <= chr && chr <= '9':
		return int(chr - '0')
	case 'a' <= chr && chr <= 'f':
		return int(chr - 'a' + 10)
	case 'A' <= chr && chr <= 'F':
		return int(chr - 'A' + 10)
	}
	return 16 // Larger than any legal digit value
}

// 7.2
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

// 7.3
func isLineTerminator(chr rune) bool {
	switch chr {
	case '\u000a', '\u000d', '\u2028', '\u2029':
		return true
	}
	return false
}

type parserState struct {
	c scanner.Checkpoint

	tok scanner.Token

	// TODO errors
}

func (p *parser) mark() parserState {
	return parserState{
		c:   p.scanner.Checkpoint(),
		tok: p.token,
	}
}

func (p *parser) restore(state parserState) {
	p.scanner.Rewind(state.c)
	p.token = state.tok
}

func (p *parser) peek() scanner.Token {
	st := p.mark()
	tok := p.scanner.Next()
	p.restore(st)
	return tok
}

func (p *parser) parseTemplateCharacters() (literal string, parsed string, finished bool, parseErr, err string) {
	offset := p.chrOffset
	var end int
	length := 0
	isUnicode := false
	hasCR := false
	for {
		chr := p.chr
		if chr < 0 {
			goto unterminated
		}
		p.scanner.NextRune()
		if chr == '`' {
			finished = true
			end = p.chrOffset - 1
			break
		}
		if chr == '\\' {
			if p.chr == '\n' || p.chr == '\r' || p.chr == '\u2028' || p.chr == '\u2029' || p.chr < 0 {
				if p.chr == '\r' {
					hasCR = true
				}
				//p.scanNewline()
			} else {
				if p.chr == '8' || p.chr == '9' {
					if parseErr == "" {
						parseErr = "\\8 and \\9 are not allowed in template strings."
					}
				}
				//l, u := p.scanEscape('`')
				//length += l
				//if u {
				//	isUnicode = true
				//}
			}
			continue
		}
		if chr == '$' && p.chr == '{' {
			//p.read()
			end = p.chrOffset - 2
			break
		}
		if chr >= utf8.RuneSelf {
			isUnicode = true
			if chr > 0xFFFF {
				length++
			}
		} else if chr == '\r' {
			hasCR = true
			if p.chr == '\n' {
				length--
			}
		}
		length++
	}
	literal = p.str[offset:end]
	if hasCR {
		literal = normaliseCRLF(literal)
	}
	if parseErr == "" {
		parsed, parseErr = parseStringLiteral(literal, length, isUnicode, true)
	}
	p.insertSemicolon = true
	return
unterminated:
	err = errUnexpectedEndOfInput
	finished = true
	return
}

func normaliseCRLF(s string) string {
	var buf strings.Builder
	buf.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\r' {
			buf.WriteByte('\n')
			if i < len(s)-1 && s[i+1] == '\n' {
				i++
			}
		} else {
			buf.WriteByte(s[i])
		}
	}
	return buf.String()
}

func hex2decimal(chr byte) (value rune, ok bool) {
	{
		chr := rune(chr)
		switch {
		case '0' <= chr && chr <= '9':
			return chr - '0', true
		case 'a' <= chr && chr <= 'f':
			return chr - 'a' + 10, true
		case 'A' <= chr && chr <= 'F':
			return chr - 'A' + 10, true
		}
		return
	}
}

func parseNumberLiteral(literal string) (value float64, err error) {
	// TODO Is Uint okay? What about -MAX_UINT
	n, err := strconv.ParseInt(literal, 0, 64)
	if err == nil {
		return float64(n), nil
	}

	parseIntErr := err // Save this first error, just in case

	value, err = strconv.ParseFloat(literal, 64)
	if err == nil {
		return
	} else if err.(*strconv.NumError).Err == strconv.ErrRange {
		// Infinity, etc.
		return value, nil
	}

	err = parseIntErr

	if err.(*strconv.NumError).Err == strconv.ErrRange {
		if len(literal) > 2 && literal[0] == '0' && (literal[1] == 'X' || literal[1] == 'x') {
			// Could just be a very large number (e.g. 0x8000000000000000)
			var value float64
			literal = literal[2:]
			for _, chr := range literal {
				digit := digitValue(chr)
				if digit >= 16 {
					goto error
				}
				value = value*16 + float64(digit)
			}
			return value, nil
		}
	}

error:
	return 0, errors.New("Illegal numeric literal")
}

func parseStringLiteral(literal string, length int, unicode, strict bool) (string, string) {
	var sb strings.Builder
	var chars []uint16
	if unicode {
		chars = make([]uint16, 1, length+1)
		// BOM
		chars[0] = 0xFEFF
	} else {
		sb.Grow(length)
	}
	str := literal
	for len(str) > 0 {
		switch chr := str[0]; {
		// We do not explicitly handle the case of the quote
		// value, which can be: " ' /
		// This assumes we're already passed a partially well-formed literal
		case chr >= utf8.RuneSelf:
			chr, size := utf8.DecodeRuneInString(str)
			if chr <= 0xFFFF {
				chars = append(chars, uint16(chr))
			} else {
				first, second := utf16.EncodeRune(chr)
				chars = append(chars, uint16(first), uint16(second))
			}
			str = str[size:]
			continue
		case chr != '\\':
			if unicode {
				chars = append(chars, uint16(chr))
			} else {
				sb.WriteByte(chr)
			}
			str = str[1:]
			continue
		}

		if len(str) <= 1 {
			panic("len(str) <= 1")
		}
		chr := str[1]
		var value rune
		if chr >= utf8.RuneSelf {
			str = str[1:]
			var size int
			value, size = utf8.DecodeRuneInString(str)
			str = str[size:] // \ + <character>
			if value == '\u2028' || value == '\u2029' {
				continue
			}
		} else {
			str = str[2:] // \<character>
			switch chr {
			case 'b':
				value = '\b'
			case 'f':
				value = '\f'
			case 'n':
				value = '\n'
			case 'r':
				value = '\r'
			case 't':
				value = '\t'
			case 'v':
				value = '\v'
			case 'x', 'u':
				size := 0
				switch chr {
				case 'x':
					size = 2
				case 'u':
					if str == "" || str[0] != '{' {
						size = 4
					}
				}
				if size > 0 {
					if len(str) < size {
						return "", fmt.Sprintf("invalid escape: \\%s: len(%q) != %d", string(chr), str, size)
					}
					for j := 0; j < size; j++ {
						decimal, ok := hex2decimal(str[j])
						if !ok {
							return "", fmt.Sprintf("invalid escape: \\%s: %q", string(chr), str[:size])
						}
						value = value<<4 | decimal
					}
				} else {
					str = str[1:]
					var val rune
					value = -1
					for ; size < len(str); size++ {
						if str[size] == '}' {
							if size == 0 {
								return "", fmt.Sprintf("invalid escape: \\%s", string(chr))
							}
							size++
							value = val
							break
						}
						decimal, ok := hex2decimal(str[size])
						if !ok {
							return "", fmt.Sprintf("invalid escape: \\%s: %q", string(chr), str[:size+1])
						}
						val = val<<4 | decimal
						if val > utf8.MaxRune {
							return "", fmt.Sprintf("undefined Unicode code-point: %q", str[:size+1])
						}
					}
					if value == -1 {
						return "", fmt.Sprintf("unterminated \\u{: %q", str)
					}
				}
				str = str[size:]
				if chr == 'x' {
					break
				}
				if value > utf8.MaxRune {
					panic("value > utf8.MaxRune")
				}
			case '0':
				if len(str) == 0 || '0' > str[0] || str[0] > '7' {
					value = 0
					break
				}
				fallthrough
			case '1', '2', '3', '4', '5', '6', '7':
				if strict {
					return "", "Octal escape sequences are not allowed in this context"
				}
				value = rune(chr) - '0'
				j := 0
				for ; j < 2; j++ {
					if len(str) < j+1 {
						break
					}
					chr := str[j]
					if '0' > chr || chr > '7' {
						break
					}
					decimal := rune(str[j]) - '0'
					value = (value << 3) | decimal
				}
				str = str[j:]
			case '\\':
				value = '\\'
			case '\'', '"':
				value = rune(chr)
			case '\r':
				if len(str) > 0 {
					if str[0] == '\n' {
						str = str[1:]
					}
				}
				fallthrough
			case '\n':
				continue
			default:
				value = rune(chr)
			}
		}
		if unicode {
			if value <= 0xFFFF {
				chars = append(chars, uint16(value))
			} else {
				first, second := utf16.EncodeRune(value)
				chars = append(chars, uint16(first), uint16(second))
			}
		} else {
			if value >= utf8.RuneSelf {
				return "", "Unexpected unicode character"
			}
			sb.WriteByte(byte(value))
		}
	}

	if unicode {
		if len(chars) != length+1 {
			panic(fmt.Errorf("unexpected unicode length while parsing '%s'", literal))
		}
		return string(utf16.Decode(chars)), ""
	}
	if sb.Len() != length {
		panic(fmt.Errorf("unexpected length while parsing '%s'", literal))
	}
	return sb.String(), ""
}
