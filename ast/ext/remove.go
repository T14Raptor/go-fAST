package ext

import "github.com/t14raptor/go-fast/ast"

// RemoveHelper is a visitor that assists in removing specific nodes from ASTs during traversal.
//
// If you override a Visit method that has deletion logic:
//   - [RemoveHelper.VisitStatements]
//   - [RemoveHelper.VisitExpressions]
//   - [RemoveHelper.VisitSequenceExpression]
//   - [RemoveHelper.VisitVariableDeclarators]
//   - [RemoveHelper.VisitVariableDeclaration]
//   - [RemoveHelper.VisitClassElements]
//   - [RemoveHelper.VisitProperties]
//
// make sure to either call the base implementation or handle removal manually.
type RemoveHelper struct {
	ast.NoopVisitor
	remove bool
}

// Remove marks the current node for removal.
func (v *RemoveHelper) Remove() {
	v.remove = true
}

func (v *RemoveHelper) VisitStatements(n *ast.Statements) {
	w := 0
	for i := 0; i < len(*n); i++ {
		(*n)[i].VisitWith(v.V)
		if v.remove {
			v.remove = false
			continue
		}
		if w != i {
			(*n)[w] = (*n)[i]
		}
		w++
	}
	if w == len(*n) {
		return
	}

	clear((*n)[w:])
	*n = (*n)[:w]
}

func (v *RemoveHelper) VisitExpressions(n *ast.Expressions) {
	w := 0
	for i := 0; i < len(*n); i++ {
		(*n)[i].VisitWith(v.V)
		if v.remove {
			v.remove = false
			continue
		}
		if w != i {
			(*n)[w] = (*n)[i]
		}
		w++
	}
	if w == len(*n) {
		return
	}

	clear((*n)[w:])
	*n = (*n)[:w]
}

func (v *RemoveHelper) VisitSequenceExpression(n *ast.SequenceExpression) {
	n.VisitChildrenWith(v.V)
	if len(n.Sequence) == 0 {
		v.Remove()
	}
}

func (v *RemoveHelper) VisitVariableDeclarators(n *ast.VariableDeclarators) {
	w := 0
	for i := 0; i < len(*n); i++ {
		(*n)[i].VisitWith(v.V)
		if v.remove {
			v.remove = false
			continue
		}
		if w != i {
			(*n)[w] = (*n)[i]
		}
		w++
	}
	if w == len(*n) {
		return
	}

	clear((*n)[w:])
	*n = (*n)[:w]
}

func (v *RemoveHelper) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	n.VisitChildrenWith(v.V)
	if len(n.List) == 0 {
		v.Remove()
	}
}

func (v *RemoveHelper) VisitClassElements(n *ast.ClassElements) {
	w := 0
	for i := 0; i < len(*n); i++ {
		(*n)[i].VisitWith(v.V)
		if v.remove {
			v.remove = false
			continue
		}
		if w != i {
			(*n)[w] = (*n)[i]
		}
		w++
	}
	if w == len(*n) {
		return
	}

	clear((*n)[w:])
	*n = (*n)[:w]
}

func (v *RemoveHelper) VisitProperties(n *ast.Properties) {
	w := 0
	for i := 0; i < len(*n); i++ {
		(*n)[i].VisitWith(v.V)
		if v.remove {
			v.remove = false
			continue
		}
		if w != i {
			(*n)[w] = (*n)[i]
		}
		w++
	}
	if w == len(*n) {
		return
	}

	clear((*n)[w:])
	*n = (*n)[:w]
}
