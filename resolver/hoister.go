package resolver

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/parser/scanner/token"
)

// hoister handles the first phase of resolution: it walks a statement
// list and declares every var and function-declaration name in the
// enclosing scope (and tracks the catch-parameter set)
type hoister struct {
	ast.NoopVisitor

	resolver *Resolver

	inBlock     bool
	inCatchBody bool

	excludedFromCatch map[string]struct{}
	catchParamDecls   map[string]struct{}
}

func getHoister(r *Resolver) *hoister {
	h := &hoister{}
	h.resolver = r
	h.inBlock = false
	h.inCatchBody = false
	clear(h.excludedFromCatch)
	clear(h.catchParamDecls)
	h.V = h
	return h
}

func putHoister(h *hoister) {
	_ = h
}

func (h *hoister) addIdent(id *ast.Identifier) {
	if h.inCatchBody {
		if _, ok := h.catchParamDecls[id.Name]; ok {
			if mark, _ := h.resolver.lookupContext(id.Name); mark != UnresolvedMark {
				return
			}
		}
		if h.excludedFromCatch == nil {
			h.excludedFromCatch = make(map[string]struct{}, 2)
		}
		h.excludedFromCatch[id.Name] = struct{}{}
	} else if _, ok := h.catchParamDecls[id.Name]; ok {
		if _, excluded := h.excludedFromCatch[id.Name]; !excluded {
			return
		}
	}

	h.resolver.modify(id, DeclKindVar)
}

// VisitStatements processes vars and function declarations before the
// rest so forward references inside expressions can resolve to bindings
// later in source order.
func (h *hoister) VisitStatements(n *ast.Statements) {
	for i := range *n {
		switch (*n)[i].Stmt.(type) {
		case *ast.VariableDeclaration, *ast.FunctionDeclaration, *ast.ClassDeclaration:
			(*n)[i].VisitWith(h)
		}
	}
	for i := range *n {
		switch (*n)[i].Stmt.(type) {
		case *ast.VariableDeclaration, *ast.FunctionDeclaration, *ast.ClassDeclaration:
		default:
			(*n)[i].VisitWith(h)
		}
	}
}

func (h *hoister) VisitBlockStatement(n *ast.BlockStatement) {
	old := h.inBlock
	h.inBlock = true
	n.VisitChildrenWith(h)
	h.inBlock = old
}

// Resolver.VisitCatchStatement declares the parameter; the hoister only
// records the param name so a redeclaring `var` inside the catch body
// is handled
func (h *hoister) VisitCatchStatement(n *ast.CatchStatement) {
	var paramName string
	if n.Parameter != nil {
		// Patterns (object/array binding) aren't covered by B.3.5;
		// leaving paramName empty skips the special-case handling.
		if id, ok := n.Parameter.Target.(*ast.Identifier); ok {
			paramName = id.Name
		}
	}

	oldInCatchBody := h.inCatchBody
	oldExclude := h.excludedFromCatch
	h.excludedFromCatch = nil

	var hadParam bool
	if paramName != "" {
		if h.catchParamDecls == nil {
			h.catchParamDecls = make(map[string]struct{}, 1)
		}
		_, hadParam = h.catchParamDecls[paramName]
		h.catchParamDecls[paramName] = struct{}{}
	}

	h.inCatchBody = true
	n.Body.VisitWith(h)
	h.inCatchBody = oldInCatchBody

	if paramName != "" && !hadParam {
		delete(h.catchParamDecls, paramName)
	}
	h.excludedFromCatch = oldExclude
}

func (h *hoister) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	if h.inBlock && n.Token != token.Var {
		return
	}
	n.VisitChildrenWith(h)
}

func (h *hoister) VisitBindingTarget(n *ast.BindingTarget) {
	if ident, ok := n.Target.(*ast.Identifier); ok {
		h.addIdent(ident)
		return
	}
	n.VisitChildrenWith(h)
}

func (h *hoister) VisitFunctionDeclaration(n *ast.FunctionDeclaration) {
	if _, ok := h.catchParamDecls[n.Function.Name.Name]; ok {
		return
	}
	if h.inBlock {
		if kind, declared := h.resolver.current.isDeclared(n.Function.Name.Name); declared {
			if kind != DeclKindVar && kind != DeclKindFunction {
				return
			}
		}
	}
	h.resolver.modify(n.Function.Name, DeclKindFunction)
}

func (h *hoister) VisitClassDeclaration(n *ast.ClassDeclaration) {
	if n == nil || n.Class == nil || n.Class.Name == nil || n.Class.Name.Name == "" {
		return
	}
	h.resolver.modify(n.Class.Name, DeclKindClass)
}

func (h *hoister) VisitSwitchStatement(n *ast.SwitchStatement) {
	n.Discriminant.VisitWith(h)

	old := h.inBlock
	h.inBlock = true
	n.Body.VisitWith(h)
	h.inBlock = old
}

// for-loop heads are block scopes for let/const. var still hoists out
// thanks to the Token guard in VisitVariableDeclaration.
func (h *hoister) VisitForStatement(n *ast.ForStatement)     { h.visitForLike(n) }
func (h *hoister) VisitForInStatement(n *ast.ForInStatement) { h.visitForLike(n) }
func (h *hoister) VisitForOfStatement(n *ast.ForOfStatement) { h.visitForLike(n) }

func (h *hoister) visitForLike(n ast.VisitableNode) {
	old := h.inBlock
	h.inBlock = true
	n.VisitChildrenWith(h)
	h.inBlock = old
}

func (h *hoister) VisitArrowFunctionLiteral(*ast.ArrowFunctionLiteral) {}
func (h *hoister) VisitExpression(*ast.Expression)                     {}
func (h *hoister) VisitFunctionLiteral(*ast.FunctionLiteral)           {}

type idsFinder struct {
	ast.NoopVisitor

	found []ast.Id
}

func findIds(n ast.VisitableNode) []ast.Id {
	v := &idsFinder{}
	v.V = v
	n.VisitWith(v)
	return v.found
}

func (v *idsFinder) VisitExpression(*ast.Expression) {}

func (v *idsFinder) VisitIdentifier(n *ast.Identifier) {
	v.found = append(v.found, n.ToId())
}
