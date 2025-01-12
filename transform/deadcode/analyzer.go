package deadcode

import "github.com/t14raptor/go-fast/ast"

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
	curClassId *ast.Id
	curFuncId  *ast.Id
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

	if a.curClassId != nil && *a.curClassId == id {
		return
	}
	if a.curFuncId != nil && *a.curFuncId == id {
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

func (a *Analyzer) VisitAssignExpression(n *ast.AssignExpression) {
	n.VisitChildrenWith(a)

}
