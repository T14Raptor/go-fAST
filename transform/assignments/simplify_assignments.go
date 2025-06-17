package assignments

import (
	"fmt"
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/resolver"
)

type VariableInfo struct {
	Initializer  *ast.Expression
	Declaration  *ast.VariableDeclarator
	IsReassigned bool
	ReadCount    int
	AliasOf      *ast.Id
	DeclNode     ast.Node
	IsFunction   bool
}

func extractIdentifier(target *ast.BindingTarget) *ast.Identifier {
	if target != nil {
		if ident, ok := target.Target.(*ast.Identifier); ok {
			return ident
		}
	}
	return nil
}

func isSimpleInitializer(expr *ast.Expression) bool {
	if expr == nil {
		return false
	}
	switch expr.Expr.(type) {
	case *ast.Identifier,
		*ast.NullLiteral,
		*ast.BooleanLiteral,
		*ast.NumberLiteral,
		*ast.StringLiteral,
		*ast.ObjectLiteral: // TODO: hmm is this right?
		return true
	default:
		return false
	}
}

type RedundantAssignmentCollector struct {
	ast.NoopVisitor
	variables  map[ast.Id]*VariableInfo
	scopeStack []map[string]*VariableInfo
	// Stores info for ALL variables encountered across all scopes, keyed by unique ID.
	// This is what gets passed to the next passes.
	globalVariables map[ast.Id]*VariableInfo
}

func NewRedundantAssignmentCollector() *RedundantAssignmentCollector {
	r := &RedundantAssignmentCollector{
		variables: make(map[ast.Id]*VariableInfo),
	}
	r.V = r
	return r
}

func (c *RedundantAssignmentCollector) VisitVariableDeclaration(decl *ast.VariableDeclaration) {
	// `var` might have been pre-scanned, `let`/`const` are added here.
	// Function declarations handled separately.
	for i := range decl.List {
		d := &decl.List[i]
		ident := extractIdentifier(d.Target)

		if ident == nil { // Destructuring or complex target
			if d.Target != nil {
				d.Target.VisitWith(c) // Visit the structure to find identifiers within
			}
			// Visit initializer *after* target
			if d.Initializer != nil {
				d.Initializer.VisitWith(c)
			}
			continue // Skip simple identifier logic below
		}

		// Simple identifier target
		id := ident.ToId()
		info, exists := c.variables[id]
		if !exists {
			// Not found during pre-scan (e.g., let/const or var not pre-scanned)
			info = &VariableInfo{
				Declaration:  d,
				DeclNode:     d,
				IsReassigned: false,
				ReadCount:    0,
			}
			c.variables[id] = info
			// fmt.Printf("DEBUG: VarDecl Pass - Registered: %s (ID %v)\n", ident.Name, id)
		} else {
			// Existed (e.g., pre-scanned var/function), update with details if needed
			// Be careful not to overwrite IsFunction flag etc. if it was a function
			if !info.IsFunction { // Only update DeclNode/Declaration if not function
				info.Declaration = d
				info.DeclNode = d
			}
			// fmt.Printf("DEBUG: VarDecl Pass - Found Existing: %s (ID %v)\n", ident.Name, id)
		}

		// Set/update initializer *before* visiting it
		info.Initializer = d.Initializer

		// Visit initializer *after* the variable ID is definitely in the map
		if d.Initializer != nil {
			d.Initializer.VisitWith(c)
		}
	}
}
func (c *RedundantAssignmentCollector) preScanScopeDeclarations(stmts ast.Statements) {
	for _, stmt := range stmts {
		// Find function declarations first (most important for hoisting)
		if funcDecl, ok := stmt.Stmt.(*ast.FunctionDeclaration); ok {
			if funcDecl.Function.Name != nil {
				id := funcDecl.Function.Name.ToId()
				// Only add if not already known (handles block-scoped functions potentially)
				if _, exists := c.variables[id]; !exists {
					c.variables[id] = &VariableInfo{
						DeclNode:     funcDecl, // Store the declaration node
						IsFunction:   true,
						IsReassigned: false, // Functions aren't typically "reassigned" by declaration
						ReadCount:    0,
					}
				}
			}
		}
	}
}

func (c *RedundantAssignmentCollector) processScopeStatements(stmts ast.Statements) {
	for _, stmt := range stmts {
		stmt.VisitWith(c) // Visit each statement normally now
	}
}

func (c *RedundantAssignmentCollector) VisitProgram(prog *ast.Program) {
	c.preScanScopeDeclarations(prog.Body)
	c.processScopeStatements(prog.Body)
}
func (c *RedundantAssignmentCollector) VisitBlockStatement(block *ast.BlockStatement) {
	c.preScanScopeDeclarations(block.List)
	c.processScopeStatements(block.List)
}

func (c *RedundantAssignmentCollector) VisitAssignExpression(expr *ast.AssignExpression) {
	// Determine RHS type *before* visiting it
	rhsIdent, rhsIsIdent := expr.Right.Expr.(*ast.Identifier)

	// Visit RHS first - ensures reads are counted correctly before assignment effect
	expr.Right.VisitWith(c)

	// Process LHS
	lhsIdent, lhsIsIdent := expr.Left.Expr.(*ast.Identifier)

	if lhsIsIdent {
		lhsId := lhsIdent.ToId()
		if info, exists := c.variables[lhsId]; exists {
			info.IsReassigned = true // Always mark as reassigned on assignment

			// Check if assigning a simple identifier (potential alias source)
			if rhsIsIdent {
				rhsId := rhsIdent.ToId()
				// Ensure the RHS identifier corresponds to a known variable declaration
				if _, rhsVarExists := c.variables[rhsId]; rhsVarExists {
					// Set the alias information
					info.AliasOf = &rhsId
					//fmt.Printf("DEBUG: Var %s (ID %v) assigned alias from %s (ID %v)\n", lhsIdent.Name, lhsId, rhsIdent.Name, rhsId)
				} else {
					// RHS identifier is not a known variable, clear any old alias
					info.AliasOf = nil
				}
			} else {
				// Assigned something complex, breaks any previous alias chain for LHS
				info.AliasOf = nil
			}
		}
		// Don't visit LHS identifier itself again
	} else {
		expr.Left.VisitWith(c)
		// TODO: Consider if assignment to complex LHS should clear AliasOf for vars *within* it?
		// Probably not necessary for this optimization level.
	}
}

func (c *RedundantAssignmentCollector) VisitFunctionDeclaration(decl *ast.FunctionDeclaration) {
	// The function name ID should already be in c.variables from the pre-scan.
	// We just need to visit the literal to process parameters and the body contents.
	// The VisitFunctionLiteral method now handles the body processing correctly.
	if decl.Function != nil {
		decl.Function.VisitWith(c)
	}
}

func (c *RedundantAssignmentCollector) VisitFunctionLiteral(funcLit *ast.FunctionLiteral) {
	// 1. Process Parameters normally - Add them to the variables map
	for i := range funcLit.ParameterList.List {
		paramDeclarator := &funcLit.ParameterList.List[i]
		if paramIdent := extractIdentifier(paramDeclarator.Target); paramIdent != nil {
			id := paramIdent.ToId()
			if _, exists := c.variables[id]; !exists { // Should not exist from outer scope normally
				c.variables[id] = &VariableInfo{
					Initializer:  paramDeclarator.Initializer,
					Declaration:  paramDeclarator, // Point to the parameter's declarator
					DeclNode:     paramDeclarator,
					IsReassigned: false,
					ReadCount:    0,
				}
			}
			// Visit default initializer AFTER param is registered
			if paramDeclarator.Initializer != nil {
				paramDeclarator.Initializer.VisitWith(c)
			}
		} else { // Destructuring etc.
			if paramDeclarator.Target != nil {
				paramDeclarator.Target.VisitWith(c) // Visit target structure
			}
			if paramDeclarator.Initializer != nil {
				paramDeclarator.Initializer.VisitWith(c)
			}
		}
	}
	//// Handle Rest parameter
	//if funcLit.ParameterList.Rest != nil {
	//	if restIdent, ok := funcLit.ParameterList.Rest.(*ast.Identifier); ok {
	//		id := restIdent.ToId()
	//		if _, exists := c.variables[id]; !exists {
	//			c.variables[id] = &VariableInfo{
	//				// Declaration: nil, // No specific declarator for rest param identifier itself
	//				DeclNode: restIdent, // Point to the identifier node
	//				// IsReassigned: false, // Array itself isn't reassigned by syntax
	//				// ReadCount: 0,
	//			}
	//		}
	//	} else {
	//		funcLit.ParameterList.Rest.VisitWith(c)
	//	}
	//}

	// 2. Process the function body using the scope handling methods
	//if block, ok := funcLit.Body.List.(*ast.BlockStatement); ok {
	c.preScanScopeDeclarations(funcLit.Body.List) // Find nested function decls first
	c.processScopeStatements(funcLit.Body.List)   // Then process statements

	if funcLit.Body != nil {
		// Handle non-block bodies (e.g., arrow function expression body `() => expr;`)
		// These don't have their own block scope typically, and don't contain declarations
		funcLit.Body.VisitWith(c)
	}
}

func (c *RedundantAssignmentCollector) VisitUpdateExpression(expr *ast.UpdateExpression) {
	if ident, ok := expr.Operand.Expr.(*ast.Identifier); ok {
		id := ident.ToId() // Use unique ID
		if info, exists := c.variables[id]; exists {
			info.IsReassigned = true
		}
	} else {
		expr.Operand.VisitWith(c)
	}
}

func (c *RedundantAssignmentCollector) VisitIdentifier(ident *ast.Identifier) {
	id := ident.ToId() // Use unique ID
	if info, exists := c.variables[id]; exists {
		// Check context: Is this a read, or part of a declaration/LHS handled elsewhere?
		// Assuming context is handled by overrides (like in Replacer), this counts reads.
		info.ReadCount++
	}
}

func (c *RedundantAssignmentCollector) VisitMemberExpression(memberExpr *ast.MemberExpression) {
	memberExpr.Object.VisitWith(c)
}

type RedundantAssignmentReplacer struct {
	ast.NoopVisitor
	variables map[ast.Id]*VariableInfo
}

func NewRedundantAssignmentReplacer(vars map[ast.Id]*VariableInfo) *RedundantAssignmentReplacer {
	r := &RedundantAssignmentReplacer{
		variables: vars,
	}
	r.V = r
	return r
}

func (r *RedundantAssignmentReplacer) VisitExpression(expr *ast.Expression) {
	ident, isIdent := expr.Expr.(*ast.Identifier)
	if !isIdent {
		expr.VisitChildrenWith(r)
		return
	}

	originalId := ident.ToId()
	originalInfo, exists := r.variables[originalId]

	if !exists {
		return
	}

	finalAliasId, resolved := resolveAliasChain(originalId, r.variables)

	// Check if the final resolved ID is different from the original one
	// And ensure the resolution was successful (no cycle/error)
	if resolved && finalAliasId != originalId {
		finalAliasInfo, finalAliasExists := r.variables[finalAliasId]

		if finalAliasExists { // Ensure the final target info exists
			var replacementName string

			if finalAliasInfo.IsFunction {
				// Target is a function, get name from DeclNode
				if funcDecl, ok := finalAliasInfo.DeclNode.(*ast.FunctionDeclaration); ok && funcDecl.Function.Name != nil {
					replacementName = funcDecl.Function.Name.Name
				} else {
					fmt.Printf("DEBUG: Could not extract function name for final alias ID %v from DeclNode %T\n", finalAliasId, finalAliasInfo.DeclNode)
				}
			} else {
				// Target is likely a var/param, try Declaration first
				if finalAliasInfo.Declaration != nil {
					if identNode := extractIdentifier(finalAliasInfo.Declaration.Target); identNode != nil {
						replacementName = identNode.Name
					}
				}
				// Fallback: If Declaration was nil (e.g., maybe a param/rest identifier stored directly in DeclNode?)
				if replacementName == "" {
					if identNode, ok := finalAliasInfo.DeclNode.(*ast.Identifier); ok {
						replacementName = identNode.Name
					} else {
						fmt.Printf("DEBUG: Could not extract variable/param name for final alias ID %v from Declaration or DeclNode %T\n", finalAliasId, finalAliasInfo.DeclNode)
					}
				}
			}

			// Proceed with replacement if a name was found
			if replacementName != "" {
				// fmt.Printf("DEBUG: Replacing usage of %s (ID %v) with final alias %s (ID %v)\n", ident.Name, originalId, replacementName, finalAliasId)
				// Replace with the resolved name
				expr.Expr = &ast.Identifier{Name: replacementName} // Use resolved name

				// Adjust ReadCounts: Decrement for original, Increment for final target.
				originalInfo.ReadCount--
				finalAliasInfo.ReadCount++ // Increment read count of the actual target 'b'

				return // Replacement done
			} else {
				//fmt.Printf("DEBUG: Failed to find a valid replacement name for final alias ID %v\n", finalAliasId)
			}
		} else {
			//fmt.Printf("DEBUG: Final alias target ID %v not found in variables map\n", finalAliasId)
		}

	}

	// Only run if not replaced by an alias
	if !originalInfo.IsReassigned && isSimpleInitializer(originalInfo.Initializer) {
		clonedNode := originalInfo.Initializer.Clone()
		if clonedNode == nil {
			// fmt.Printf("Error cloning initializer for %s (ID %v)\n", ident.Name, originalId)
			return
		}
		if clonedExpr, ok := clonedNode.Expr.(ast.Expr); ok {
			// fmt.Printf("DEBUG: Replacing usage of %s (ID %v) with its simple initializer\n", ident.Name, originalId)
			originalInfo.ReadCount--
			expr.Expr = clonedExpr
			return
		} else {
			// fmt.Printf("Error: Cloned initializer node type %T is not an ast.Expr for %s\n", clonedNode.Expr, ident.Name)
		}
	}
}

func isSideEffectFreeAssignment(expr *ast.AssignExpression, vars map[ast.Id]*VariableInfo) bool {
	// Basic case: identifier = identifier
	_, lhsIsIdent := expr.Left.Expr.(*ast.Identifier)
	_, rhsIsIdent := expr.Right.Expr.(*ast.Identifier)
	if lhsIsIdent && rhsIsIdent {
		// Could add checks here: ensure RHS isn't a getter, LHS isn't a setter etc.
		return true
	}

	// Maybe allow identifier = simple_literal?
	if lhsIsIdent && isSimpleInitializer(&ast.Expression{Expr: expr.Right.Expr}) { // Check if RHS expr is simple literal type
		// Ensure LHS is not a setter? Assume false for now.
		return false // Let's be conservative, only allow ident=ident for now
	}

	return false
}

func (r *RedundantAssignmentReplacer) VisitVariableDeclarator(decl *ast.VariableDeclarator) {
	// Only visit the initializer, NOT the target identifier
	if decl.Initializer != nil {
		decl.Initializer.VisitWith(r)
	}
}

func (r *RedundantAssignmentReplacer) VisitAssignExpression(expr *ast.AssignExpression) {
	if _, isIdent := expr.Left.Expr.(*ast.Identifier); !isIdent {
		expr.Left.VisitWith(r)
	}
	expr.Right.VisitWith(r)
}

func (r *RedundantAssignmentReplacer) VisitUpdateExpression(expr *ast.UpdateExpression) {

}

type CombinedRemover struct {
	ast.NoopVisitor
	variables map[ast.Id]*VariableInfo
}

func NewCombinedRemover(vars map[ast.Id]*VariableInfo) *CombinedRemover {
	r := &CombinedRemover{
		variables: vars,
	}
	r.V = r
	return r
}

// checkDeclaratorRemoval determines if a single declarator can be removed.
func (r *CombinedRemover) checkDeclaratorRemoval(d *ast.VariableDeclarator) bool {
	ident, ok := d.Target.Target.(*ast.Identifier)
	if !ok {
		return false // Cannot remove complex targets
	}

	id := ident.ToId()
	if info, exists := r.variables[id]; exists {
		// Condition: Simple initializer, NOT function, never reassigned, AND read count zero
		if !info.IsFunction && isSimpleInitializer(info.Initializer) && !info.IsReassigned && info.ReadCount <= 0 {
			return true
		}
	}
	return false
}

func (r *CombinedRemover) VisitBlockStatement(block *ast.BlockStatement) {
	newList, modified := r.filterStmtList(block.List)
	if modified {
		block.List = newList
	}
}

func (r *CombinedRemover) VisitProgram(prog *ast.Program) {
	newList, modified := r.filterStmtList(prog.Body)
	if modified {
		prog.Body = newList
	}
}

func Simplify(p *ast.Program, resolve bool) {
	if resolve {
		resolver.Resolve(p)
	}

	// Pass 1: Collect Info (Variable declarations, assignments, initial reads)
	collector := NewRedundantAssignmentCollector()
	p.VisitWith(collector)

	// Pass 2: Replace Usages (Replace identifiers with simple initializers where possible)
	replacer := NewRedundantAssignmentReplacer(collector.variables)
	p.VisitWith(replacer)

	// Pass 3: Remove Redundant Declarators & Empty Declaration Statements
	// This combined pass modifies declaration lists and removes empty statements directly.
	remover := NewCombinedRemover(collector.variables)
	p.VisitWith(remover)
}
