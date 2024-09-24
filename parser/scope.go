package parser

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

	labels []string
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

func (s *scope) hasLabel(name string) bool {
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
