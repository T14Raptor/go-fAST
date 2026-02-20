package ext

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

// MayHaveSideEffectsStmt returns true if the statement may have side effects.
func MayHaveSideEffectsStmt(stmt ast.Statement) bool {
	switch stmt.Kind() {
	case ast.StmtBlock:
		s := stmt.MustBlock()
		for _, stmt := range s.List {
			if MayHaveSideEffectsStmt(stmt) {
				return true
			}
		}
		return false
	case ast.StmtEmpty:
		return false
	case ast.StmtLabelled:
		s := stmt.MustLabelled()
		return MayHaveSideEffectsStmt(*s.Statement)
	case ast.StmtIf:
		s := stmt.MustIf()
		if MayHaveSideEffects(s.Test) || MayHaveSideEffectsStmt(*s.Consequent) {
			return true
		}
		if s.Alternate != nil && MayHaveSideEffectsStmt(*s.Alternate) {
			return true
		}
		return false
	case ast.StmtSwitch:
		s := stmt.MustSwitch()
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
	case ast.StmtTry:
		s := stmt.MustTry()
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
	case ast.StmtClassDecl:
		s := stmt.MustClassDecl()
		return classHasSideEffect(s.Class)
	case ast.StmtFuncDecl:
		// TODO: Check in_strict mode like swc
	case ast.StmtVarDecl:
		s := stmt.MustVarDecl()
		return s.Token == token.Var
	case ast.StmtExpression:
		s := stmt.MustExpression()
		return MayHaveSideEffects(s.Expression)
	}
	return true
}
