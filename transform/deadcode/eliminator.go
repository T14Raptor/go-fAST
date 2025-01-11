package deadcode

import (
	"go/token"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/tools/fastgraph"
)

type TreeShaker struct {
	ast.NoopVisitor

	changed bool
	pass    int

	inFunc      bool
	inBlockStmt bool
	varDeclKind token.Token

	data Data

	bindings map[ast.Id]struct{}
}

type Data struct {
	usedNames map[ast.Id]VarInfo
	graph     fastgraph.DirectedGraph[int, VarInfo]
	entries   map[int]struct{}
	graphIx   map[ast.Id]int
}

type VarInfo struct {
	// This does not include self-references in a function.
	Usage int
	// This does not include self-references in a function.
	Assign int
}
