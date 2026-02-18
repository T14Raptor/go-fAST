package scanner

import (
	"errors"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

type Scanner struct {
	Token Token

	src Source

	EscapedStr string // escape-processed string for current token

	errors *error
}

func NewScanner(src string, errors *error) *Scanner {
	return &Scanner{
		src: NewSource(src),

		errors: errors,
	}
}

func (s *Scanner) error(d Error) {
	*s.errors = errors.Join(*s.errors, d)
}

// unexpectedErr reports an unexpected character/end error at the current position.
func (s *Scanner) unexpectedErr() {
	if s.src.EOF() {
		s.error(unexpectedEnd(s.src.Offset()))
	} else {
		b := s.src.PeekByteUnchecked()
		start := s.src.Offset()
		s.src.pos++
		s.error(invalidCharacter(rune(b), start, s.src.Offset()))
	}
}

// unterminatedRange returns the range from the current Token start to the current position.
func (s *Scanner) unterminatedRange() (ast.Idx, ast.Idx) {
	return s.Token.Idx0, s.src.Offset()
}

type Checkpoint struct {
	pos        ast.Idx
	tok        Token
	escapedStr string
	errors     error
}

func (s *Scanner) Checkpoint() Checkpoint {
	return Checkpoint{
		pos:        s.src.pos,
		tok:        s.Token,
		escapedStr: s.EscapedStr,
		errors:     *s.errors,
	}
}

func (s *Scanner) Rewind(c Checkpoint) {
	s.src.pos = c.pos
	s.Token = c.tok
	s.EscapedStr = c.escapedStr
	*s.errors = c.errors
}

func (s *Scanner) Offset() ast.Idx {
	return s.src.Offset()
}

func (s *Scanner) NextRune() (rune, bool) {
	return s.src.NextRune()
}

func (s *Scanner) NextByte() (byte, bool) {
	return s.src.NextByte()
}

func (s *Scanner) ConsumeRune() rune {
	r, _ := s.src.NextRune()
	return r
}

func (s *Scanner) ConsumeByte() byte {
	return s.src.NextByteUnchecked()
}

func (s *Scanner) PeekRune() (rune, bool) {
	return s.src.PeekRune()
}

func (s *Scanner) PeekByte() (byte, bool) {
	return s.src.PeekByte()
}

func (s *Scanner) AdvanceIfByteEquals(b byte) bool {
	return s.src.AdvanceIfByteEquals(b)
}

func (s *Scanner) NextTemplatePart() Token {
	s.Token.Idx0 = s.src.Offset() - 1
	s.Token.Kind = s.ReadTemplateLiteral(token.TemplateMiddle, token.TemplateTail)
	s.Token.Idx1 = s.src.Offset()
	return s.Token
}
