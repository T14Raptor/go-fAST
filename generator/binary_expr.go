package generator

import "github.com/t14raptor/go-fast/ast"

type binaryExprEntry struct {
	op        string
	rightPrec ast.Precedence
	right     *ast.Expression
	wrap      bool
}

// genBinaryExpr linearizes nested binary/logical trees into an iterative
// loop instead of recursing down the left spine.
func (g *GenVisitor) genBinaryExpr(expr *ast.Expression, minPrec ast.Precedence, ctx context) {
	stack := g.binaryStack[:0]

descend:
	for {
		var opStr string
		var opPrec, leftPrec, rightPrec ast.Precedence
		var left, right *ast.Expression

		switch expr.Kind() {
		case ast.ExprBinary:
			n := expr.MustBinary()

			opStr, opPrec = n.Operator.String(), n.Operator.Precedence()
			left, right = n.Left, n.Right

			leftPrec, rightPrec = opPrec, opPrec+1
			if opPrec.IsRightAssociative() {
				leftPrec, rightPrec = opPrec+1, opPrec
			}

			// -x ** y is a syntax error; force parens on unary left of **.
			if n.Operator == ast.BinaryExponential {
				if left.IsUnary() {
					leftPrec = ast.PrecedenceCall
				}
			}

			// for-init: bare `in` is ambiguous with for-in.
			if n.Operator == ast.BinaryIn && ctx&ctxForbidIn != 0 {
				minPrec = opPrec + 1
			}
		case ast.ExprLogical:
			n := expr.MustLogical()

			opStr, opPrec = n.Operator.String(), n.Operator.Precedence()
			left, right = n.Left, n.Right

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

		stack = append(stack, binaryExprEntry{
			op:        opStr,
			rightPrec: rightPrec,
			right:     right,
			wrap:      wrap,
		})

		expr, minPrec, ctx = left, leftPrec, 0
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

		g.genExpr(e.right, e.rightPrec, 0)

		if e.wrap {
			g.writeByte(')')
		}
	}

	g.binaryStack = stack[:0]
}
