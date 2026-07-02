package resolver

import (
	"github.com/t14raptor/go-fast/ast"
)

type hoister struct {
	ast.NoopVisitor

	resolver *resolver
	kind     declKind
	inBlock  bool

	inCatchBody bool

	excludedFromCatch map[string]struct{}
	catchParamDecls   map[string]struct{}

	binder hoisterBinder
}

func newHoister(resolver *resolver) *hoister {
	h := &hoister{
		resolver: resolver,
		kind:     declKindVar,
	}
	h.binder.h = h
	h.binder.V = &h.binder
	return h
}

// hoisterBinder registers each binding identifier of a pattern for hoisting,
// skipping default values and computed keys (see resolverBinder). One instance
// lives on the hoister and is reused.
type hoisterBinder struct {
	ast.NoopVisitor

	h *hoister
}

func (b *hoisterBinder) VisitExpression(*ast.Expression) {}

func (b *hoisterBinder) VisitIdentifier(n *ast.Identifier) { b.h.addIdentifier(n) }

func (h *hoister) addIdentifier(id *ast.Identifier) {
	if h.inCatchBody {
		if _, ok := h.catchParamDecls[id.Name]; ok {
			if r, _ := h.resolver.lookupContext(id.Name); r != ast.UnresolvedContext {
				id.ScopeContext = r
				return
			}
		}

		if h.excludedFromCatch == nil {
			h.excludedFromCatch = make(map[string]struct{})
		}
		h.excludedFromCatch[id.Name] = struct{}{}
	} else if _, ok := h.catchParamDecls[id.Name]; ok {
		if _, excluded := h.excludedFromCatch[id.Name]; !excluded {
			return
		}
	}

	h.resolver.modify(id, h.kind)
}

func (h *hoister) VisitBlockStatement(n *ast.BlockStatement) {
	old := h.inBlock
	h.inBlock = true
	n.VisitChildrenWith(h)
	h.inBlock = old
}

func (h *hoister) VisitCatchClause(n *ast.CatchClause) {
	oldExclude := h.excludedFromCatch
	h.excludedFromCatch = nil
	oldInCatchBody := h.inCatchBody

	var paramName string
	if n.Parameter != nil {
		if ident, ok := n.Parameter.Identifier(); ok {
			paramName = ident.Name
		}
	}

	var hadParam bool
	if paramName != "" {
		if h.catchParamDecls == nil {
			h.catchParamDecls = make(map[string]struct{})
		}
		_, hadParam = h.catchParamDecls[paramName]
		h.catchParamDecls[paramName] = struct{}{}
	}

	h.inCatchBody = true
	n.Body.VisitWith(h)

	if paramName != "" && !hadParam {
		delete(h.catchParamDecls, paramName)
	}
	h.inCatchBody = oldInCatchBody
	h.excludedFromCatch = oldExclude
}

func (h *hoister) VisitStatements(n *ast.Statements) {
	others := make(ast.Statements, 0, len(*n))
	for i := range *n {
		switch (*n)[i].Kind() {
		case ast.StmtVarDecl:
			(*n)[i].MustVarDecl().VisitWith(h)
		case ast.StmtFuncDecl:
			(*n)[i].MustFuncDecl().VisitWith(h)
		default:
			others = append(others, (*n)[i])
		}
	}

	for i := range others {
		others[i].VisitWith(h)
	}
}

func (h *hoister) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	if h.inBlock && n.Kind != ast.VarKindVar {
		return
	}

	oldKind := h.kind
	h.kind = declKindVar
	n.VisitChildrenWith(h)
	h.kind = oldKind
}

func (h *hoister) VisitPattern(n *ast.Pattern) {
	n.VisitWith(&h.binder)
}

func (h *hoister) VisitForStatement(n *ast.ForStatement) {
	if n.Initializer != nil {
		if decl, ok := n.Initializer.VarDecl(); ok && decl.Kind == ast.VarKindVar {
			decl.VisitWith(h)
		}
	}

	old := h.inBlock
	h.inBlock = true
	n.Body.VisitWith(h)
	h.inBlock = old
}

func (h *hoister) VisitForInStatement(n *ast.ForInStatement) {
	if decl, ok := n.Into.VarDecl(); ok && decl.Kind == ast.VarKindVar {
		decl.VisitWith(h)
	}

	old := h.inBlock
	h.inBlock = true
	n.Body.VisitWith(h)
	h.inBlock = old
}

func (h *hoister) VisitForOfStatement(n *ast.ForOfStatement) {
	if decl, ok := n.Into.VarDecl(); ok && decl.Kind == ast.VarKindVar {
		decl.VisitWith(h)
	}

	old := h.inBlock
	h.inBlock = true
	n.Body.VisitWith(h)
	h.inBlock = old
}

func (h *hoister) VisitFunctionDeclaration(n *ast.FunctionDeclaration) {
	if _, ok := h.catchParamDecls[n.Function.Name.Name]; ok {
		return
	}

	if h.inBlock {
		if kind, declared := h.resolver.current.isDeclared(n.Function.Name.Name); declared {
			if kind != declKindVar && kind != declKindFunction {
				return
			}
		}
	}

	h.resolver.modify(n.Function.Name, declKindFunction)
}

func (h *hoister) VisitSwitchStatement(n *ast.SwitchStatement) {
	n.Discriminant.VisitWith(h)

	old := h.inBlock
	h.inBlock = true
	n.Body.VisitWith(h)
	h.inBlock = old
}

func (h *hoister) VisitArrowFunctionLiteral(*ast.ArrowFunctionLiteral) {}
func (h *hoister) VisitExpression(*ast.Expression)                     {}
func (h *hoister) VisitFunctionLiteral(*ast.FunctionLiteral)           {}
