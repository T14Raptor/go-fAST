package deadcode

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
	"github.com/t14raptor/go-fast/transform/utils"
)

type ScopeKind int

const (
	ScopeKindFunction ScopeKind = iota
	ScopeKindArrowFunction
)

type Scope struct {
	parent *Scope
	kind   ScopeKind

	bindingsAffectedByEval map[ast.Id]struct{}
	foundDirectEval        bool

	bindingsAffectedByArguments map[ast.Id]struct{}
	foundArguments              bool

	// Used to construct a graph
	// This includes all bindings to current node
	astPath []ast.Id
}

func (s *Scope) IsAstEmptyPath() bool {
	if len(s.astPath) != 0 {
		return false
	}
	if s.parent != nil {
		return s.parent.IsAstEmptyPath()
	}
	return true
}

type Analyzer struct {
	ast.NoopVisitor

	inVarDecl  bool
	scope      *Scope
	data       *Data
	curClassId ast.Id
	curFuncId  ast.Id
}

func (a *Analyzer) WithAstPath(ids []ast.Id, op func(*Analyzer)) {
	prevLen := len(a.scope.astPath)
	a.scope.astPath = append(a.scope.astPath, ids...)
	op(a)
	a.scope.astPath = a.scope.astPath[:prevLen]
}

func (a *Analyzer) WithScope(kind ScopeKind, op func(*Analyzer)) {
	v := &Analyzer{
		scope: &Scope{
			parent: a.scope.parent,
		},
		data:       a.data,
		curClassId: a.curClassId,
		curFuncId:  a.curFuncId,
	}
	v.V = v
	op(v)
	child := v.scope
	child.parent = nil

	// If we found eval, mark all declarations in scope and upper as used
	if v.scope.foundDirectEval {
		for id := range child.bindingsAffectedByEval {
			if name, ok := a.data.usedNames[id]; !ok {
				a.data.usedNames[id] = VarInfo{Usage: 1}
			} else {
				name.Usage++
				a.data.usedNames[id] = name
			}
		}
		a.scope.foundDirectEval = true
	}

	// If we found arguments, mark all declarations in scope and upper as used
	if child.foundArguments {
		for id := range child.bindingsAffectedByArguments {
			if name, ok := a.data.usedNames[id]; !ok {
				a.data.usedNames[id] = VarInfo{Usage: 1}
			} else {
				name.Usage++
				a.data.usedNames[id] = name
			}
		}
		if kind == ScopeKindFunction {
			a.scope.foundArguments = true
		}
	}
}

func (a *Analyzer) Add(id ast.Id, assign bool) {
	if id.Name == "arguments" {
		a.scope.foundArguments = true
	}

	if a.curClassId == id {
		return
	}
	if a.curFuncId == id {
		return
	}

	if a.scope.IsAstEmptyPath() {
		// Add references from top level items into graph
		a.data.entries[id] = struct{}{}
	} else {
		for scope := a.scope; scope != nil; scope = scope.parent {
			for _, component := range scope.astPath {
				a.data.AddDependencyEdge(component, id, assign)
			}

			if scope.kind == ScopeKindFunction && scope.astPath != nil {
				break
			}
		}
	}

	if assign {
		if info, ok := a.data.usedNames[id]; !ok {
			a.data.usedNames[id] = VarInfo{Assign: 1}
		} else {
			info.Assign++
			a.data.usedNames[id] = info
		}
	} else {
		if info, ok := a.data.usedNames[id]; !ok {
			a.data.usedNames[id] = VarInfo{Usage: 1}
		} else {
			info.Usage++
			a.data.usedNames[id] = info
		}
	}
}

func (a *Analyzer) VisitCallExpression(n *ast.CallExpression) {
	n.VisitChildrenWith(a)

	if ident, ok := n.Callee.Expr.(*ast.Identifier); ok {
		if ident.Name == "eval" {
			a.scope.foundDirectEval = true
		}
	}
}

func (a *Analyzer) VisitClassDeclaration(n *ast.ClassDeclaration) {
	a.WithAstPath([]ast.Id{n.Class.Name.ToId()}, func(v *Analyzer) {
		old := v.curClassId
		v.curClassId = n.Class.Name.ToId()
		n.VisitChildrenWith(v)
		v.curClassId = old
	})
}

func (a *Analyzer) VisitExpression(n *ast.Expression) {
	old := a.inVarDecl
	a.inVarDecl = false
	n.VisitChildrenWith(a)
	if ident, ok := n.Expr.(*ast.Identifier); ok {
		a.Add(ident.ToId(), false)
	}
	a.inVarDecl = old
}

func (a *Analyzer) VisitAssignExpression(n *ast.AssignExpression) {
	if ident, ok := n.Left.Expr.(*ast.Identifier); ok && n.Operator == token.Assign {
		a.Add(ident.ToId(), true)
		n.Right.VisitWith(a)
	} else {
		n.VisitChildrenWith(a)
	}
}

func (a *Analyzer) VisitArrowFunctionLiteral(n *ast.ArrowFunctionLiteral) {
	a.WithScope(ScopeKindArrowFunction, func(v *Analyzer) {
		n.VisitChildrenWith(v)

		if v.scope.foundDirectEval {
			v.scope.bindingsAffectedByEval = utils.CollectDeclarations(n)
		}
	})
}

func (a *Analyzer) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	a.WithScope(ScopeKindFunction, func(v *Analyzer) {
		n.VisitChildrenWith(v)

		if v.scope.foundDirectEval {
			v.scope.bindingsAffectedByEval = utils.CollectDeclarations(n)
		}

		if v.scope.foundArguments {
			v.scope.bindingsAffectedByArguments = utils.CollectIdentifiers(&n.ParameterList)
		}
	})
}

func (a *Analyzer) VisitFunctionDeclaration(n *ast.FunctionDeclaration) {
	a.WithAstPath([]ast.Id{n.Function.Name.ToId()}, func(v *Analyzer) {
		old := v.curFuncId
		v.curFuncId = n.Function.Name.ToId()
		n.VisitChildrenWith(v)
		v.curFuncId = old
	})
}

func (a *Analyzer) VisitBindingTarget(n *ast.BindingTarget) {
	n.VisitChildrenWith(a)
	if !a.inVarDecl {
		if ident, ok := n.Target.(*ast.Identifier); ok {
			a.Add(ident.ToId(), true)
		}
	}
}

func (a *Analyzer) VisitProperty(n *ast.Property) {
	if short, ok := n.Prop.(*ast.PropertyShort); ok {
		a.Add(short.Name.ToId(), false)
	}
}

func (a *Analyzer) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	old := a.inVarDecl
	for _, decl := range n.List {
		a.inVarDecl = true
		decl.Target.VisitWith(a)
		a.inVarDecl = false
		if decl.Initializer != nil {
			decl.Initializer.VisitWith(a)
		}
	}
	a.inVarDecl = old
}
