package ext

import "github.com/t14raptor/go-fast/ast"

// RemoveVisitor is a helper visitor that can help remove nodes from the AST.
//
// If you override a Visit method that has deletion logic:
//   - [RemoveVisitor.VisitStatements]
//   - [RemoveVisitor.VisitExpressions]
//   - [RemoveVisitor.VisitSequenceExpression]
//   - [RemoveVisitor.VisitVariableDeclarators]
//   - [RemoveVisitor.VisitVariableDeclaration]
//   - [RemoveVisitor.VisitClassElements]
//   - [RemoveVisitor.VisitProperties]
//
// make sure to either call the base implementation or handle removal manually.
type RemoveVisitor struct {
	ast.NoopVisitor
	remove bool
}

// Remove marks the current node for removal.
func (v *RemoveVisitor) Remove() {
	v.remove = true
}

func (v *RemoveVisitor) VisitStatements(n *ast.Statements) {
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

func (v *RemoveVisitor) VisitExpressions(n *ast.Expressions) {
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

func (v *RemoveVisitor) VisitSequenceExpression(n *ast.SequenceExpression) {
	n.VisitChildrenWith(v.V)
	if len(n.Sequence) == 0 {
		v.Remove()
	}
}

func (v *RemoveVisitor) VisitVariableDeclarators(n *ast.VariableDeclarators) {
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

func (v *RemoveVisitor) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	n.VisitChildrenWith(v.V)
	if len(n.List) == 0 {
		v.Remove()
	}
}

func (v *RemoveVisitor) VisitClassElements(n *ast.ClassElements) {
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

func (v *RemoveVisitor) VisitProperties(n *ast.Properties) {
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
