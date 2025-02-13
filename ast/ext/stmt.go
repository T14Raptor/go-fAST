package ext

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

// MayHaveSideEffectsStmt returns true if the statement may have side effects.
func MayHaveSideEffectsStmt(stmt ast.Statement) bool {
	switch s := stmt.Stmt.(type) {
	case *ast.BlockStatement:
		for _, stmt := range s.List {
			if MayHaveSideEffectsStmt(stmt) {
				return true
			}
		}
		return false
	case *ast.EmptyStatement:
		return false
	case *ast.LabelledStatement:
		return MayHaveSideEffectsStmt(*s.Statement)
	case *ast.IfStatement:
		if MayHaveSideEffects(s.Test) || MayHaveSideEffectsStmt(*s.Consequent) {
			return true
		}
		if s.Alternate != nil && MayHaveSideEffectsStmt(*s.Alternate) {
			return true
		}
		return false
	case *ast.SwitchStatement:
		if MayHaveSideEffects(s.Discriminant) {
			return true
		}
		for _, c := range s.Body {
			if MayHaveSideEffects(c.Test) {
				return true
			}
			for _, stmt := range c.Consequent {
				if MayHaveSideEffectsStmt(stmt) {
					return true
				}
			}
		}
		return false
	case *ast.TryStatement:
		for _, stmt := range s.Body.List {
			if MayHaveSideEffectsStmt(stmt) {
				return true
			}
		}
		if s.Catch != nil {
			for _, stmt := range s.Catch.Body.List {
				if MayHaveSideEffectsStmt(stmt) {
					return true
				}
			}
		}
		if s.Finally != nil {
			for _, stmt := range s.Finally.List {
				if MayHaveSideEffectsStmt(stmt) {
					return true
				}
			}
		}
		return false
	case *ast.ClassDeclaration:
		return classHasSideEffect(s.Class)
	case *ast.FunctionDeclaration:
		// TODO: Check in_strict mode like swc
	case *ast.VariableDeclaration:
		return s.Token == token.Var
	case *ast.ExpressionStatement:
		return MayHaveSideEffects(s.Expression)
	}
	return true
}
