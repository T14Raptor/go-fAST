package generator

import (
	"strings"

	"github.com/t14raptor/go-fast/ast"
)

type state struct {
	out    *strings.Builder
	node   ast.Node
	parent *state
	indent int
}

func (s *state) wrap(node ast.Node) *state {
	return &state{
		out:    s.out,
		node:   node,
		parent: s,
		indent: s.indent,
	}
}

func (s *state) line() {
	s.out.WriteString("\n")
}

func (s *state) lineAndPad() {
	s.line()
	s.out.WriteString(strings.Repeat("    ", s.indent))
}
