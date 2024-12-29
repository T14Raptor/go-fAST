package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/nukilabs/unicodeid"
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
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

func isDigit(chr rune, base int) bool {
	return digitValue(chr) < base
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

func (p *parser) scanIdentifierTail(startOffset int) (string, bool, string) {
	hasEscape := false
	length := 0
	for chr := p._peek(); isIdentifierPart(chr); chr = p._peek() {
		p.read()

		length++
		if chr == '\\' {
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

		p.read()
	}

	return p.str[startOffset : p.chrOffset+1], hasEscape, ""
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
	idx                                ast.Idx
	tok                                token.Token
	literal                            string
	parsedLiteral                      string
	implicitSemicolon, insertSemicolon bool
	chr                                rune
	chrOffset, offset                  int
	errorCount                         int
}

func (p *parser) mark(state *parserState) *parserState {
	if state == nil {
		state = &parserState{}
	}
	state.idx, state.tok, state.literal, state.parsedLiteral, state.implicitSemicolon, state.insertSemicolon, state.chr, state.chrOffset, state.offset =
		p.idx, p.token, p.literal, p.parsedLiteral, p.implicitSemicolon, p.insertSemicolon, p.chr, p.chrOffset, p.offset

	state.errorCount = len(p.errors)
	return state
}

func (p *parser) restore(state *parserState) {
	p.idx, p.token, p.literal, p.parsedLiteral, p.implicitSemicolon, p.insertSemicolon, p.chr, p.chrOffset, p.offset =
		state.idx, state.tok, state.literal, state.parsedLiteral, state.implicitSemicolon, state.insertSemicolon, state.chr, state.chrOffset, state.offset
	p.errors = p.errors[:state.errorCount]
}

func (p *parser) peek() token.Token {
	implicitSemicolon, insertSemicolon, chr, chrOffset, offset := p.implicitSemicolon, p.insertSemicolon, p.chr, p.chrOffset, p.offset
	tok, _, _ := p.scan()
	p.implicitSemicolon, p.insertSemicolon, p.chr, p.chrOffset, p.offset = implicitSemicolon, insertSemicolon, chr, chrOffset, offset
	return tok
}

func (p *parser) scan() (tkn token.Token, literal string, idx ast.Idx) {
	p.implicitSemicolon = false

	for {
		b, ok := p._peekByte()
		if !ok {
			return token.Eof, "", 0
		}

		fmt.Println("hello", string(b))

		// SAFETY: `byte` is byte value at current position in source
		kind := byteHandlers[b](p)
		if kind != token.Skip {
			fmt.Println(kind, p.literal)
			return kind, p.literal, p.idx
		}
	}
}

func (p *parser) switch2(tkn0, tkn1 token.Token) token.Token {
	if p.chr == '=' {
		p.read()
		return tkn1
	}
	return tkn0
}

func (p *parser) switch3(tkn0, tkn1 token.Token, chr2 rune, tkn2 token.Token) token.Token {
	if p.chr == '=' {
		p.read()
		return tkn1
	}
	if p.chr == chr2 {
		p.read()
		return tkn2
	}
	return tkn0
}

func (p *parser) switch4(tkn0, tkn1 token.Token, chr2 rune, tkn2, tkn3 token.Token) token.Token {
	if p.chr == '=' {
		p.read()
		return tkn1
	}
	if p.chr == chr2 {
		p.read()
		if p.chr == '=' {
			p.read()
			return tkn3
		}
		return tkn2
	}
	return tkn0
}

func (p *parser) switch6(tkn0, tkn1 token.Token, chr2 rune, tkn2, tkn3 token.Token, chr3 rune, tkn4, tkn5 token.Token) token.Token {
	if p.chr == '=' {
		p.read()
		return tkn1
	}
	if p.chr == chr2 {
		p.read()
		if p.chr == '=' {
			p.read()
			return tkn3
		}
		if p.chr == chr3 {
			p.read()
			if p.chr == '=' {
				p.read()
				return tkn5
			}
			return tkn4
		}
		return tkn2
	}
	return tkn0
}

func (p *parser) _peek() rune {
	if p.offset < p.length {
		return rune(p.str[p.offset])
	}
	return -1
}

func (p *parser) peek2Bytes() ([2]byte, bool) {
	if p.offset+1 < p.length {
		return [2]byte{p.str[p.offset], p.str[p.offset+1]}, true
	}
	return [2]byte{}, false
}

func (p *parser) _peekByte() (byte, bool) {
	if p.offset < p.length {
		return p.str[p.offset], true
	}
	return 0, false
}

func (p *parser) advanceIfAsciiEquals(b byte) (matched bool) {
	nextB, ok := p._peekByte()
	if !ok {
		return false
	}
	if matched = nextB == b; matched {
		p.chrOffset = p.offset
		p.offset++
	}
	return matched
}

func (p *parser) read() {
	if p.offset < p.length {
		p.chrOffset = p.offset
		chr, width := rune(p.str[p.offset]), 1
		if chr >= utf8.RuneSelf { // !ASCII
			chr, width = utf8.DecodeRuneInString(p.str[p.offset:])
			if chr == utf8.RuneError && width == 1 {
				p.error("Invalid UTF-8 character")
			}
		}
		p.offset += width
		p.chr = chr
	} else {
		p.chrOffset = p.length
		p.chr = -1 // Eof
	}
}

func (p *parser) skipSingleLineComment() {
	for p.chr != -1 {
		p.read()
		if isLineTerminator(p.chr) {
			return
		}
	}
}

func (p *parser) skipMultiLineComment() (hasLineTerminator bool) {
	p.read()
	for p.chr >= 0 {
		chr := p.chr
		if chr == '\r' || chr == '\n' || chr == '\u2028' || chr == '\u2029' {
			hasLineTerminator = true
			break
		}
		p.read()
		if chr == '*' && p.chr == '/' {
			p.read()
			return
		}
	}
	for p.chr >= 0 {
		chr := p.chr
		p.read()
		if chr == '*' && p.chr == '/' {
			p.read()
			return
		}
	}

	p.errorUnexpected(p.chr)
	return
}

func (p *parser) skipWhiteSpaceCheckLineTerminator() bool {
	for {
		switch p.chr {
		case ' ', '\t', '\f', '\v', '\u00a0', '\ufeff':
			p.read()
			continue
		case '\r':
			if p._peek() == '\n' {
				p.read()
			}
			fallthrough
		case '\u2028', '\u2029', '\n':
			return true
		}
		if p.chr >= utf8.RuneSelf {
			if unicode.IsSpace(p.chr) {
				p.read()
				continue
			}
		}
		break
	}
	return false
}

func (p *parser) skipWhiteSpace() {
	for {
		c := p.chr

		// Fast path for common ASCII whitespace
		if c == ' ' || c == '\t' || c == '\f' || c == '\v' || c == '\u00a0' || c == '\ufeff' {
			p.read()
			continue
		}

		// Handle line terminators
		if c == '\r' {
			// Check if the next character is '\n' without calling p._peek()
			if p.chrOffset < len(p.str) && p.str[p.chrOffset] == '\n' {
				p.read()
			}
			if p.insertSemicolon {
				return
			}
			p.read()
			continue
		}
		if c == '\n' || c == '\u2028' || c == '\u2029' {
			if p.insertSemicolon {
				return
			}
			p.read()
			continue
		}

		// Handle non-ASCII whitespace
		if c >= utf8.RuneSelf && unicode.IsSpace(c) {
			p.read()
			continue
		}

		// If none of the above matched, we're done
		break
	}
}

func (p *parser) scanMantissa(base int) {
	for digitValue(p.chr) < base {
		p.read()
	}
}

func (p *parser) scanEscape(quote rune) (int, bool) {
	var length, base uint32
	chr := p.chr
	switch chr {
	case '0', '1', '2', '3', '4', '5', '6', '7':
		//    Octal:
		length, base = 3, 8
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', '"', '\'':
		p.read()
		return 1, false
	case '\r':
		p.read()
		if p.chr == '\n' {
			p.read()
			return 2, false
		}
		return 1, false
	case '\n':
		p.read()
		return 1, false
	case '\u2028', '\u2029':
		p.read()
		return 1, true
	case 'x':
		p.read()
		length, base = 2, 16
	case 'u':
		p.read()
		if p.chr == '{' {
			p.read()
			length, base = 0, 16
		} else {
			length, base = 4, 16
		}
	default:
		p.read() // Always make progress
	}

	if base > 0 {
		var value uint32
		if length > 0 {
			for ; length > 0 && p.chr != quote && p.chr >= 0; length-- {
				digit := uint32(digitValue(p.chr))
				if digit >= base {
					break
				}
				value = value*base + digit
				p.read()
			}
		} else {
			for p.chr != quote && p.chr >= 0 && value < utf8.MaxRune {
				if p.chr == '}' {
					p.read()
					break
				}
				digit := uint32(digitValue(p.chr))
				if digit >= base {
					break
				}
				value = value*base + digit
				p.read()
			}
		}
		chr = rune(value)
	}
	if chr >= utf8.RuneSelf {
		if chr > 0xFFFF {
			return 2, true
		}
		return 1, true
	}
	return 1, false
}

func (p *parser) scanString(offset int, parse bool) (literal string, parsed string, err string) {
	// " ' /
	quote := rune(p.str[offset])
	length := 0
	isUnicode := false
	for p.chr != quote {
		chr := p.chr
		if chr == '\n' || chr == '\r' || chr < 0 {
			goto newline
		}
		if quote == '/' && (p.chr == '\u2028' || p.chr == '\u2029') {
			goto newline
		}
		p.read()
		if chr == '\\' {
			if p.chr == '\n' || p.chr == '\r' || p.chr == '\u2028' || p.chr == '\u2029' || p.chr < 0 {
				if quote == '/' {
					goto newline
				}
				p.scanNewline()
			} else {
				l, u := p.scanEscape(quote)
				length += l
				if u {
					isUnicode = true
				}
			}
			continue
		} else if chr == '[' && quote == '/' {
			// Allow a slash (/) in a bracket character class ([...])
			// TODO Fix this, this is hacky...
			quote = -1
		} else if chr == ']' && quote == -1 {
			quote = '/'
		}
		if chr >= utf8.RuneSelf {
			isUnicode = true
			if chr > 0xFFFF {
				length++
			}
		}
		length++
	}

	// " ' /
	p.read()
	literal = p.str[offset:p.chrOffset]
	if parse {
		// TODO strict
		parsed, err = parseStringLiteral(literal[1:len(literal)-1], length, isUnicode, false)
	}
	return

newline:
	p.scanNewline()
	errStr := "String not terminated"
	if quote == '/' {
		errStr = "Invalid regular expression: missing /"
		p.error(errStr)
	}
	return "", "", errStr
}

func (p *parser) scanNewline() {
	if p.chr == '\u2028' || p.chr == '\u2029' {
		p.read()
		return
	}
	if p.chr == '\r' {
		p.read()
		if p.chr != '\n' {
			return
		}
	}
	p.read()
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
		p.read()
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
				p.scanNewline()
			} else {
				if p.chr == '8' || p.chr == '9' {
					if parseErr == "" {
						parseErr = "\\8 and \\9 are not allowed in template strings."
					}
				}
				l, u := p.scanEscape('`')
				length += l
				if u {
					isUnicode = true
				}
			}
			continue
		}
		if chr == '$' && p.chr == '{' {
			p.read()
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

func (p *parser) readZero() token.Token {
	offset := p.offset
	p.read()
	base := 0
	switch p.chr {
	case 'x', 'X':
		base = 16
	case 'o', 'O':
		base = 8
	case 'b', 'B':
		base = 2
	case '.':
		p.read()
		p.scanMantissa(10)
	case 'e', 'E':
		p.read()
		if p.chr == '-' || p.chr == '+' {
			p.read()
		}
		if isDecimalDigit(p.chr) {
			p.read()
			p.scanMantissa(10)
		} else {
			p.literal = p.str[offset:p.chrOffset]
			return token.Illegal
		}
	default:
		// legacy octal
		p.scanMantissa(8)
	}
	if base > 0 {
		p.read()
		if !isDigit(p.chr, base) {
			p.literal = p.str[offset:p.chrOffset]
			return token.Illegal
		}
		p.scanMantissa(base)
	}
	p.literal = p.str[offset:p.chrOffset]
	return token.Number
}

func (p *parser) scanNumericLiteral(decimalPoint bool) (token.Token, string) {
	offset := p.chrOffset
	tkn := token.Number

	if decimalPoint {
		offset--
		p.scanMantissa(10)
	} else {
		if p.chr == '0' {
			p.read()
			base := 0
			switch p.chr {
			case 'x', 'X':
				base = 16
			case 'o', 'O':
				base = 8
			case 'b', 'B':
				base = 2
			case '.', 'e', 'E':
				// no-op
			default:
				// legacy octal
				p.scanMantissa(8)
				goto end
			}
			if base > 0 {
				p.read()
				if !isDigit(p.chr, base) {
					return token.Illegal, p.str[offset:p.chrOffset]
				}
				p.scanMantissa(base)
				goto end
			}
		} else {
			p.scanMantissa(10)
		}
		if p.chr == '.' {
			p.read()
			p.scanMantissa(10)
		}
	}

	if p.chr == 'e' || p.chr == 'E' {
		p.read()
		if p.chr == '-' || p.chr == '+' {
			p.read()
		}
		if isDecimalDigit(p.chr) {
			p.read()
			p.scanMantissa(10)
		} else {
			return token.Illegal, p.str[offset:p.chrOffset]
		}
	}
end:
	if isIdentifierStart(p.chr) || isDecimalDigit(p.chr) {
		return token.Illegal, p.str[offset:p.chrOffset]
	}

	return tkn, p.str[offset:p.chrOffset]
}
