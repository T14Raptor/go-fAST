package ast

import "slices"

// RemoveVisitor is a visitor that can remove nodes from the AST.
type RemoveVisitor struct {
	NoopVisitor
	remove bool
}

// Remove marks the current node for removal.
//
// If you override a Visit method that has deletion logic:
// 	- [RemoveVisitor.VisitStatements]
// 	- [RemoveVisitor.VisitExpressions]
// 	- [RemoveVisitor.VisitSequenceExpression]
// 	- [RemoveVisitor.VisitVariableDeclarators]
// 	- [RemoveVisitor.VisitVariableDeclaration]
// 	- [RemoveVisitor.VisitClassElements]
// 	- [RemoveVisitor.VisitProperties]
// make sure to either call the base implementation or handle removal manually.
func (v *RemoveVisitor) Remove() {
	v.remove = true
}

func (v *RemoveVisitor) VisitStatements(n *Statements) {
	count := len(*n)
	for i := 0; i < count; {
		(*n)[i].VisitWith(v.V)
		if v.remove {
			v.remove = false
			*n = slices.Delete(*n, i, i+1)
			count--
		} else {
			i++
		}
	}
}

func (v *RemoveVisitor) VisitExpressions(n *Expressions) {
	count := len(*n)
	for i := 0; i < count; {
		(*n)[i].VisitWith(v.V)
		if v.remove {
			v.remove = false
			*n = slices.Delete(*n, i, i+1)
			count--
		} else {
			i++
		}
	}
}

func (v *RemoveVisitor) VisitSequenceExpression(n *SequenceExpression) {
	n.VisitChildrenWith(v.V)
	if len(n.Sequence) == 0 {
		v.Remove()
	}
}

func (v *RemoveVisitor) VisitVariableDeclarators(n *VariableDeclarators) {
	count := len(*n)
	for i := 0; i < count; {
		(*n)[i].VisitWith(v.V)
		if v.remove {
			v.remove = false
			*n = slices.Delete(*n, i, i+1)
			count--
		} else {
			i++
		}
	}
}

func (v *RemoveVisitor) VisitVariableDeclaration(n *VariableDeclaration) {
	n.VisitChildrenWith(v.V)
	if len(n.List) == 0 {
		v.Remove()
	}
}

func (v *RemoveVisitor) VisitClassElements(n *ClassElements) {
	count := len(*n)
	for i := 0; i < count; {
		(*n)[i].VisitWith(v.V)
		if v.remove {
			v.remove = false
			*n = slices.Delete(*n, i, i+1)
			count--
		} else {
			i++
		}
	}
}

func (v *RemoveVisitor) VisitProperties(n *Properties) {
	count := len(*n)
	for i := 0; i < count; {
		(*n)[i].VisitWith(v.V)
		if v.remove {
			v.remove = false
			*n = slices.Delete(*n, i, i+1)
			count--
		} else {
			i++
		}
	}
}
