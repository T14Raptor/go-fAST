package ast

// RemoveVisitor is a helper visitor that can remove nodes from the AST.
type RemoveVisitor struct {
	NoopVisitor
	remove bool
}

// Remove marks the current node for removal.
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
func (v *RemoveVisitor) Remove() {
	v.remove = true
}

func (v *RemoveVisitor) VisitStatements(n *Statements) {
	w := 0
	for i := 0; i < len(*n); i++ {
		(*n)[i].VisitWith(v.V)
		if v.remove {
			v.remove = false
			continue
		}
		(*n)[w] = (*n)[i]
		w++
	}

	clear((*n)[w:])
	*n = (*n)[:w]
}

func (v *RemoveVisitor) VisitExpressions(n *Expressions) {
	w := 0
	for i := 0; i < len(*n); i++ {
		(*n)[i].VisitWith(v.V)
		if v.remove {
			v.remove = false
			continue
		}
		(*n)[w] = (*n)[i]
		w++
	}

	clear((*n)[w:])
	*n = (*n)[:w]
}

func (v *RemoveVisitor) VisitSequenceExpression(n *SequenceExpression) {
	n.VisitChildrenWith(v.V)
	if len(n.Sequence) == 0 {
		v.Remove()
	}
}

func (v *RemoveVisitor) VisitVariableDeclarators(n *VariableDeclarators) {
	w := 0
	for i := 0; i < len(*n); i++ {
		(*n)[i].VisitWith(v.V)
		if v.remove {
			v.remove = false
			continue
		}
		(*n)[w] = (*n)[i]
		w++
	}

	clear((*n)[w:])
	*n = (*n)[:w]
}

func (v *RemoveVisitor) VisitVariableDeclaration(n *VariableDeclaration) {
	n.VisitChildrenWith(v.V)
	if len(n.List) == 0 {
		v.Remove()
	}
}

func (v *RemoveVisitor) VisitClassElements(n *ClassElements) {
	w := 0
	for i := 0; i < len(*n); i++ {
		(*n)[i].VisitWith(v.V)
		if v.remove {
			v.remove = false
			continue
		}
		(*n)[w] = (*n)[i]
		w++
	}

	clear((*n)[w:])
	*n = (*n)[:w]
}

func (v *RemoveVisitor) VisitProperties(n *Properties) {
	w := 0
	for i := 0; i < len(*n); i++ {
		(*n)[i].VisitWith(v.V)
		if v.remove {
			v.remove = false
			continue
		}
		(*n)[w] = (*n)[i]
		w++
	}

	clear((*n)[w:])
	*n = (*n)[:w]
}
