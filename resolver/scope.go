package resolver

import "github.com/t14raptor/go-fast/ast"

type (
	scopeKind int

	declKind int

	identType int
)

const (
	scopeKindBlock    scopeKind = 0
	scopeKindFunction scopeKind = 1

	declKindVar      declKind = 0
	declKindFunction declKind = 1

	identTypeRef     identType = 0 // Reference (read)
	identTypeBinding identType = 1 // Binding (declaration)
)

type scope struct {
	parent *scope

	kind scopeKind

	ctx ast.ScopeContext

	declaredSymbols map[string]declKind
}

func (s *scope) isDeclared(id string) (declKind, bool) {
	for sc := s; sc != nil; sc = sc.parent {
		if kind, exists := sc.declaredSymbols[id]; exists {
			return kind, true
		}
	}
	return 0, false
}
