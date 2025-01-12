package deadcode

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/ast/ext"
	"github.com/t14raptor/go-fast/token"
)

type TreeShaker struct {
	ast.NoopVisitor

	changed bool
	pass    int

	inFunc      bool
	inBlockStmt bool
	varDeclKind token.Token

	data Data

	bindings map[ast.Id]struct{}
}

func (ts *TreeShaker) CanDropBinding(id ast.Id, isVar bool) bool {
	if isVar {
		if !ts.inFunc {
			return false
		}
	} else if !ts.inBlockStmt {
		return false
	}

	if name, ok := ts.data.usedNames[id]; ok {
		return name.Usage == 0 && name.Assign == 0
	}
	return true
}

func (ts *TreeShaker) CanDropAssignmentTo(id ast.Id, isVar bool) bool {
	if isVar {
		if !ts.inFunc {
			return false
		}
	} else if !ts.inBlockStmt {
		return false
	}

	// Abort if the variable is declared on top level scope.
	if _, ok := ts.data.entries[id]; ok {
		return false
	}

	if _, ok := ts.bindings[id]; ok {
		if name, ok := ts.data.usedNames[id]; ok {
			return name.Usage == 0
		}
	}
	return false
}

func (ts *TreeShaker) OptimizeBinaryExpression(n *ast.Expression) {
	binExpr, ok := n.Expr.(*ast.BinaryExpression)
	if !ok {
		return
	}

	switch binExpr.Operator {
	case token.LogicalAnd:
		if val := ext.AsPureBool(binExpr.Left); val.Known() && !val.Val() {
			n.Expr = binExpr.Left.Expr
			ts.changed = true
		}
	case token.LogicalOr:
		if val := ext.AsPureBool(binExpr.Left); val.Known() && val.Val() {
			n.Expr = binExpr.Left.Expr
			ts.changed = true
		}
	}
}
