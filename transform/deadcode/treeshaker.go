package deadcode

import (
	"slices"
	"sync/atomic"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/ast/ext"
	"github.com/t14raptor/go-fast/token"
	"github.com/t14raptor/go-fast/tools/fastgraph"
	"github.com/t14raptor/go-fast/transform/resolver"
	"github.com/t14raptor/go-fast/transform/utils"
)

type TreeShaker struct {
	ast.NoopVisitor
	changed bool

	data     Data
	bindings map[ast.Id]struct{}

	remove atomic.Bool
}

func (ts *TreeShaker) CanDropBinding(id ast.Id, isVar bool) bool {
	if name, ok := ts.data.usedNames[id]; ok {
		return name.Usage == 0 && name.Assign == 0
	}
	return true
}

func (ts *TreeShaker) CanDropAssignmentTo(id ast.Id, isVar bool) bool {
	if _, ok := ts.bindings[id]; ok {
		if name, ok := ts.data.usedNames[id]; ok {
			return name.Usage == 0
		}
	}
	return false
}

func (ts *TreeShaker) VisitExpressions(n *ast.Expressions) {
	for i := len(*n) - 1; i >= 0; i-- {
		(*n)[i].VisitWith(ts)

		if ts.remove.CompareAndSwap(true, false) {
			*n = slices.Delete(*n, i, i+1)
		}
	}
}

func (ts *TreeShaker) VisitStatements(n *ast.Statements) {
	for i := len(*n) - 1; i >= 0; i-- {
		(*n)[i].VisitWith(ts)

		if ts.remove.CompareAndSwap(true, false) {
			*n = slices.Delete(*n, i, i+1)
			continue
		}

		switch stmt := (*n)[i].Stmt.(type) {
		case *ast.EmptyStatement:
			*n = slices.Delete(*n, i, i+1)
		case *ast.BlockStatement:
			if len(stmt.List) == 0 {
				*n = slices.Delete(*n, i, i+1)
			}
		}
	}
}

func (ts *TreeShaker) VisitAssignExpression(n *ast.AssignExpression) {
	n.VisitChildrenWith(ts)

	if ident, ok := n.Left.Expr.(*ast.Identifier); ok {
		if ts.CanDropAssignmentTo(ident.ToId(), false) && !ext.MayHaveSideEffects(n.Right) {
			ts.changed = true
			ts.remove.Store(true)
		}
	}
}

func (ts *TreeShaker) VisitFunctionDeclaration(n *ast.FunctionDeclaration) {
	n.VisitChildrenWith(ts)

	if ts.CanDropBinding(n.Function.Name.ToId(), true) {
		ts.changed = true
		ts.remove.Store(true)
	}
}

func (ts *TreeShaker) VisitClassDeclaration(n *ast.ClassDeclaration) {
	n.VisitChildrenWith(ts)

	if ts.CanDropBinding(n.Class.Name.ToId(), false) {
		if n.Class.SuperClass != nil && ext.MayHaveSideEffects(n.Class.SuperClass) {
			return
		}

		if slices.ContainsFunc(n.Class.Body, func(elem ast.ClassElement) bool {
			switch elem := elem.Element.(type) {
			case *ast.MethodDefinition:
				return elem.Computed
			case *ast.FieldDefinition:
				return elem.Computed || (elem.Initializer != nil && ext.MayHaveSideEffects(elem.Initializer))
			case *ast.ClassStaticBlock:
				return true
			default:
				return false
			}
		}) {
			return
		}

		ts.changed = true
		ts.remove.Store(true)
	}
}

func (ts *TreeShaker) VisitExpression(n *ast.Expression) {
	n.VisitChildrenWith(ts)

	switch expr := n.Expr.(type) {
	case *ast.BinaryExpression:
		switch expr.Operator {
		case token.LogicalAnd:
			if val := ext.AsPureBool(expr.Left); val.Known() && !val.Val() {
				n.Expr = expr.Left.Expr
				ts.changed = true
			}
		case token.LogicalOr:
			if val := ext.AsPureBool(expr.Left); val.Known() && val.Val() {
				n.Expr = expr.Left.Expr
				ts.changed = true
			}
		}
	}
}

func (ts *TreeShaker) VisitStatement(n *ast.Statement) {
	n.VisitChildrenWith(ts)

	if varDecl, ok := n.Stmt.(*ast.VariableDeclaration); ok {
		if len(varDecl.List) == 0 {
			ts.remove.Store(true)
		} else {
			// If all name is droppable, do so.
			if slices.ContainsFunc(varDecl.List, func(v ast.VariableDeclarator) bool {
				if ident, ok := v.Target.Target.(*ast.Identifier); ok {
					return !ts.CanDropBinding(ident.ToId(), varDecl.Token == token.Var)
				}
				return true
			}) {
				return
			}

			var exprs []ast.Expression
			for _, v := range varDecl.List {
				if v.Initializer != nil {
					exprs = append(exprs, *v.Initializer)
				}
			}

			if len(exprs) == 0 {
				n.Stmt = &ast.EmptyStatement{}
			} else if len(exprs) == 1 {
				n.Stmt = &ast.ExpressionStatement{Expression: &exprs[0]}
			} else {
				n.Stmt = &ast.ExpressionStatement{Expression: &ast.Expression{
					Expr: &ast.SequenceExpression{Sequence: exprs},
				}}
			}
		}
	}
}

func (ts *TreeShaker) VisitUnaryExpression(n *ast.UnaryExpression) {
	if n.Operator == token.Delete {
		return
	}
	n.VisitChildrenWith(ts)
}

func (ts *TreeShaker) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	for i := len(n.List) - 1; i >= 0; i-- {
		if ident, ok := n.List[i].Target.Target.(*ast.Identifier); ok {
			canDrop := true
			if n.List[i].Initializer != nil {
				canDrop = !ext.MayHaveSideEffects(n.List[i].Initializer)
			}
			if canDrop && ts.CanDropBinding(ident.ToId(), n.Token == token.Var) {
				ts.changed = true
				n.List = slices.Delete(n.List, i, i+1)
			}
		}
	}
}

func (ts *TreeShaker) VisitProgram(n *ast.Program) {
	if len(ts.bindings) == 0 {
		ts.bindings = utils.CollectDeclarations(n)
	}

	data := Data{
		usedNames: make(map[ast.Id]VarInfo),
		graph:     fastgraph.New[ast.Id, VarInfo](),
		entries:   make(map[ast.Id]struct{}),
	}

	analyzer := &Analyzer{
		data:  &data,
		scope: &Scope{},
	}
	analyzer.V = analyzer
	n.VisitWith(analyzer)

	data.SubtractCycles()
	ts.data = data

	n.VisitChildrenWith(ts)
}

func Eliminate(p *ast.Program, resolve bool) {
	if resolve {
		resolver.Resolve(p)
	}

	treeshaker := &TreeShaker{changed: true}
	treeshaker.V = treeshaker
	for treeshaker.changed {
		treeshaker.changed = false
		p.VisitWith(treeshaker)
	}
}
