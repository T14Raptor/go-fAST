package parser

// parseOptions holds tunable parser behavior. Its zero value is the default
// behavior, so passing no [Option] to [Parse] leaves parsing unchanged.
type parseOptions struct {
	preserveParens bool
}

// Option customizes parsing behavior. Pass options to [Parse] or [ParseBytes].
type Option func(*parseOptions)

// PreserveParens makes the parser keep grouping parentheses as explicit
// [ast.ParenthesizedExpression] nodes instead of discarding them.
//
// By default `(expr)` parses to the inner expression directly, whose byte
// offsets exclude the parentheses; a parent node spanning such a child can
// therefore report a source range with an unbalanced parenthesis. Opting in
// keeps the parentheses in the tree so every node's [ast.Node.Idx0]/Idx1
// range slices to balanced source.
func PreserveParens() Option {
	return func(o *parseOptions) { o.preserveParens = true }
}
