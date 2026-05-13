package resolver

import "github.com/t14raptor/go-fast/ast"

type DeclKind int

const (
	DeclKindVar DeclKind = iota
	DeclKindFunction
	DeclKindClass
)

type ScopeKind int

const (
	ScopeKindBlock ScopeKind = iota
	ScopeKindFunction
)

type Scope struct {
	parent *Scope

	kind ScopeKind
	ctx  ast.ScopeContext

	declaredSymbols map[string]DeclKind
	shadows         []shadowEntry
}

// shadowEntry records a chainView slot this scope overwrote when it
// first declared name. popScope walks shadows in reverse to restore the
// outer binding (or remove the entry if there was none).
type shadowEntry struct {
	name    string
	prev    ast.ScopeContext
	hadPrev bool
}

func (s *Scope) isDeclared(id string) (DeclKind, bool) {
	for scope := s; scope != nil; scope = scope.parent {
		if scope.declaredSymbols == nil {
			continue
		}
		if k, exists := scope.declaredSymbols[id]; exists {
			return k, true
		}
	}
	return 0, false
}

func getScope(parent *Scope, kind ScopeKind, ctx ast.ScopeContext) *Scope {
	return &Scope{
		parent: parent,
		kind:   kind,
		ctx:    ctx,
	}
}

func putScope(s *Scope) {
	_ = s
}
