package parser

import (
	"github.com/t14raptor/go-fast/unistring"
)

type scope struct {
	outer        *scope
	allowIn      bool
	allowLet     bool
	inIteration  bool
	inSwitch     bool
	inFuncParams bool
	inFunction   bool
	inAsync      bool
	allowAwait   bool
	allowYield   bool

	labels []unistring.String
}

func (p *parser) openScope() {
	p.scope = &scope{
		outer:   p.scope,
		allowIn: true,
	}
}

func (p *parser) closeScope() {
	p.scope = p.scope.outer
}

func (s *scope) hasLabel(name unistring.String) bool {
	for _, label := range s.labels {
		if label == name {
			return true
		}
	}
	if s.outer != nil && !s.inFunction {
		return s.outer.hasLabel(name)
	}
	return false
}
