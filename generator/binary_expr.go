package generator

import "github.com/t14raptor/go-fast/ast"

type binaryExprEntry struct {
	op        string
	rightPrec ast.Precedence
	right     ast.Expr
	wrap      bool
	innerCtx  context
}

// genBinaryExpr linearizes nested binary/logical trees into an iterative
// loop instead of recursing down the left spine.
//
// Note: stack must be a fresh slice per call, not a shared scratch buffer.
// The unwind loop calls g.genExpr(e.right, ...) which can recursively enter
// genBinaryExpr for the right subtree; if both calls aliased the same backing
// store, the inner's `[:0]` + appends would clobber the outer's entries in
// place and the outer's unwind would then read corrupted entries (observed:
// malformed output with a stray `)` on patterns like
// `a && b ? c >> (d & e) : f` where the right subtree is itself binary).
func (g *GenVisitor) genBinaryExpr(expr ast.Expr, minPrec ast.Precedence, ctx context) {
	var stack []binaryExprEntry

descend:
	for {
		var opStr string
		var opPrec, leftPrec, rightPrec ast.Precedence
		var left, right ast.Expr

		switch n := expr.(type) {
		case *ast.BinaryExpression:
			opStr, opPrec = n.Operator.String(), n.Operator.Precedence()
			left, right = n.Left.Expr, n.Right.Expr

			leftPrec, rightPrec = opPrec, opPrec+1
			if opPrec.IsRightAssociative() {
				leftPrec, rightPrec = opPrec+1, opPrec
			}

			// -x ** y is a syntax error; force parens on unary left of **.
			if n.Operator == ast.BinaryExponential {
				if _, ok := left.(*ast.UnaryExpression); ok {
					leftPrec = ast.PrecedenceCall
				}
			}

			// for-init: bare `in` is ambiguous with for-in.
			if n.Operator == ast.BinaryIn && ctx&ctxForbidIn != 0 {
				minPrec = opPrec + 1
			}
		case *ast.LogicalExpression:
			opStr, opPrec = n.Operator.String(), n.Operator.Precedence()
			left, right = n.Left.Expr, n.Right.Expr

			leftPrec, rightPrec = opPrec, opPrec+1

			// Spec forbids mixing ?? with && or || without explicit parens.
			if n.Operator == ast.LogicalCoalesce {
				leftPrec = ast.PrecedenceLogicalAnd + 1
				rightPrec = leftPrec
			}
		default:
			// Leftmost leaf — print it and unwind.
			g.genExpr(expr, minPrec, ctx)
			break descend
		}

		wrap := opPrec < minPrec
		if wrap {
			g.writeByte('(')
		}

		// ctxForbidIn must carry into both subtrees unless this node is
		// wrapped in parens, which delimits the subexpression and makes
		// any inner `in` unambiguous. Non-forbidden context bits are
		// reset at each descent (they're position-specific and don't
		// transit through operators).
		innerCtx := context(0)
		if !wrap {
			innerCtx = ctx & ctxForbidIn
		}

		stack = append(stack, binaryExprEntry{
			op:        opStr,
			rightPrec: rightPrec,
			right:     right,
			wrap:      wrap,
			innerCtx:  innerCtx,
		})

		expr, minPrec, ctx = left, leftPrec, innerCtx
	}

	for i := len(stack) - 1; i >= 0; i-- {
		e := &stack[i]

		if len(e.op) > 2 {
			// Keyword operators (in, instanceof) always need spaces.
			g.writeByte(' ')
			g.writeString(e.op)
			g.writeByte(' ')
		} else {
			g.space()
			g.writeString(e.op)
			g.space()
		}

		g.genExpr(e.right, e.rightPrec, e.innerCtx)

		if e.wrap {
			g.writeByte(')')
		}
	}
}
