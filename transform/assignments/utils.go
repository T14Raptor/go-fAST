package assignments

import "github.com/t14raptor/go-fast/ast"

func resolveAliasChain(startId ast.Id, variables map[ast.Id]*VariableInfo) (finalId ast.Id, success bool) {
	currentId := startId
	seen := make(map[ast.Id]bool)

	for i := 0; i < 100; i++ { // Limit iterations to prevent infinite loops in unforeseen cases
		info, exists := variables[currentId]
		if !exists {
			// Should not happen if collector worked, but indicates break in chain
			return startId, false // Return original on failure
		}

		// Found the end of the chain?
		if info.AliasOf == nil {
			return currentId, true // Successfully found the end
		}

		// Move to the next link
		nextId := *info.AliasOf

		// Check for cycle
		if seen[nextId] {
			//fmt.Printf("DEBUG: Cycle detected in alias chain starting from ID %v at ID %v\n", startId, nextId)
			return startId, false // Return original on cycle
		}
		seen[currentId] = true // Mark current node as visited *before* moving
		currentId = nextId
	}

	//fmt.Printf("DEBUG: Alias chain resolution exceeded max iterations for start ID %v\n", startId)
	return startId, false // Failed to resolve within limit
}

// filterAndModifyDeclarators removes eligible declarators from a list.
// It returns the potentially modified list and a boolean indicating if the *entire*
// declaration became empty as a result.
func (r *CombinedRemover) filterAndModifyDeclarators(declarators ast.VariableDeclarators) (ast.VariableDeclarators, bool) {
	newList := make(ast.VariableDeclarators, 0, len(declarators))
	modified := false
	// Use index to get correct pointer from the original slice
	for i := range declarators {
		dPtr := declarators[i] // Get pointer to the actual declarator

		// First, visit children of the declarator (specifically the initializer)
		// to handle nested removals before deciding on this declarator.
		if dPtr.Initializer != nil {
			dPtr.Initializer.VisitWith(r)
		}
		// Now check if THIS declarator can be removed, using the correct pointer
		if !r.checkDeclaratorRemoval(&dPtr) { // *** Use dPtr here ***
			newList = append(newList, dPtr) // Keep it
		} else {
			modified = true // We removed at least one
		}
	}

	if !modified {
		return declarators, false // Return original slice, totally empty = false
	}
	return newList, len(newList) == 0 // Return new slice and whether it's now empty
}

// filterStmtList processes a list of statements, removing empty VariableDeclarations.
// It returns the new list and whether modifications occurred.
func (r *CombinedRemover) filterStmtList(stmts ast.Statements) (ast.Statements, bool) {
	newStmts := make(ast.Statements, 0, len(stmts))
	listModified := false

	for _, stmt := range stmts {
		keepStmt := true

		// --- Modification Start: Check for removable assignments ---
		if exprStmt, ok := stmt.Stmt.(*ast.ExpressionStatement); ok {
			if assignExpr, ok := exprStmt.Expression.Expr.(*ast.AssignExpression); ok {
				// Check if LHS is a simple identifier
				if lhsIdent, ok := assignExpr.Left.Expr.(*ast.Identifier); ok {
					lhsId := lhsIdent.ToId()
					// Check if variable info exists and ReadCount is zero AFTER Replacer pass
					if info, exists := r.variables[lhsId]; exists && info.ReadCount <= 0 {
						// Check if the assignment itself is likely side-effect free
						if isSideEffectFreeAssignment(assignExpr, r.variables) {
							// fmt.Printf("DEBUG: Removing redundant assignment to %s (ID %v)\n", lhsIdent.Name, lhsId)
							keepStmt = false // Mark statement for removal
							listModified = true
						}
					}
				}
			}
		}

		if varDecl, ok := stmt.Stmt.(*ast.VariableDeclaration); ok {
			newList, isEmpty := r.filterAndModifyDeclarators(varDecl.List)
			if isEmpty {
				keepStmt = false // Already marked for removal if declaration empty
				listModified = true
			} else if len(newList) != len(varDecl.List) {
				varDecl.List = newList
				listModified = true
			}
			// If keepStmt is true here, varDecl wasn't fully emptied
		}

		if keepStmt {
			newStmts = append(newStmts, stmt)
			// Visit children of kept statements (unless it was a varDecl whose init was visited)
			if _, ok := stmt.Stmt.(*ast.VariableDeclaration); !ok {
				stmt.VisitWith(r) // Visit children of other statement types (including ExpressionStatement if kept)
			}
		}
	}

	return newStmts, listModified
}
