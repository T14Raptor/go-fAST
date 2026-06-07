package ast

// ScopeContext is the per-scope identity attached to every resolved
// [Identifier]. The resolver assigns each new lexical scope a unique value;
// two identifiers share a context if they refer to the same binding.
//
// Two values are reserved:
//   - [UnresolvedContext] (zero) means "not yet resolved" or "free reference".
//   - [TopLevelContext]   (one)  is the program-scope context.
type ScopeContext int32

const (
	UnresolvedContext ScopeContext = 0
	TopLevelContext   ScopeContext = 1
)
