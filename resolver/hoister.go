package resolver

import "github.com/t14raptor/go-fast/ast"

/*
	Loosely inspired from https://rustdoc.swc.rs/swc_ecma_transforms_base/fn.resolver.html
*/

type Hoister struct {
	resolver          *Resolver
	kind              DeclKind
	inBlock           bool
	inCatchBody       bool
	excludedFromCatch map[string]struct{}
	catchParamDecls   map[string]struct{}
}

func NewHoister(resolver *Resolver) *Hoister {
	return &Hoister{
		resolver:          resolver,
		kind:              DeclKindVar,
		inBlock:           false,
		inCatchBody:       false,
		excludedFromCatch: make(map[string]struct{}),
		catchParamDecls:   make(map[string]struct{}),
	}
}

func (h *Hoister) addPatId(id *ast.Identifier) {
	if h.inCatchBody {
		if _, ok := h.catchParamDecls[id.Name]; ok {
			r, _ := h.resolver.markForRef(id.Name)
			if r != UnresolvedMark {
				return
			}
		}
		h.excludedFromCatch[id.Name] = struct{}{}
	} else {
		if _, ok := h.catchParamDecls[id.Name]; ok {
			if _, excluded := h.excludedFromCatch[id.Name]; !excluded {
				return
			}
		}
	}
	h.resolver.modify(id, h.kind)
}

func (h *Hoister) VisitVariableStatement(n *ast.VariableDeclaration) {
	if h.inBlock {
		//if n.Kind != ast.VarKind {
		//	return
		//}
	}

	oldKind := h.kind
	h.kind = DeclKindVar
	for _, decl := range n.List {
		h.VisitVariableDeclarator(decl)
	}
	h.kind = oldKind
}

func (h *Hoister) VisitVariableDeclarator(n *ast.VariableDeclarator) {
	if ident, ok := n.Target.Target.(*ast.Identifier); ok {
		h.addIdent(ident)
	}
}

func (h *Hoister) addIdent(id *ast.Identifier) {
	if h.inCatchBody {
		if _, ok := h.catchParamDecls[id.Name]; ok {
			r, _ := h.resolver.markForRef(id.Name)
			if r != UnresolvedMark {
				return
			}
		}
		h.excludedFromCatch[id.Name] = struct{}{}
	} else {
		if _, ok := h.catchParamDecls[id.Name]; ok {
			if _, excluded := h.excludedFromCatch[id.Name]; !excluded {
				return
			}
		}
	}
	h.resolver.modify(id, h.kind)
}

func (h *Hoister) VisitFunctionDeclaration(n *ast.FunctionDeclaration) {
	if _, ok := h.catchParamDecls[n.Function.Name.Name]; ok {
		return
	}

	if h.inBlock {
		if symbol := h.resolver.current.findVarInfo(n.Function.Name.Name); symbol != nil {
			//if symbol.DeclType == DeclKindLexical || symbol.DeclType == DeclKindParam {
			//	return
			//}
		}
	}

	h.resolver.modify(n.Function.Name, DeclKindFunction)
}
