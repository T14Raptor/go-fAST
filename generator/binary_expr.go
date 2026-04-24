package generator

import "github.com/t14raptor/go-fast/ast"

type binaryExprEntry struct {
	op        string
	rightPrec ast.Precedence
	right     ast.Expr
	wrap      bool
	ctx       context
}

// genBinaryExpr linearizes nested binary/logical trees into an iterative
// loop instead of recursing down the left spine.
func (g *GenVisitor) genBinaryExpr(expr ast.Expr, minPrec ast.Precedence, ctx context) {
	base := len(g.binaryStack)

descend:
	for {
		var opStr string
		var opPrec, leftPrec, rightPrec ast.Precedence
		var left, right ast.Expr
		var isIn bool

		switch n := expr.(type) {
		case *ast.BinaryExpression:
			opStr, opPrec = n.Operator.String(), n.Operator.Precedence()
			left, right = n.Left.Expr, n.Right.Expr
			isIn = n.Operator == ast.BinaryIn

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

		// Wrap when precedence demands it, or when this is a bare `in`
		// in a for-init context (ambiguous with for-in).
		wrap := opPrec < minPrec || (isIn && ctx&ctxForbidIn != 0)
		if wrap {
			g.writeByte('(')
		}

		// Children inherit forbid-in unless our parens already delimit
		// the for-init subexpression. Other context bits don't cross
		// operators, so they're dropped here.
		if wrap {
			ctx = 0
		} else {
			ctx &= ctxForbidIn
		}

		g.binaryStack = append(g.binaryStack, binaryExprEntry{
			op:        opStr,
			rightPrec: rightPrec,
			right:     right,
			wrap:      wrap,
			ctx:       ctx,
		})

		expr, minPrec = left, leftPrec
	}

	for {
		length := len(g.binaryStack)
		if length == 0 || length-1 < base {
			break
		}
		e := g.binaryStack[length-1]
		g.binaryStack = g.binaryStack[:length-1]

		if e.op == "in" || e.op == "instanceof" {
			// Keyword operators (in, instanceof) always need spaces.
			g.writeByte(' ')
			g.writeString(e.op)
			g.writeByte(' ')
		} else {
			g.space()
			g.writeString(e.op)
			g.space()
		}

		g.genExpr(e.right, e.rightPrec, e.ctx)

		if e.wrap {
			g.writeByte(')')
		}
	}
}
