package generator

// context carries flags threaded through expression generation that can't be
// expressed by precedence alone.
type context uint8

const (
	ctxForbidIn   context = 1 << iota // for-init position: bare `in` needs parens
	ctxForbidCall                     // new-callee position: bare call needs parens
)
