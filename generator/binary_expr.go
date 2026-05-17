package generator

import "github.com/t14raptor/go-fast/ast"

type binaryExprEntry struct {
	op        string
	rightPrec ast.Precedence
	right     *ast.Expression
	wrap      bool
	ctx       context
}

// genBinaryExpr linearizes nested binary/logical trees into an iterative
// loop instead of recursing down the left spine.
func (g *GenVisitor) genBinaryExpr(expr *ast.Expression, minPrec ast.Precedence, ctx context) {
	base := len(g.binaryStack)

descend:
	for {
		var opStr string
		var opPrec, leftPrec, rightPrec ast.Precedence
		var left, right *ast.Expression
		var isIn bool

		switch expr.Kind() {
		case ast.ExprBinary:
			n := expr.MustBinary()

			opStr, opPrec = n.Operator.String(), n.Operator.Precedence()
			left, right = n.Left, n.Right
			isIn = n.Operator == ast.BinaryIn

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
			if g.opts.Minified && needsBinaryOperatorSeparator(e.op, e.right) {
				g.writeByte(' ')
			} else {
				g.space()
			}
		}

		g.genExpr(e.right, e.rightPrec, e.ctx)

		if e.wrap {
			g.writeByte(')')
		}
	}
}

func needsBinaryOperatorSeparator(op string, right *ast.Expression) bool {
	switch op {
	case "+":
		if unary, ok := right.Unary(); ok && unary.Operator == ast.UnaryPlus {
			return true
		}
		if update, ok := right.Update(); ok && !update.Postfix && update.Operator == ast.UpdateIncrement {
			return true
		}
	case "-":
		if unary, ok := right.Unary(); ok && unary.Operator == ast.UnaryNegation {
			return true
		}
		if update, ok := right.Update(); ok && !update.Postfix && update.Operator == ast.UpdateDecrement {
			return true
		}
	case "/":
		return startsWithRegExpLiteral(right)
	}
	return false
}

func startsWithRegExpLiteral(expr *ast.Expression) bool {
	if expr == nil {
		return false
	}
	if expr.IsRegExpLit() {
		return true
	}
	if member, ok := expr.Member(); ok {
		return startsWithRegExpLiteral(member.Object)
	}
	if call, ok := expr.Call(); ok {
		return startsWithRegExpLiteral(call.Callee)
	}
	if newExpr, ok := expr.New(); ok {
		return startsWithRegExpLiteral(newExpr.Callee)
	}
	if tmpl, ok := expr.TmplLit(); ok && tmpl.Tag != nil {
		return startsWithRegExpLiteral(tmpl.Tag)
	}
	if priv, ok := expr.PrivDot(); ok {
		return startsWithRegExpLiteral(priv.Left)
	}
	if opt, ok := expr.Optional(); ok {
		return startsWithRegExpLiteral(opt.Expr)
	}
	if chain, ok := expr.OptChain(); ok {
		return startsWithRegExpLiteral(chain.Base)
	}
	return false
}
