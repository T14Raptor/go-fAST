package resolver

import (
	"maps"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

type hoister struct {
	ast.NoopVisitor

	resolver *Resolver
	kind     DeclKind
	inBlock  bool

	inCatchBody bool

	excludedFromCatch map[string]struct{}
	catchParamDecls   map[string]struct{}
}

func newHoister(resolver *Resolver) *hoister {
	return &hoister{
		resolver:          resolver,
		kind:              DeclKindVar,
		excludedFromCatch: make(map[string]struct{}),
		catchParamDecls:   make(map[string]struct{}),
	}
}

func (h *hoister) addIdent(id *ast.Identifier) {
	if h.inCatchBody {
		if _, ok := h.catchParamDecls[id.Name]; ok {
			if r, _ := h.resolver.lookupContext(id.Name); r != UnresolvedMark {
				return
			}
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

func (h *hoister) VisitCatchStatement(n *ast.CatchStatement) {
	oldExclude := h.excludedFromCatch
	h.excludedFromCatch = make(map[string]struct{})
	oldInCatchBody := h.inCatchBody

	if n.Parameter != nil {
		if params := findIds(n.Parameter); len(params) == 1 {
			h.catchParamDecls[params[0].Name] = struct{}{}
		}
	}

	old := maps.Clone(h.catchParamDecls)

	h.inCatchBody = true
	n.Body.VisitWith(h)
	h.inCatchBody = false
	if n.Parameter != nil {
		n.Parameter.VisitWith(h)
	}

	h.catchParamDecls = old
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
	if h.inBlock && n.Token != token.Var {
		return
	}

	oldKind := h.kind
	h.kind = DeclKindVar
	n.VisitChildrenWith(h)
	h.kind = oldKind
}

func (h *hoister) VisitBindingTarget(n *ast.BindingTarget) {
	if ident, ok := n.Ident(); ok {
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
