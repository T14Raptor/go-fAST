package resolver

import "github.com/t14raptor/go-fast/ast"

type (
	// declKind is a byte because it is stored in every symEntry and trailEntry,
	// where its width drives the resolver's peak memory.
	declKind byte

	identType byte
)

const (
	declKindVar      declKind = 0
	declKindFunction declKind = 1
	declKindClass    declKind = 2

	identTypeRef     identType = 0 // reference (read)
	identTypeBinding identType = 1 // binding (declaration)
)

// symEntry is a name's innermost visible binding in resolver.symbols.
type symEntry struct {
	ctx  ast.ScopeContext
	kind declKind
}

// trailEntry records the binding a declaration shadowed, so popScope can restore it.
type trailEntry struct {
	name string
	prev symEntry
	had  bool
}

// scope is one lexical scope. Scopes are stack-disciplined and unreferenced once
// popped, so the resolver recycles them through a free list. Bindings live in a
// single global table (resolver.symbols) rather than a per-scope map; each scope
// only marks where its declarations begin in the undo trail, making a reference
// lookup one O(1) probe instead of a walk up a chain of maps.
type scope struct {
	parent *scope

	ctx ast.ScopeContext

	// trailBase indexes resolver.trail where this scope's declarations begin;
	// popScope rewinds to here to restore shadowed bindings.
	trailBase int
}
