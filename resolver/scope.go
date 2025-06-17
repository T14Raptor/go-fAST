package resolver

import "github.com/t14raptor/go-fast/ast"

type DeclKind int

const (
	DeclKindVar DeclKind = iota
	DeclKindFunction
)

type ScopeKind int

const (
	ScopeKindBlock ScopeKind = iota
	ScopeKindFunction
)

type Scope struct {
	parent *Scope

	kind ScopeKind

	ctx ast.ScopeContext

	declaredSymbols map[string]DeclKind
}

func (s *Scope) isDeclared(id string) (DeclKind, bool) {
	for scope := s; scope != nil; scope = scope.parent {
		if declKind, exists := scope.declaredSymbols[id]; exists {
			return declKind, true
		}
	}
	return 0, false
}
