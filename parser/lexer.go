package parser

import (
	"errors"
	"math/big"
	"strconv"
	"strings"
)

// isBigIntLiteral reports whether literal is a BigInt numeric literal
// (i.e. ends with the 'n' suffix). The scanner includes the 'n' in the
// token's raw text.
func isBigIntLiteral(literal string) bool {
	return len(literal) > 0 && literal[len(literal)-1] == 'n'
}

// parseBigIntLiteral parses a BigInt numeric literal (including the trailing
// 'n' suffix) into a *big.Int. It supports decimal, hex (0x), octal (0o) and
// binary (0b) bases, as well as numeric separators ('_').
func parseBigIntLiteral(literal string) (*big.Int, error) {
	if !isBigIntLiteral(literal) {
		return nil, errors.New("not a BigInt literal")
	}
	s := literal[:len(literal)-1]
	if strings.ContainsRune(s, '_') {
		s = strings.ReplaceAll(s, "_", "")
	}
	base := 10
	if len(s) >= 2 && s[0] == '0' {
		switch s[1] {
		case 'x', 'X':
			base = 16
			s = s[2:]
		case 'o', 'O':
			base = 8
			s = s[2:]
		case 'b', 'B':
			base = 2
			s = s[2:]
		}
	}
	v, ok := new(big.Int).SetString(s, base)
	if !ok {
		return nil, errors.New("illegal BigInt literal")
	}
	return v, nil
}

func parseNumberLiteral(literal string) (value float64, err error) {
	// TODO Is Uint okay? What about -MAX_UINT
	n, err := strconv.ParseInt(literal, 0, 64)
	if err == nil {
		return float64(n), nil
	}

	parseIntErr := err // Save this first errorf, just in case

	value, err = strconv.ParseFloat(literal, 64)
	if err == nil {
		return
	} else if errors.Is(err, strconv.ErrRange) {
		// Infinity, etc.
		return value, nil
	}

	if errors.Is(parseIntErr, strconv.ErrRange) {
		if len(literal) > 2 && literal[0] == '0' && (literal[1] == 'X' || literal[1] == 'x') {
			// Could just be a very large number (e.g. 0x8000000000000000)
			var value float64
			literal = literal[2:]
			for _, chr := range literal {
				digit := digitValue(chr)
				if digit >= 16 {
					return 0, errors.New("illegal numeric literal")
				}
				value = value*16 + float64(digit)
			}
			return value, nil
		}
	}
	return 0, errors.New("illegal numeric literal")
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
