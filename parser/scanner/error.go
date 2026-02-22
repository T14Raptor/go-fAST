package scanner

import (
	"fmt"
	"github.com/t14raptor/go-fast/ast"
)

type Error struct {
	Message string
	Start   ast.Idx
	End     ast.Idx
}

func (d Error) Error() string {
	return d.Message
}

func invalidCharacter(c rune, start, end ast.Idx) Error {
	return Error{
		Message: fmt.Sprintf("Invalid character `%c`", c),
		Start:   start,
		End:     end,
	}
}

func unexpectedEnd(offset ast.Idx) Error {
	return Error{
		Message: "Unexpected end of file",
		Start:   offset,
		End:     offset,
	}
}

func unterminatedString(start, end ast.Idx) Error {
	return Error{
		Message: "Unterminated string",
		Start:   start,
		End:     end,
	}
}

func unterminatedTemplateLiteral(start, end ast.Idx) Error {
	return Error{
		Message: "Unterminated template literal",
		Start:   start,
		End:     end,
	}
}

func unterminatedMultiLineComment(start, end ast.Idx) Error {
	return Error{
		Message: "Unterminated multiline comment",
		Start:   start,
		End:     end,
	}
}

func unterminatedRegExp(start, end ast.Idx) Error {
	return Error{
		Message: "Unterminated regular expression",
		Start:   start,
		End:     end,
	}
}

func invalidEscapeSequence(start, end ast.Idx) Error {
	return Error{
		Message: "Invalid escape sequence",
		Start:   start,
		End:     end,
	}
}

func invalidNumberEnd(start, end ast.Idx) Error {
	return Error{
		Message: "Invalid characters after number",
		Start:   start,
		End:     end,
	}
}

func invalidUnicodeEscapeSequence(start, end ast.Idx) Error {
	return Error{
		Message: "Invalid Unicode escape sequence",
		Start:   start,
		End:     end,
	}
}

func regExpFlag(c byte, start, end ast.Idx) Error {
	return Error{
		Message: fmt.Sprintf("Invalid regular expression flag `%c`", c),
		Start:   start,
		End:     end,
	}
}

func regExpFlagTwice(c byte, start, end ast.Idx) Error {
	return Error{
		Message: fmt.Sprintf("Duplicate regular expression flag `%c`", c),
		Start:   start,
		End:     end,
	}
}
