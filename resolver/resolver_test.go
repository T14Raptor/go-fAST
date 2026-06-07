package resolver_test

import (
	"testing"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/parser"
	"github.com/t14raptor/go-fast/resolver"
)

type idVisitor struct {
	ast.NoopVisitor
	name string
	ids  []ast.Id
}

func idsForName(t *testing.T, src, name string) []ast.Id {
	t.Helper()

	program, err := parser.Parse(src)
	if err != nil {
		t.Fatal(err)
	}

	resolver.Resolve(program)

	visitor := &idVisitor{name: name}
	visitor.V = visitor
	program.VisitWith(visitor)
	return visitor.ids
}

func (v *idVisitor) VisitIdentifier(n *ast.Identifier) {
	if n.Name == v.name {
		v.ids = append(v.ids, n.ToId())
	}
}

func requireIDCount(t *testing.T, ids []ast.Id, want int) {
	t.Helper()
	if len(ids) != want {
		t.Fatalf("expected %d identifiers, got %d: %+v", want, len(ids), ids)
	}
}

func TestFunctionRestParameterCreatesFunctionScopeBinding(t *testing.T) {
	ids := idsForName(t, `function P8() { function inner(...P8) { P8.push(undefined); } P8(); }`, "P8")
	requireIDCount(t, ids, 4)

	outer := ids[0]
	rest := ids[1]
	use := ids[2]
	outerUse := ids[3]

	if rest.ScopeContext == outer.ScopeContext {
		t.Fatalf("rest parameter resolved to outer binding: rest=%+v outer=%+v", rest, outer)
	}
	if use.ScopeContext != rest.ScopeContext {
		t.Fatalf("rest parameter use resolved to wrong binding: use=%+v rest=%+v", use, rest)
	}
	if outerUse.ScopeContext != outer.ScopeContext {
		t.Fatalf("outer use resolved to wrong binding: use=%+v outer=%+v", outerUse, outer)
	}
}

func TestArrowRestParameterCreatesFunctionScopeBinding(t *testing.T) {
	ids := idsForName(t, `let rest; const fn = (...rest) => rest; rest;`, "rest")
	requireIDCount(t, ids, 4)

	outer := ids[0]
	rest := ids[1]
	use := ids[2]
	outerUse := ids[3]

	if rest.ScopeContext == outer.ScopeContext {
		t.Fatalf("arrow rest parameter resolved to outer binding: rest=%+v outer=%+v", rest, outer)
	}
	if use.ScopeContext != rest.ScopeContext {
		t.Fatalf("arrow rest use resolved to wrong binding: use=%+v rest=%+v", use, rest)
	}
	if outerUse.ScopeContext != outer.ScopeContext {
		t.Fatalf("outer rest use resolved to wrong binding: use=%+v outer=%+v", outerUse, outer)
	}
}

func TestDestructuredParameterCreatesFunctionScopeBinding(t *testing.T) {
	xs := idsForName(t, `let x; const fn = ([x]) => x; x;`, "x")
	requireIDCount(t, xs, 4)

	outer := xs[0]
	param := xs[1]
	use := xs[2]
	outerUse := xs[3]

	if param.ScopeContext == outer.ScopeContext {
		t.Fatalf("destructured parameter resolved to outer binding: param=%+v outer=%+v", param, outer)
	}
	if use.ScopeContext != param.ScopeContext {
		t.Fatalf("destructured parameter use resolved to wrong binding: use=%+v param=%+v", use, param)
	}
	if outerUse.ScopeContext != outer.ScopeContext {
		t.Fatalf("outer destructured-name use resolved to wrong binding: use=%+v outer=%+v", outerUse, outer)
	}

	ys := idsForName(t, `let y; const fn = ({a: y}) => y; y;`, "y")
	requireIDCount(t, ys, 4)

	outer = ys[0]
	param = ys[1]
	use = ys[2]
	outerUse = ys[3]

	if param.ScopeContext == outer.ScopeContext {
		t.Fatalf("object destructured parameter resolved to outer binding: param=%+v outer=%+v", param, outer)
	}
	if use.ScopeContext != param.ScopeContext {
		t.Fatalf("object destructured parameter use resolved to wrong binding: use=%+v param=%+v", use, param)
	}
	if outerUse.ScopeContext != outer.ScopeContext {
		t.Fatalf("outer object destructured-name use resolved to wrong binding: use=%+v outer=%+v", outerUse, outer)
	}
}

func TestDestructuredParameterComputedKeyAndDefaultScopes(t *testing.T) {
	xs := idsForName(t, `let k, x; const fn = ({[k]: x}) => x; k; x;`, "x")
	requireIDCount(t, xs, 4)

	outerX := xs[0]
	paramX := xs[1]
	bodyX := xs[2]
	outerUseX := xs[3]

	if paramX.ScopeContext == outerX.ScopeContext {
		t.Fatalf("computed-key destructured binding resolved to outer scope: param=%+v outer=%+v", paramX, outerX)
	}
	if bodyX.ScopeContext != paramX.ScopeContext {
		t.Fatalf("computed-key body use resolved to wrong binding: use=%+v param=%+v", bodyX, paramX)
	}
	if outerUseX.ScopeContext != outerX.ScopeContext {
		t.Fatalf("outer computed-key binding use resolved to wrong binding: use=%+v outer=%+v", outerUseX, outerX)
	}

	ks := idsForName(t, `let k, x; const fn = ({[k]: x}) => x; k; x;`, "k")
	requireIDCount(t, ks, 3)

	outerK := ks[0]
	computedK := ks[1]
	outerUseK := ks[2]

	if computedK.ScopeContext != outerK.ScopeContext {
		t.Fatalf("computed key should resolve as reference to outer binding: key=%+v outer=%+v", computedK, outerK)
	}
	if outerUseK.ScopeContext != outerK.ScopeContext {
		t.Fatalf("outer computed key use resolved to wrong binding: use=%+v outer=%+v", outerUseK, outerK)
	}

	ys := idsForName(t, `let x, y; const fn = ({x = y}) => x; y;`, "y")
	requireIDCount(t, ys, 3)

	outerY := ys[0]
	defaultY := ys[1]
	outerUseY := ys[2]

	if defaultY.ScopeContext != outerY.ScopeContext {
		t.Fatalf("default initializer should resolve as reference to outer binding: init=%+v outer=%+v", defaultY, outerY)
	}
	if outerUseY.ScopeContext != outerY.ScopeContext {
		t.Fatalf("outer default-name use resolved to wrong binding: use=%+v outer=%+v", outerUseY, outerY)
	}
}

func TestParameterDefaultCanReferenceLaterParameterBinding(t *testing.T) {
	ids := idsForName(t, `let b; function f(a = b, b) { return b; } b;`, "b")
	requireIDCount(t, ids, 5)

	outer := ids[0]
	defaultUse := ids[1]
	param := ids[2]
	bodyUse := ids[3]
	outerUse := ids[4]

	if param.ScopeContext == outer.ScopeContext {
		t.Fatalf("later parameter resolved to outer binding: param=%+v outer=%+v", param, outer)
	}
	if defaultUse.ScopeContext != param.ScopeContext {
		t.Fatalf("default initializer resolved to wrong binding: use=%+v param=%+v", defaultUse, param)
	}
	if bodyUse.ScopeContext != param.ScopeContext {
		t.Fatalf("body use resolved to wrong binding: use=%+v param=%+v", bodyUse, param)
	}
	if outerUse.ScopeContext != outer.ScopeContext {
		t.Fatalf("outer later-parameter-name use resolved to wrong binding: use=%+v outer=%+v", outerUse, outer)
	}
}

func TestCatchParameterCreatesCatchScopeBinding(t *testing.T) {
	ids := idsForName(t, `let e; try { throw 1; } catch (e) { e; } e;`, "e")
	requireIDCount(t, ids, 4)

	outer := ids[0]
	param := ids[1]
	bodyUse := ids[2]
	outerUse := ids[3]

	if param.ScopeContext == outer.ScopeContext {
		t.Fatalf("catch parameter resolved to outer binding: param=%+v outer=%+v", param, outer)
	}
	if bodyUse.ScopeContext != param.ScopeContext {
		t.Fatalf("catch body use resolved to wrong binding: use=%+v param=%+v", bodyUse, param)
	}
	if outerUse.ScopeContext != outer.ScopeContext {
		t.Fatalf("outer catch-name use resolved to wrong binding: use=%+v outer=%+v", outerUse, outer)
	}
}

func TestCatchParameterDoesNotLeakIntoLaterHoisting(t *testing.T) {
	ids := idsForName(t, `let e; function f() { try { throw 1; } catch (e) {} e; var e; } e;`, "e")
	requireIDCount(t, ids, 5)

	outer := ids[0]
	catchParam := ids[1]
	preUse := ids[2]
	varBinding := ids[3]
	outerUse := ids[4]

	if catchParam.ScopeContext == outer.ScopeContext {
		t.Fatalf("catch parameter resolved to outer binding: param=%+v outer=%+v", catchParam, outer)
	}
	if varBinding.ScopeContext == outer.ScopeContext || varBinding.ScopeContext == catchParam.ScopeContext {
		t.Fatalf("var binding resolved to wrong scope: var=%+v outer=%+v catch=%+v", varBinding, outer, catchParam)
	}
	if preUse.ScopeContext != varBinding.ScopeContext {
		t.Fatalf("post-catch pre-var use was not hoisted to var binding: use=%+v var=%+v", preUse, varBinding)
	}
	if outerUse.ScopeContext != outer.ScopeContext {
		t.Fatalf("outer catch-name use resolved to wrong binding: use=%+v outer=%+v", outerUse, outer)
	}
}

func TestCatchBodyVarWithExistingOuterVarResolvesToOuterVar(t *testing.T) {
	ids := idsForName(t, `function f() { var e; try { throw 1; } catch (e) { var e; } e; }`, "e")
	requireIDCount(t, ids, 4)

	functionVar := ids[0]
	catchParam := ids[1]
	catchVar := ids[2]
	postCatchUse := ids[3]

	if catchParam.ScopeContext == functionVar.ScopeContext {
		t.Fatalf("catch parameter resolved to function var: param=%+v var=%+v", catchParam, functionVar)
	}
	if catchVar.ScopeContext != functionVar.ScopeContext {
		t.Fatalf("catch body var resolved outside function var: catchVar=%+v functionVar=%+v", catchVar, functionVar)
	}
	if postCatchUse.ScopeContext != functionVar.ScopeContext {
		t.Fatalf("post-catch use resolved outside function var: use=%+v functionVar=%+v", postCatchUse, functionVar)
	}
}

func TestDestructuredCatchParameterDoesNotHoistOutsideCatch(t *testing.T) {
	ids := idsForName(t, `let e; try { throw {}; } catch ({e}) { e; } e;`, "e")
	requireIDCount(t, ids, 4)

	outer := ids[0]
	param := ids[1]
	bodyUse := ids[2]
	outerUse := ids[3]

	if param.ScopeContext == outer.ScopeContext {
		t.Fatalf("destructured catch parameter resolved to outer binding: param=%+v outer=%+v", param, outer)
	}
	if bodyUse.ScopeContext != param.ScopeContext {
		t.Fatalf("destructured catch body use resolved to wrong binding: use=%+v param=%+v", bodyUse, param)
	}
	if outerUse.ScopeContext != outer.ScopeContext {
		t.Fatalf("outer destructured catch-name use resolved to wrong binding: use=%+v outer=%+v", outerUse, outer)
	}
}

func TestDestructuredVarDeclarationIsHoisted(t *testing.T) {
	ids := idsForName(t, `var x; function f(o) { x; var {x} = o; }`, "x")
	requireIDCount(t, ids, 3)

	outer := ids[0]
	preUse := ids[1]
	binding := ids[2]

	if binding.ScopeContext == outer.ScopeContext {
		t.Fatalf("destructured var binding resolved to outer scope: binding=%+v outer=%+v", binding, outer)
	}
	if preUse.ScopeContext != binding.ScopeContext {
		t.Fatalf("pre-declaration var use was not hoisted: use=%+v binding=%+v", preUse, binding)
	}
}
