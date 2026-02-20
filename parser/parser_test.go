package parser_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/generator"
	"github.com/t14raptor/go-fast/parser"
	"github.com/t14raptor/go-fast/token"
)

func TestIssue26(t *testing.T) {
	code := `const a = {}
const c = { a: 1 }
for (a.b in c) {
  console.log(a.b)
}`
	_, err := parser.ParseFile(code)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// mustParse parses code and fails the test if there's an error.
func mustParse(t *testing.T, code string) *ast.Program {
	t.Helper()
	p, err := parser.ParseFile(code)
	if err != nil {
		t.Fatalf("Failed to parse:\n%s\nError: %v", code, err)
	}
	return p
}

// roundTrip parses code, regenerates it, and returns the output.
func roundTrip(t *testing.T, code string) string {
	t.Helper()
	p := mustParse(t, code)
	return strings.TrimSpace(generator.Generate(p))
}

// assertRoundTrip parses code, regenerates it, and checks that the output
// matches the expected string.
func assertRoundTrip(t *testing.T, code, want string) {
	t.Helper()
	got := roundTrip(t, code)
	if got != want {
		t.Errorf("roundTrip(%q)\n  got:  %s\n  want: %s", code, got, want)
	}
}

// firstStmt returns the concrete statement node from the i-th top-level statement.
func firstStmt(p *ast.Program, i int) ast.VisitableNode {
	return p.Body[i].Unwrap()
}

// exprOf extracts the inner concrete expression from an ExpressionStatement.
func exprOf(s ast.VisitableNode) ast.VisitableNode {
	return s.(*ast.ExpressionStatement).Expression.Unwrap()
}

// initializerExpr extracts the initializer expression from the first
// VariableDeclarator of a VariableDeclaration statement.
func initializerExpr(s ast.VisitableNode) ast.VisitableNode {
	init := s.(*ast.VariableDeclaration).List[0].Initializer
	if init == nil {
		return nil
	}
	return init.Unwrap()
}

// bodyOf extracts the BlockStatement body from a FunctionDeclaration.
func bodyOf(s ast.VisitableNode) *ast.BlockStatement {
	return s.(*ast.FunctionDeclaration).Function.Body
}

// ===========================================================================
// AST STRUCTURE VERIFICATION TESTS
// ===========================================================================

// ---------------------------------------------------------------------------
// Array literal Expressions slice — correct length, element types, values
// ---------------------------------------------------------------------------

func TestArrayLiteralAST(t *testing.T) {
	p := mustParse(t, "var a = [1, 'two', true, null]")
	arr := initializerExpr(firstStmt(p, 0)).(*ast.ArrayLiteral)

	if got := len(arr.Value); got != 4 {
		t.Fatalf("array length = %d; want 4", got)
	}

	if !arr.Value[0].IsNumLit() {
		t.Errorf("arr[0] kind = %v; want NumLit", arr.Value[0].Kind())
	}
	if n := arr.Value[0].MustNumLit(); n.Value != 1 {
		t.Errorf("arr[0] value = %v; want 1", n.Value)
	}
	if !arr.Value[1].IsStrLit() {
		t.Errorf("arr[1] kind = %v; want StrLit", arr.Value[1].Kind())
	}
	if s := arr.Value[1].MustStrLit(); s.Value != "two" {
		t.Errorf("arr[1] value = %q; want \"two\"", s.Value)
	}
	if !arr.Value[2].IsBoolLit() {
		t.Errorf("arr[2] kind = %v; want BoolLit", arr.Value[2].Kind())
	}
	if !arr.Value[3].IsNullLit() {
		t.Errorf("arr[3] kind = %v; want NullLit", arr.Value[3].Kind())
	}
}

func TestArrayLiteralElisionsAST(t *testing.T) {
	p := mustParse(t, "var a = [1,,2,,3]")
	arr := initializerExpr(firstStmt(p, 0)).(*ast.ArrayLiteral)

	if got := len(arr.Value); got != 5 {
		t.Fatalf("array length = %d; want 5", got)
	}
	// Positions 1 and 3 should be elisions (Expression with nil Expr).
	if !arr.Value[1].IsNone() {
		t.Errorf("arr[1] kind = %v; want None (elision)", arr.Value[1].Kind())
	}
	if !arr.Value[3].IsNone() {
		t.Errorf("arr[3] kind = %v; want None (elision)", arr.Value[3].Kind())
	}
	// Positions 0, 2, 4 should be NumberLiterals.
	for _, i := range []int{0, 2, 4} {
		if !arr.Value[i].IsNumLit() {
			t.Errorf("arr[%d] kind = %v; want NumLit", i, arr.Value[i].Kind())
		}
	}
}

func TestArrayLiteralSpreadAST(t *testing.T) {
	p := mustParse(t, "var a = [1, ...b, 2]")
	arr := initializerExpr(firstStmt(p, 0)).(*ast.ArrayLiteral)

	if got := len(arr.Value); got != 3 {
		t.Fatalf("array length = %d; want 3", got)
	}
	if !arr.Value[0].IsNumLit() {
		t.Errorf("arr[0] kind = %v; want NumLit", arr.Value[0].Kind())
	}
	if !arr.Value[1].IsSpread() {
		t.Fatalf("arr[1] kind = %v; want Spread", arr.Value[1].Kind())
	}
	spread := arr.Value[1].MustSpread()
	if !spread.Expression.IsIdent() || spread.Expression.MustIdent().Name != "b" {
		t.Errorf("spread target = %v; want identifier 'b'", spread.Expression.Kind())
	}
	if !arr.Value[2].IsNumLit() {
		t.Errorf("arr[2] kind = %v; want NumLit", arr.Value[2].Kind())
	}
}

func TestArrayLiteralLargeAST(t *testing.T) {
	n := 200
	var elems []string
	for i := 0; i < n; i++ {
		elems = append(elems, fmt.Sprintf("%d", i))
	}
	p := mustParse(t, "var a = ["+strings.Join(elems, ",")+"]")
	arr := initializerExpr(firstStmt(p, 0)).(*ast.ArrayLiteral)

	if got := len(arr.Value); got != n {
		t.Fatalf("array length = %d; want %d", got, n)
	}
	// Verify first, middle and last values.
	for _, idx := range []int{0, n / 2, n - 1} {
		if !arr.Value[idx].IsNumLit() {
			t.Errorf("arr[%d] kind = %v; want NumLit", idx, arr.Value[idx].Kind())
			continue
		}
		num := arr.Value[idx].MustNumLit()
		if num.Value != float64(idx) {
			t.Errorf("arr[%d] value = %v; want %d", idx, num.Value, idx)
		}
	}
}

// ---------------------------------------------------------------------------
// Argument list Expressions slice — correct length and element types
// ---------------------------------------------------------------------------

func TestArgumentListAST(t *testing.T) {
	p := mustParse(t, "f(1, 'a', true)")
	call := exprOf(firstStmt(p, 0)).(*ast.CallExpression)

	if got := len(call.ArgumentList); got != 3 {
		t.Fatalf("arg count = %d; want 3", got)
	}
	if !call.ArgumentList[0].IsNumLit() {
		t.Errorf("arg[0] kind = %v; want NumLit", call.ArgumentList[0].Kind())
	}
	if !call.ArgumentList[1].IsStrLit() {
		t.Errorf("arg[1] kind = %v; want StrLit", call.ArgumentList[1].Kind())
	}
	if !call.ArgumentList[2].IsBoolLit() {
		t.Errorf("arg[2] kind = %v; want BoolLit", call.ArgumentList[2].Kind())
	}
}

func TestArgumentListSpreadAST(t *testing.T) {
	p := mustParse(t, "f(1, ...args)")
	call := exprOf(firstStmt(p, 0)).(*ast.CallExpression)

	if got := len(call.ArgumentList); got != 2 {
		t.Fatalf("arg count = %d; want 2", got)
	}
	if !call.ArgumentList[0].IsNumLit() {
		t.Errorf("arg[0] kind = %v; want NumLit", call.ArgumentList[0].Kind())
	}
	if !call.ArgumentList[1].IsSpread() {
		t.Errorf("arg[1] kind = %v; want Spread", call.ArgumentList[1].Kind())
	}
}

func TestArgumentListEmptyAST(t *testing.T) {
	p := mustParse(t, "f()")
	call := exprOf(firstStmt(p, 0)).(*ast.CallExpression)

	if got := len(call.ArgumentList); got != 0 {
		t.Fatalf("arg count = %d; want 0", got)
	}
}

func TestNestedCallsAST(t *testing.T) {
	p := mustParse(t, "f(g(1, 2), h(3))")
	outer := exprOf(firstStmt(p, 0)).(*ast.CallExpression)

	if got := len(outer.ArgumentList); got != 2 {
		t.Fatalf("outer arg count = %d; want 2", got)
	}

	inner1 := outer.ArgumentList[0].MustCall()
	if got := len(inner1.ArgumentList); got != 2 {
		t.Errorf("inner1 arg count = %d; want 2", got)
	}

	inner2 := outer.ArgumentList[1].MustCall()
	if got := len(inner2.ArgumentList); got != 1 {
		t.Errorf("inner2 arg count = %d; want 1", got)
	}
}

// ---------------------------------------------------------------------------
// Sequence expression — correct length and element values
// ---------------------------------------------------------------------------

func TestSequenceExpressionAST(t *testing.T) {
	p := mustParse(t, "(1, 2, 3)")
	seq := exprOf(firstStmt(p, 0)).(*ast.SequenceExpression)

	if got := len(seq.Sequence); got != 3 {
		t.Fatalf("sequence length = %d; want 3", got)
	}
	for i, want := range []float64{1, 2, 3} {
		if !seq.Sequence[i].IsNumLit() {
			t.Errorf("seq[%d] kind = %v; want NumLit", i, seq.Sequence[i].Kind())
			continue
		}
		num := seq.Sequence[i].MustNumLit()
		if num.Value != want {
			t.Errorf("seq[%d] value = %v; want %v", i, num.Value, want)
		}
	}
}

func TestSequenceExpressionTwoAST(t *testing.T) {
	p := mustParse(t, "(10, 20)")
	seq := exprOf(firstStmt(p, 0)).(*ast.SequenceExpression)

	if got := len(seq.Sequence); got != 2 {
		t.Fatalf("sequence length = %d; want 2", got)
	}
	n1 := seq.Sequence[0].MustNumLit()
	n2 := seq.Sequence[1].MustNumLit()
	if n1.Value != 10 || n2.Value != 20 {
		t.Errorf("values = (%v, %v); want (10, 20)", n1.Value, n2.Value)
	}
}

// ---------------------------------------------------------------------------
// Template literal — correct element/expression counts and values
// ---------------------------------------------------------------------------

func TestTemplateLiteralAST(t *testing.T) {
	p := mustParse(t, "let x = `hello ${name} you are ${age} years old`")
	tmpl := initializerExpr(firstStmt(p, 0)).(*ast.TemplateLiteral)

	if got := len(tmpl.Expressions); got != 2 {
		t.Fatalf("expression count = %d; want 2", got)
	}
	if got := len(tmpl.Elements); got != 3 {
		t.Fatalf("element count = %d; want 3", got)
	}

	// Check that the expressions are identifiers with the right names.
	if !tmpl.Expressions[0].IsIdent() {
		t.Fatalf("expr[0] kind = %v; want Ident", tmpl.Expressions[0].Kind())
	}
	id1 := tmpl.Expressions[0].MustIdent()
	if id1.Name != "name" {
		t.Errorf("expr[0] name = %q; want \"name\"", id1.Name)
	}

	if !tmpl.Expressions[1].IsIdent() {
		t.Fatalf("expr[1] kind = %v; want Ident", tmpl.Expressions[1].Kind())
	}
	id2 := tmpl.Expressions[1].MustIdent()
	if id2.Name != "age" {
		t.Errorf("expr[1] name = %q; want \"age\"", id2.Name)
	}

	// Check template text parts.
	if tmpl.Elements[0].Parsed != "hello " {
		t.Errorf("elements[0] = %q; want \"hello \"", tmpl.Elements[0].Parsed)
	}
	if tmpl.Elements[1].Parsed != " you are " {
		t.Errorf("elements[1] = %q; want \" you are \"", tmpl.Elements[1].Parsed)
	}
	if tmpl.Elements[2].Parsed != " years old" {
		t.Errorf("elements[2] = %q; want \" years old\"", tmpl.Elements[2].Parsed)
	}
}

func TestTemplateLiteralNoSubstitutionAST(t *testing.T) {
	p := mustParse(t, "let x = `plain text`")
	tmpl := initializerExpr(firstStmt(p, 0)).(*ast.TemplateLiteral)

	if got := len(tmpl.Expressions); got != 0 {
		t.Fatalf("expression count = %d; want 0", got)
	}
	if got := len(tmpl.Elements); got != 1 {
		t.Fatalf("element count = %d; want 1", got)
	}
	if tmpl.Elements[0].Parsed != "plain text" {
		t.Errorf("elements[0] = %q; want \"plain text\"", tmpl.Elements[0].Parsed)
	}
}

func TestTemplateLiteralManySubstitutionsAST(t *testing.T) {
	p := mustParse(t, "let x = `${a}${b}${c}${d}${e}`")
	tmpl := initializerExpr(firstStmt(p, 0)).(*ast.TemplateLiteral)

	if got := len(tmpl.Expressions); got != 5 {
		t.Fatalf("expression count = %d; want 5", got)
	}
	if got := len(tmpl.Elements); got != 6 {
		t.Fatalf("element count = %d; want 6", got)
	}
	names := []string{"a", "b", "c", "d", "e"}
	for i, want := range names {
		if !tmpl.Expressions[i].IsIdent() {
			t.Errorf("expr[%d] kind = %v; want Ident", i, tmpl.Expressions[i].Kind())
			continue
		}
		id := tmpl.Expressions[i].MustIdent()
		if id.Name != want {
			t.Errorf("expr[%d] name = %q; want %q", i, id.Name, want)
		}
	}
}

func TestTemplateLiteralNestedAST(t *testing.T) {
	p := mustParse(t, "let x = `outer ${`inner ${y}`} end`")
	outer := initializerExpr(firstStmt(p, 0)).(*ast.TemplateLiteral)

	if got := len(outer.Expressions); got != 1 {
		t.Fatalf("outer expression count = %d; want 1", got)
	}
	if !outer.Expressions[0].IsTmplLit() {
		t.Fatalf("outer.expr[0] kind = %v; want TmplLit", outer.Expressions[0].Kind())
	}
	inner := outer.Expressions[0].MustTmplLit()
	if got := len(inner.Expressions); got != 1 {
		t.Fatalf("inner expression count = %d; want 1", got)
	}
	if !inner.Expressions[0].IsIdent() {
		t.Fatalf("inner.expr[0] kind = %v; want Ident", inner.Expressions[0].Kind())
	}
	id := inner.Expressions[0].MustIdent()
	if id.Name != "y" {
		t.Errorf("inner.expr[0] name = %q; want \"y\"", id.Name)
	}
}

func TestTaggedTemplateLiteralAST(t *testing.T) {
	p := mustParse(t, "tag`hello ${x} world`")
	tmpl := exprOf(firstStmt(p, 0)).(*ast.TemplateLiteral)

	if tmpl.Tag == nil {
		t.Fatal("tag is nil; want non-nil")
	}
	if !tmpl.Tag.IsIdent() {
		t.Fatalf("tag kind = %v; want Ident", tmpl.Tag.Kind())
	}
	tagId := tmpl.Tag.MustIdent()
	if tagId.Name != "tag" {
		t.Errorf("tag name = %q; want \"tag\"", tagId.Name)
	}
	if got := len(tmpl.Expressions); got != 1 {
		t.Fatalf("expression count = %d; want 1", got)
	}
}

// ---------------------------------------------------------------------------
// RegExp literal — pattern and flags preserved
// ---------------------------------------------------------------------------

func TestRegExpAST(t *testing.T) {
	tests := []struct {
		code    string
		pattern string
		flags   string
	}{
		{"var r = /abc/", "abc", ""},
		{"var r = /abc/gi", "abc", "gi"},
		{`var r = /^hello$/`, "^hello$", ""},
		{`var r = /\d+/g`, `\d+`, "g"},
		{`var r = /[a-z]+/i`, `[a-z]+`, "i"},
	}
	for _, tt := range tests {
		p := mustParse(t, tt.code)
		re := initializerExpr(firstStmt(p, 0)).(*ast.RegExpLiteral)
		if re.Pattern != tt.pattern {
			t.Errorf("pattern for %q = %q; want %q", tt.code, re.Pattern, tt.pattern)
		}
		if re.Flags != tt.flags {
			t.Errorf("flags for %q = %q; want %q", tt.code, re.Flags, tt.flags)
		}
	}
}

func TestRegExpInCallAST(t *testing.T) {
	p := mustParse(t, `str.replace(/foo/g, "bar")`)
	call := exprOf(firstStmt(p, 0)).(*ast.CallExpression)
	if got := len(call.ArgumentList); got != 2 {
		t.Fatalf("arg count = %d; want 2", got)
	}
	if !call.ArgumentList[0].IsRegExpLit() {
		t.Fatalf("arg[0] kind = %v; want RegExpLit", call.ArgumentList[0].Kind())
	}
	re := call.ArgumentList[0].MustRegExpLit()
	if re.Pattern != "foo" || re.Flags != "g" {
		t.Errorf("regex = /%s/%s; want /foo/g", re.Pattern, re.Flags)
	}
	if !call.ArgumentList[1].IsStrLit() {
		t.Fatalf("arg[1] kind = %v; want StrLit", call.ArgumentList[1].Kind())
	}
	str := call.ArgumentList[1].MustStrLit()
	if str.Value != "bar" {
		t.Errorf("arg[1] = %q; want \"bar\"", str.Value)
	}
}

// ---------------------------------------------------------------------------
// Statement lists — correct length and types (parseStatementList)
// ---------------------------------------------------------------------------

func TestBlockStatementListAST(t *testing.T) {
	p := mustParse(t, "function f() { var a = 1; var b = 2; var c = 3; return a + b + c; }")
	body := bodyOf(firstStmt(p, 0))

	if got := len(body.List); got != 4 {
		t.Fatalf("statement count = %d; want 4", got)
	}
	for i := 0; i < 3; i++ {
		if !body.List[i].IsVarDecl() {
			t.Errorf("stmt[%d] kind = %s; want VarDecl", i, body.List[i].Kind())
		}
	}
	if !body.List[3].IsReturn() {
		t.Errorf("stmt[3] kind = %s; want Return", body.List[3].Kind())
	}
}

func TestBlockStatementEmptyAST(t *testing.T) {
	p := mustParse(t, "function f() {}")
	body := bodyOf(firstStmt(p, 0))

	if got := len(body.List); got != 0 {
		t.Fatalf("statement count = %d; want 0", got)
	}
}

func TestBlockStatementNestedAST(t *testing.T) {
	p := mustParse(t, "function f() { if (true) { var x = 1; var y = 2; } }")
	body := bodyOf(firstStmt(p, 0))

	if got := len(body.List); got != 1 {
		t.Fatalf("outer statement count = %d; want 1", got)
	}
	if !body.List[0].IsIf() {
		t.Fatalf("stmt[0] kind = %v; want If", body.List[0].Kind())
	}
	ifStmt := body.List[0].MustIf()
	if !ifStmt.Consequent.IsBlock() {
		t.Fatalf("if body kind = %v; want Block", ifStmt.Consequent.Kind())
	}
	block := ifStmt.Consequent.MustBlock()
	if got := len(block.List); got != 2 {
		t.Fatalf("inner statement count = %d; want 2", got)
	}
}

func TestBlockStatementLargeAST(t *testing.T) {
	n := 150
	var stmts []string
	for i := 0; i < n; i++ {
		stmts = append(stmts, fmt.Sprintf("var x%d = %d;", i, i))
	}
	p := mustParse(t, "function f() { "+strings.Join(stmts, " ")+" }")
	body := bodyOf(firstStmt(p, 0))

	if got := len(body.List); got != n {
		t.Fatalf("statement count = %d; want %d", got, n)
	}
	// Spot-check first and last.
	decl0 := body.List[0].MustVarDecl()
	if decl0.List[0].Target.MustIdent().Name != "x0" {
		t.Errorf("first decl name = %q; want \"x0\"", decl0.List[0].Target.MustIdent().Name)
	}
	declN := body.List[n-1].MustVarDecl()
	want := fmt.Sprintf("x%d", n-1)
	if declN.List[0].Target.MustIdent().Name != want {
		t.Errorf("last decl name = %q; want %q", declN.List[0].Target.MustIdent().Name, want)
	}
}

// ---------------------------------------------------------------------------
// Source elements (top-level statements) — parseSourceElements
// ---------------------------------------------------------------------------

func TestSourceElementsAST(t *testing.T) {
	p := mustParse(t, `
		var x = 1;
		function f() { return x; }
		class A {}
		var y = 2;
	`)

	if got := len(p.Body); got != 4 {
		t.Fatalf("body length = %d; want 4", got)
	}
	if !p.Body[0].IsVarDecl() {
		t.Errorf("body[0] kind = %v; want VarDecl", p.Body[0].Kind())
	}
	if !p.Body[1].IsFuncDecl() {
		t.Errorf("body[1] kind = %v; want FuncDecl", p.Body[1].Kind())
	}
	if !p.Body[2].IsClassDecl() {
		t.Errorf("body[2] kind = %v; want ClassDecl", p.Body[2].Kind())
	}
	if !p.Body[3].IsVarDecl() {
		t.Errorf("body[3] kind = %v; want VarDecl", p.Body[3].Kind())
	}
}

// ---------------------------------------------------------------------------
// Switch / case — Consequent statement slice and case count
// ---------------------------------------------------------------------------

func TestSwitchCaseConsequentAST(t *testing.T) {
	p := mustParse(t, `switch (x) { case 1: a(); b(); c(); break; case 2: d(); default: e(); f(); }`)
	sw := firstStmt(p, 0).(*ast.SwitchStatement)

	if got := len(sw.Body); got != 3 {
		t.Fatalf("case count = %d; want 3", got)
	}

	// case 1: 4 statements (a(), b(), c(), break)
	if got := len(sw.Body[0].Consequent); got != 4 {
		t.Errorf("case 1 consequent = %d; want 4", got)
	}
	// Verify first statement in case 1 is expression statement calling 'a'.
	if !sw.Body[0].Consequent[0].IsExpression() {
		t.Fatalf("case1.stmt[0] kind = %v; want Expression", sw.Body[0].Consequent[0].Kind())
	}
	es := sw.Body[0].Consequent[0].MustExpression()
	if !es.Expression.IsCall() {
		t.Fatalf("case1.stmt[0].expr kind = %v; want Call", es.Expression.Kind())
	}
	call := es.Expression.MustCall()
	if id := call.Callee.MustIdent(); id.Name != "a" {
		t.Errorf("case1.stmt[0] callee = %q; want \"a\"", id.Name)
	}
	// Last statement should be break.
	if !sw.Body[0].Consequent[3].IsBreak() {
		t.Errorf("case1.stmt[3] kind = %v; want Break", sw.Body[0].Consequent[3].Kind())
	}

	// case 2: 1 statement (d())
	if got := len(sw.Body[1].Consequent); got != 1 {
		t.Errorf("case 2 consequent = %d; want 1", got)
	}

	// default: 2 statements (e(), f())
	if got := len(sw.Body[2].Consequent); got != 2 {
		t.Errorf("default consequent = %d; want 2", got)
	}
	if sw.Body[2].Test != nil {
		t.Errorf("default test should be nil")
	}
}

func TestSwitchCaseTestExpressionsAST(t *testing.T) {
	p := mustParse(t, `switch (x) { case "a": break; case "b": break; }`)
	sw := firstStmt(p, 0).(*ast.SwitchStatement)

	if got := len(sw.Body); got != 2 {
		t.Fatalf("case count = %d; want 2", got)
	}
	s1 := sw.Body[0].Test.MustStrLit()
	if s1.Value != "a" {
		t.Errorf("case 0 test = %q; want \"a\"", s1.Value)
	}
	s2 := sw.Body[1].Test.MustStrLit()
	if s2.Value != "b" {
		t.Errorf("case 1 test = %q; want \"b\"", s2.Value)
	}
}

// ---------------------------------------------------------------------------
// Object literal — Properties slice, keys/values
// ---------------------------------------------------------------------------

func TestObjectLiteralAST(t *testing.T) {
	p := mustParse(t, "var o = { a: 1, b: 'two', c: true }")
	obj := initializerExpr(firstStmt(p, 0)).(*ast.ObjectLiteral)

	if got := len(obj.Value); got != 3 {
		t.Fatalf("property count = %d; want 3", got)
	}

	// First property: a: 1
	p1 := obj.Value[0].MustKeyed()
	if id := p1.Key.MustStrLit(); id.Value != "a" {
		t.Errorf("prop[0] key = %q; want \"a\"", id.Value)
	}
	if n := p1.Value.MustNumLit(); n.Value != 1 {
		t.Errorf("prop[0] value = %v; want 1", n.Value)
	}

	// Second property: b: 'two'
	p2 := obj.Value[1].MustKeyed()
	if s := p2.Value.MustStrLit(); s.Value != "two" {
		t.Errorf("prop[1] value = %q; want \"two\"", s.Value)
	}

	// Third property: c: true
	p3 := obj.Value[2].MustKeyed()
	if b := p3.Value.MustBoolLit(); !b.Value {
		t.Errorf("prop[2] value = %v; want true", b.Value)
	}
}

func TestObjectLiteralShorthandAST(t *testing.T) {
	p := mustParse(t, "var o = { x, y, z }")
	obj := initializerExpr(firstStmt(p, 0)).(*ast.ObjectLiteral)

	if got := len(obj.Value); got != 3 {
		t.Fatalf("property count = %d; want 3", got)
	}
	names := []string{"x", "y", "z"}
	for i, want := range names {
		ps, ok := obj.Value[i].Short()
		if !ok {
			t.Errorf("prop[%d] kind = %v; want PropShort", i, obj.Value[i].Kind())
			continue
		}
		if ps.Name.Name != want {
			t.Errorf("prop[%d] name = %q; want %q", i, ps.Name.Name, want)
		}
	}
}

func TestObjectLiteralSpreadAST(t *testing.T) {
	p := mustParse(t, "var o = { a: 1, ...b, c: 2 }")
	obj := initializerExpr(firstStmt(p, 0)).(*ast.ObjectLiteral)

	if got := len(obj.Value); got != 3 {
		t.Fatalf("property count = %d; want 3", got)
	}
	if _, ok := obj.Value[0].Keyed(); !ok {
		t.Errorf("prop[0] kind = %v; want PropKeyed", obj.Value[0].Kind())
	}
	spread, ok := obj.Value[1].Spread()
	if !ok {
		t.Fatalf("prop[1] kind = %v; want PropSpread", obj.Value[1].Kind())
	}
	if id := spread.Expression.MustIdent(); id.Name != "b" {
		t.Errorf("spread target = %q; want \"b\"", id.Name)
	}
	if _, ok := obj.Value[2].Keyed(); !ok {
		t.Errorf("prop[2] kind = %v; want PropKeyed", obj.Value[2].Kind())
	}
}

// ---------------------------------------------------------------------------
// For loop — test/update expressions (arena-wrapped even when nil)
// ---------------------------------------------------------------------------

func TestForStatementFullAST(t *testing.T) {
	p := mustParse(t, "for (var i = 0; i < 10; i++) {}")
	forStmt := firstStmt(p, 0).(*ast.ForStatement)

	if forStmt.Initializer == nil {
		t.Fatal("initializer is nil; want non-nil")
	}
	if forStmt.Test == nil {
		t.Fatal("test is nil; want non-nil")
	}
	if forStmt.Update == nil {
		t.Fatal("update is nil; want non-nil")
	}

	// Test should be a binary expression: i < 10
	if !forStmt.Test.IsBinary() {
		t.Fatalf("test kind = %v; want Binary", forStmt.Test.Kind())
	}
	bin := forStmt.Test.MustBinary()
	if bin.Operator != token.Less {
		t.Errorf("test op = %v; want <", bin.Operator)
	}

	// Update should be an update expression: i++
	if !forStmt.Update.IsUpdate() {
		t.Fatalf("update kind = %v; want Update", forStmt.Update.Kind())
	}
	upd := forStmt.Update.MustUpdate()
	if !upd.Postfix {
		t.Errorf("update postfix = false; want true")
	}
}

func TestForStatementEmptyAST(t *testing.T) {
	p := mustParse(t, "for (;;) {}")
	forStmt := firstStmt(p, 0).(*ast.ForStatement)

	if forStmt.Initializer != nil {
		t.Errorf("initializer should be nil")
	}
	// Test and Update are *Expression wrappers (may have nil inner Expr).
	if forStmt.Test != nil && !forStmt.Test.IsNone() {
		t.Errorf("test should be nil or have nil Expr")
	}
	if forStmt.Update != nil && !forStmt.Update.IsNone() {
		t.Errorf("update should be nil or have nil Expr")
	}
}

// ---------------------------------------------------------------------------
// Variable declarations — declarator list
// ---------------------------------------------------------------------------

func TestVariableDeclarationAST(t *testing.T) {
	p := mustParse(t, "let x = 1, y = 2, z = 3")
	decl := firstStmt(p, 0).(*ast.VariableDeclaration)

	if got := len(decl.List); got != 3 {
		t.Fatalf("declarator count = %d; want 3", got)
	}
	names := []string{"x", "y", "z"}
	values := []float64{1, 2, 3}
	for i := range names {
		id := decl.List[i].Target.MustIdent()
		if id.Name != names[i] {
			t.Errorf("decl[%d] name = %q; want %q", i, id.Name, names[i])
		}
		num := decl.List[i].Initializer.MustNumLit()
		if num.Value != values[i] {
			t.Errorf("decl[%d] value = %v; want %v", i, num.Value, values[i])
		}
	}
}

// ---------------------------------------------------------------------------
// If/else — correct structure with alternates
// ---------------------------------------------------------------------------

func TestIfElseChainAST(t *testing.T) {
	p := mustParse(t, "if (a) { x(); } else if (b) { y(); } else { z(); }")
	ifStmt := firstStmt(p, 0).(*ast.IfStatement)

	// Test should be identifier 'a'.
	if id := ifStmt.Test.MustIdent(); id.Name != "a" {
		t.Errorf("if test = %q; want \"a\"", id.Name)
	}
	if ifStmt.Alternate == nil {
		t.Fatal("alternate is nil; want else-if chain")
	}

	// Alternate should be another IfStatement (else if).
	if !ifStmt.Alternate.IsIf() {
		t.Fatalf("alternate kind = %v; want If", ifStmt.Alternate.Kind())
	}
	elseIf := ifStmt.Alternate.MustIf()
	if id := elseIf.Test.MustIdent(); id.Name != "b" {
		t.Errorf("else-if test = %q; want \"b\"", id.Name)
	}
	if elseIf.Alternate == nil {
		t.Fatal("else-if alternate is nil; want else block")
	}

	// Final else should be a BlockStatement.
	if !elseIf.Alternate.IsBlock() {
		t.Errorf("final else kind = %v; want Block", elseIf.Alternate.Kind())
	}
}

// ---------------------------------------------------------------------------
// Arrow functions — parameter list and body
// ---------------------------------------------------------------------------

func TestArrowFunctionAST(t *testing.T) {
	p := mustParse(t, "var f = (a, b, c) => a + b + c")
	arrow := initializerExpr(firstStmt(p, 0)).(*ast.ArrowFunctionLiteral)

	if got := len(arrow.ParameterList.List); got != 3 {
		t.Fatalf("param count = %d; want 3", got)
	}
	names := []string{"a", "b", "c"}
	for i, want := range names {
		id := arrow.ParameterList.List[i].Target.MustIdent()
		if id.Name != want {
			t.Errorf("param[%d] = %q; want %q", i, id.Name, want)
		}
	}
}

// ---------------------------------------------------------------------------
// Class — body elements
// ---------------------------------------------------------------------------

func TestClassBodyAST(t *testing.T) {
	p := mustParse(t, `class A {
		x = 1;
		constructor() {}
		method() {}
		get prop() { return 1; }
	}`)
	cls := firstStmt(p, 0).(*ast.ClassDeclaration).Class

	if got := len(cls.Body); got != 4 {
		t.Fatalf("class body length = %d; want 4", got)
	}
}

// ---------------------------------------------------------------------------
// Conditional expression — ternary structure
// ---------------------------------------------------------------------------

func TestConditionalExpressionAST(t *testing.T) {
	p := mustParse(t, "var r = a ? 1 : 2")
	cond := initializerExpr(firstStmt(p, 0)).(*ast.ConditionalExpression)

	if id := cond.Test.MustIdent(); id.Name != "a" {
		t.Errorf("test = %q; want \"a\"", id.Name)
	}
	if n := cond.Consequent.MustNumLit(); n.Value != 1 {
		t.Errorf("consequent = %v; want 1", n.Value)
	}
	if n := cond.Alternate.MustNumLit(); n.Value != 2 {
		t.Errorf("alternate = %v; want 2", n.Value)
	}
}

// ---------------------------------------------------------------------------
// Try/catch/finally — structure verification
// ---------------------------------------------------------------------------

func TestTryCatchFinallyAST(t *testing.T) {
	p := mustParse(t, "try { a(); b(); } catch (e) { c(); } finally { d(); }")
	tr := firstStmt(p, 0).(*ast.TryStatement)

	if got := len(tr.Body.List); got != 2 {
		t.Errorf("try body statements = %d; want 2", got)
	}
	if tr.Catch == nil {
		t.Fatal("catch is nil; want non-nil")
	}
	if got := len(tr.Catch.Body.List); got != 1 {
		t.Errorf("catch body statements = %d; want 1", got)
	}
	catchParam := tr.Catch.Parameter.MustIdent()
	if catchParam.Name != "e" {
		t.Errorf("catch param = %q; want \"e\"", catchParam.Name)
	}
	if tr.Finally == nil {
		t.Fatal("finally is nil; want non-nil")
	}
	if got := len(tr.Finally.List); got != 1 {
		t.Errorf("finally body statements = %d; want 1", got)
	}
}

// ---------------------------------------------------------------------------
// Round-trip tests — verify output is semantically correct
// ---------------------------------------------------------------------------

func TestRoundTripArray(t *testing.T) {
	assertRoundTrip(t, "var a = [1, 2, 3]", "var a = [1, 2, 3];")
}

func TestRoundTripTemplate(t *testing.T) {
	assertRoundTrip(t, "var x = `hello ${name}`", "var x = `hello ${name}`;")
}

func TestRoundTripRegExp(t *testing.T) {
	assertRoundTrip(t, "var r = /abc/gi", "var r = /abc/gi;")
}

func TestRoundTripArrow(t *testing.T) {
	assertRoundTrip(t, "var f = (x, y) => x + y", "var f = (x, y) => x + y;")
}

func TestRoundTripForLoop(t *testing.T) {
	assertRoundTrip(t, "for (var i = 0; i < 10; i++) {}", "for (var i = 0; i < 10; i++) {}")
}

func TestRoundTripSwitch(t *testing.T) {
	got := roundTrip(t, `switch (x) { case 1: a(); break; default: b(); }`)
	if !strings.Contains(got, "case 1:") {
		t.Errorf("output missing 'case 1:': %s", got)
	}
	if !strings.Contains(got, "default:") {
		t.Errorf("output missing 'default:': %s", got)
	}
}

// ---------------------------------------------------------------------------
// Deeply nested / stress tests — verifies scratch buffer nesting is correct
// ---------------------------------------------------------------------------

func TestDeeplyNestedArraysAST(t *testing.T) {
	p := mustParse(t, "var a = [[1, 2], [3, [4, [5]]]]")
	outer := initializerExpr(firstStmt(p, 0)).(*ast.ArrayLiteral)
	if got := len(outer.Value); got != 2 {
		t.Fatalf("outer length = %d; want 2", got)
	}

	inner0 := outer.Value[0].MustArrLit()
	if got := len(inner0.Value); got != 2 {
		t.Errorf("inner0 length = %d; want 2", got)
	}

	inner1 := outer.Value[1].MustArrLit()
	if got := len(inner1.Value); got != 2 {
		t.Errorf("inner1 length = %d; want 2", got)
	}

	// inner1[1] should be [4, [5]]
	nested := inner1.Value[1].MustArrLit()
	if got := len(nested.Value); got != 2 {
		t.Errorf("nested length = %d; want 2", got)
	}
	deepest := nested.Value[1].MustArrLit()
	if got := len(deepest.Value); got != 1 {
		t.Errorf("deepest length = %d; want 1", got)
	}
	if n := deepest.Value[0].MustNumLit(); n.Value != 5 {
		t.Errorf("deepest value = %v; want 5", n.Value)
	}
}

func TestDeeplyNestedCallsAST(t *testing.T) {
	p := mustParse(t, "f(g(1, 2), h(3, i(4, 5)))")
	f := exprOf(firstStmt(p, 0)).(*ast.CallExpression)

	if got := len(f.ArgumentList); got != 2 {
		t.Fatalf("f args = %d; want 2", got)
	}

	g := f.ArgumentList[0].MustCall()
	if got := len(g.ArgumentList); got != 2 {
		t.Errorf("g args = %d; want 2", got)
	}

	h := f.ArgumentList[1].MustCall()
	if got := len(h.ArgumentList); got != 2 {
		t.Errorf("h args = %d; want 2", got)
	}

	i := h.ArgumentList[1].MustCall()
	if got := len(i.ArgumentList); got != 2 {
		t.Errorf("i args = %d; want 2", got)
	}
	if n := i.ArgumentList[0].MustNumLit(); n.Value != 4 {
		t.Errorf("i arg[0] = %v; want 4", n.Value)
	}
	if n := i.ArgumentList[1].MustNumLit(); n.Value != 5 {
		t.Errorf("i arg[1] = %v; want 5", n.Value)
	}
}

func TestDeeplyNestedStatementsAST(t *testing.T) {
	p := mustParse(t, `function f() {
		if (a) {
			while (b) {
				for (;;) {
					var x = 1;
					var y = 2;
				}
			}
		}
	}`)
	body := bodyOf(firstStmt(p, 0))
	if got := len(body.List); got != 1 {
		t.Fatalf("outer body = %d; want 1", got)
	}

	ifBlock := body.List[0].MustIf().Consequent.MustBlock()
	if got := len(ifBlock.List); got != 1 {
		t.Fatalf("if body = %d; want 1", got)
	}

	whileBlock := ifBlock.List[0].MustWhile().Body.MustBlock()
	if got := len(whileBlock.List); got != 1 {
		t.Fatalf("while body = %d; want 1", got)
	}

	forBlock := whileBlock.List[0].MustFor().Body.MustBlock()
	if got := len(forBlock.List); got != 2 {
		t.Fatalf("for body = %d; want 2", got)
	}
}

func TestMixedNestedSliceBuildersAST(t *testing.T) {
	// This test exercises all scratch buffer paths interleaved: arrays, calls,
	// templates, sequences, statement lists, and switch case consequents.
	p := mustParse(t, `function f() {
		var a = [1, g(2, 3), [4, 5]];
		switch (a[0]) {
			case 1:
				return h([...a], `+"`t ${a}`"+`);
			default:
				return (1, 2, 3);
		}
	}`)
	body := bodyOf(firstStmt(p, 0))

	if got := len(body.List); got != 2 {
		t.Fatalf("body statements = %d; want 2", got)
	}

	// First statement: var a = [1, g(2, 3), [4, 5]]
	arr := body.List[0].MustVarDecl().List[0].Initializer.MustArrLit()
	if got := len(arr.Value); got != 3 {
		t.Fatalf("array length = %d; want 3", got)
	}
	call := arr.Value[1].MustCall()
	if got := len(call.ArgumentList); got != 2 {
		t.Errorf("g() args = %d; want 2", got)
	}
	innerArr := arr.Value[2].MustArrLit()
	if got := len(innerArr.Value); got != 2 {
		t.Errorf("inner array length = %d; want 2", got)
	}

	// Second statement: switch
	sw := body.List[1].MustSwitch()
	if got := len(sw.Body); got != 2 {
		t.Fatalf("case count = %d; want 2", got)
	}
	// case 1: has 1 statement (return)
	if got := len(sw.Body[0].Consequent); got != 1 {
		t.Errorf("case 1 consequent = %d; want 1", got)
	}
	// default: has 1 statement (return)
	if got := len(sw.Body[1].Consequent); got != 1 {
		t.Errorf("default consequent = %d; want 1", got)
	}

	// Check the return in default has a sequence expression (1, 2, 3).
	retStmt := sw.Body[1].Consequent[0].MustReturn()
	seq := retStmt.Argument.MustSequence()
	if got := len(seq.Sequence); got != 3 {
		t.Errorf("sequence length = %d; want 3", got)
	}

	// Check the return in case 1 calls h with 2 args (array and template).
	retStmt1 := sw.Body[0].Consequent[0].MustReturn()
	hCall := retStmt1.Argument.MustCall()
	if got := len(hCall.ArgumentList); got != 2 {
		t.Errorf("h() args = %d; want 2", got)
	}
	// First arg is [...a]
	spreadArr := hCall.ArgumentList[0].MustArrLit()
	if got := len(spreadArr.Value); got != 1 {
		t.Errorf("[...a] length = %d; want 1", got)
	}
	if !spreadArr.Value[0].IsSpread() {
		t.Errorf("[...a][0] kind = %v; want Spread", spreadArr.Value[0].Kind())
	}
	// Second arg is template literal.
	tmpl := hCall.ArgumentList[1].MustTmplLit()
	if got := len(tmpl.Expressions); got != 1 {
		t.Errorf("template expressions = %d; want 1", got)
	}
}

// ---------------------------------------------------------------------------
// Syntax-only tests (broad coverage, no AST inspection)
// ---------------------------------------------------------------------------

func TestRegExpSyntax(t *testing.T) {
	cases := []string{
		"var r = /abc/",
		"var r = /abc/gi",
		"var r = /abc/gimsuy",
		`var r = /^hello$/`,
		`var r = /\d+/g`,
		`var r = /[a-zA-Z_$][a-zA-Z0-9_$]*/`,
		`var r = /(foo|bar|baz)/i`,
		`var r = /(?:https?:\/\/)?(?:www\.)?example\.com/`,
		`var r = /\b\w+\b/g`,
		`var r = /(?<=@)\w+/`,
		`var r = /(?<!\\)\$\{/g`,
		`if (/test/.test(str)) {}`,
		`var m = str.match(/(\d+)-(\d+)/)`,
		`var r = /\//`,
		`var r = /[/]/`,
		`var r = /a{1,3}/`,
		`x = y / z; var r = /abc/`,
	}
	for _, code := range cases {
		mustParse(t, code)
	}
}

func TestTemplateLiteralSyntax(t *testing.T) {
	cases := []string{
		"let x = `hello world`",
		"let x = ``",
		"let x = `hello ${name}`",
		"let x = `${a} + ${b} = ${a + b}`",
		"let x = `outer ${`inner`} outer`",
		"let x = `a ${`b ${`c`}`}`",
		"tag`hello`",
		"tag`hello ${name} world`",
		"`${fn(1, 2, 3)}`",
		"`${a ? b : c}`",
		"`${[1, 2, 3].join(',')}`",
		"`${(() => 42)()}`",
	}
	for _, code := range cases {
		mustParse(t, code)
	}
}

func TestControlFlowSyntax(t *testing.T) {
	cases := []string{
		"for (;;) {}",
		"for (var i = 0; i < 10; i++) {}",
		"for (let i = 0; i < 10; i++) {}",
		"for (let x of arr) {}",
		"for (let k in obj) {}",
		"for (let [a, b] of pairs) {}",
		"while (true) {}",
		"while (i < 10) { i++; }",
		"do { i++; } while (i < 10)",
		"try {} catch (e) {}",
		"try {} finally {}",
		"try {} catch (e) {} finally {}",
		"try {} catch {}",
		"if (a) {} else if (b) {} else if (c) {} else {}",
		"outer: for (;;) { inner: for (;;) { break outer; } }",
		"label: { break label; }",
	}
	for _, code := range cases {
		mustParse(t, code)
	}
}

func TestClassSyntax(t *testing.T) {
	cases := []string{
		"class A {}",
		"class A extends B {}",
		"class A { constructor() {} }",
		"class A { static x = 1; y = 2; }",
		"class A { #x; get x() { return this.#x; } }",
		"var A = class {}",
		"var A = class B extends C {}",
		"class A { static { this.x = 1; } }",
		`class A {
			method() {}
			async asyncMethod() {}
			*generatorMethod() {}
			async *asyncGenMethod() {}
			get prop() { return 1; }
			set prop(v) {}
			static staticMethod() {}
			[Symbol.iterator]() {}
		}`,
	}
	for _, code := range cases {
		mustParse(t, code)
	}
}

func TestGeneratorSyntax(t *testing.T) {
	cases := []string{
		"function* gen() { yield 1; yield 2; yield 3; }",
		"function* gen() { yield* other(); }",
		"var g = function*() { yield 1; }",
		"async function* gen() { yield await fetch(url); }",
	}
	for _, code := range cases {
		mustParse(t, code)
	}
}

func TestDestructuringSyntax(t *testing.T) {
	cases := []string{
		"var [a, b, c] = arr",
		"var [a, , b] = arr",
		"var [a, ...rest] = arr",
		"var [[a, b], [c, d]] = arr",
		"var [a = 1, b = 2] = arr",
		"var { a, b, c } = obj",
		"var { a: x, b: y } = obj",
		"var { a = 1, b = 2 } = obj",
		"var { a: { b: { c } } } = obj",
		"var { ...rest } = obj",
		"function f({ a, b }) { return a + b; }",
		"function f([a, b]) { return a + b; }",
		"var f = ({ x, y = 0 }) => x + y",
	}
	for _, code := range cases {
		mustParse(t, code)
	}
}

func TestComplexSnippetsSyntax(t *testing.T) {
	cases := []string{
		`const result = arr.filter(x => x > 0).map(x => x * 2).reduce((a, b) => a + b, 0)`,
		`async function fetchAll(urls) {
			const results = await Promise.all(
				urls.map(async (url) => {
					const res = await fetch(url);
					return res.json();
				})
			);
			return results.flat();
		}`,
		`class EventEmitter {
			#listeners = {};
			on(event, fn) { (this.#listeners[event] ??= []).push(fn); }
			emit(event, ...args) { for (const fn of this.#listeners[event] ?? []) { fn(...args); } }
		}`,
		`function parse(token) {
			switch (token.type) {
				case "string": return JSON.parse(token.value);
				case "number": return +token.value;
				case "null": return null;
				default: throw new Error("Unknown: " + token.type);
			}
		}`,
		"a?.b",
		"a?.b?.c",
		"a?.()",
		"a?.[0]",
		"a ?? b ?? c",
		"a ? b ? c : d : e ? f : g",
		"a &&= b",
		"a ||= b",
		"a ??= b",
	}
	for _, code := range cases {
		mustParse(t, code)
	}
}

// ===========================================================================
// AUTOMATIC SEMICOLON INSERTION (ASI) TESTS
// ===========================================================================

// mustFail verifies that code produces a parse error.
func mustFail(t *testing.T, code string) {
	t.Helper()
	_, err := parser.ParseFile(code)
	if err == nil {
		t.Errorf("expected parse error for:\n%s", code)
	}
}

func TestASIReturnNewline(t *testing.T) {
	p := mustParse(t, "function f() {\n  return\n  42\n}")
	body := bodyOf(firstStmt(p, 0))
	if got := len(body.List); got != 2 {
		t.Fatalf("body statements = %d; want 2 (return + expression)", got)
	}
	ret, ok := body.List[0].Return()
	if !ok {
		t.Fatalf("stmt[0] kind = %v; want Return", body.List[0].Kind())
	}
	if ret.Argument != nil {
		t.Errorf("return argument = %v; want nil (ASI should apply)", ret.Argument.Kind())
	}
	if !body.List[1].IsExpression() {
		t.Errorf("stmt[1] kind = %v; want Expression", body.List[1].Kind())
	}
}

func TestASIReturnSameLine(t *testing.T) {
	p := mustParse(t, "function f() { return 42 }")
	body := bodyOf(firstStmt(p, 0))
	if got := len(body.List); got != 1 {
		t.Fatalf("body statements = %d; want 1", got)
	}
	ret := body.List[0].MustReturn()
	if ret.Argument == nil || ret.Argument.IsNone() {
		t.Fatal("return argument is nil; want 42")
	}
	if n := ret.Argument.MustNumLit(); n.Value != 42 {
		t.Errorf("return value = %v; want 42", n.Value)
	}
}

func TestASIReturnSemicolon(t *testing.T) {
	p := mustParse(t, "function f() { return; 42 }")
	body := bodyOf(firstStmt(p, 0))
	ret := body.List[0].MustReturn()
	if ret.Argument != nil {
		t.Errorf("return argument = %v; want nil (explicit semicolon)", ret.Argument.Kind())
	}
}

func TestASIReturnRightBrace(t *testing.T) {
	p := mustParse(t, "function f() { return }")
	body := bodyOf(firstStmt(p, 0))
	ret := body.List[0].MustReturn()
	if ret.Argument != nil {
		t.Errorf("return argument should be nil before }")
	}
}

func TestASIReturnObject(t *testing.T) {
	p := mustParse(t, "function f() { return { a: 1 } }")
	body := bodyOf(firstStmt(p, 0))
	ret := body.List[0].MustReturn()
	if ret.Argument == nil || ret.Argument.IsNone() {
		t.Fatal("return argument is nil; want object literal")
	}
	if !ret.Argument.IsObjLit() {
		t.Errorf("return argument kind = %v; want ObjLit", ret.Argument.Kind())
	}
}

func TestASIThrowNewline(t *testing.T) {
	mustFail(t, "throw\nnew Error()")
}

func TestASIBreakNewline(t *testing.T) {
	p := mustParse(t, "outer: for (;;) { inner: for (;;) { break\nouter } }")
	outerFor := firstStmt(p, 0).(*ast.LabelledStatement).Statement.MustFor()
	outerBody := outerFor.Body.MustBlock()
	innerLabelled := outerBody.List[0].MustLabelled()
	innerFor := innerLabelled.Statement.MustFor()
	innerBody := innerFor.Body.MustBlock()

	if got := len(innerBody.List); got != 2 {
		t.Fatalf("inner body = %d; want 2 (break + expr)", got)
	}
	brk := innerBody.List[0].MustBreak()
	if brk.Label != nil {
		t.Errorf("break label = %q; want nil (ASI after break)", brk.Label.Name)
	}
}

func TestASIBreakSameLine(t *testing.T) {
	p := mustParse(t, "outer: for (;;) { break outer }")
	outerFor := firstStmt(p, 0).(*ast.LabelledStatement).Statement.MustFor()
	body := outerFor.Body.MustBlock()
	brk := body.List[0].MustBreak()
	if brk.Label == nil || brk.Label.Name != "outer" {
		t.Errorf("break label = %v; want outer", brk.Label)
	}
}

func TestASIContinueNewline(t *testing.T) {
	p := mustParse(t, "outer: for (;;) { for (;;) { continue\nouter } }")
	outerFor := firstStmt(p, 0).(*ast.LabelledStatement).Statement.MustFor()
	outerBody := outerFor.Body.MustBlock()
	innerFor := outerBody.List[0].MustFor()
	innerBody := innerFor.Body.MustBlock()

	if got := len(innerBody.List); got != 2 {
		t.Fatalf("inner body = %d; want 2 (continue + expr)", got)
	}
	cont := innerBody.List[0].MustContinue()
	if cont.Label != nil {
		t.Errorf("continue label = %q; want nil (ASI after continue)", cont.Label.Name)
	}
}

func TestASIContinueSameLine(t *testing.T) {
	p := mustParse(t, "outer: for (;;) { continue outer }")
	outerFor := firstStmt(p, 0).(*ast.LabelledStatement).Statement.MustFor()
	body := outerFor.Body.MustBlock()
	cont := body.List[0].MustContinue()
	if cont.Label == nil || cont.Label.Name != "outer" {
		t.Errorf("continue label = %v; want outer", cont.Label)
	}
}

func TestASIPostfixNewline(t *testing.T) {
	p := mustParse(t, "function f() { var i = 0;\ni\n++\ni }")
	body := bodyOf(firstStmt(p, 0))
	// var i = 0; i; ++i;
	if got := len(body.List); got != 3 {
		t.Fatalf("body statements = %d; want 3", got)
	}
	es := body.List[1].MustExpression()
	if id, ok := es.Expression.Ident(); !ok || id.Name != "i" {
		t.Errorf("stmt[1] = %v; want identifier 'i'", es.Expression.Kind())
	}
	es2 := body.List[2].MustExpression()
	upd, ok := es2.Expression.Update()
	if !ok {
		t.Fatalf("stmt[2] = %v; want Update", es2.Expression.Kind())
	}
	if upd.Postfix {
		t.Errorf("update should be prefix (++i), not postfix (i++)")
	}
}

// ===========================================================================
// NUMERIC LITERAL TESTS
// ===========================================================================

func TestNumericLiterals(t *testing.T) {
	tests := []struct {
		code string
		want float64
	}{
		{"var x = 0", 0},
		{"var x = 42", 42},
		{"var x = 3.14", 3.14},
		{"var x = .5", 0.5},
		{"var x = 1.", 1.0},
		{"var x = 3e9", 3e9},
		{"var x = 3E9", 3e9},
		{"var x = 5e-324", 5e-324},
		{"var x = 1e+10", 1e+10},
		{"var x = 1.5e2", 150},
		{"var x = .5e3", 500},
		{"var x = 0xff", 255},
		{"var x = 0xFF", 255},
		{"var x = 0o77", 63},
		{"var x = 0O77", 63},
		{"var x = 0b1010", 10},
		{"var x = 0B1010", 10},
		{"var x = 0e0", 0},
		{"var x = 0.0e0", 0},
		{"var x = 1_000_000", 1_000_000},
		{"var x = 0xff_ff", 0xffff},
	}
	for _, tt := range tests {
		p := mustParse(t, tt.code)
		num := initializerExpr(firstStmt(p, 0)).(*ast.NumberLiteral)
		if num.Value != tt.want {
			t.Errorf("%s: got %v, want %v", tt.code, num.Value, tt.want)
		}
	}
}

func TestNumericLiteralInExpressions(t *testing.T) {
	cases := []string{
		"x = 1 | 3E9",
		"x = 5e-324 >> 0",
		"x = 0 << 5e-324",
		"x = 1e3 + 2e3",
		"x = 1.5e2 * 2",
		"x = -1e10",
	}
	for _, code := range cases {
		mustParse(t, code)
	}
}

// ===========================================================================
// OPERATOR PRECEDENCE TESTS
// ===========================================================================

func TestPrecedenceMultiplicativeOverAdditive(t *testing.T) {
	p := mustParse(t, "var x = a + b * c")
	bin := initializerExpr(firstStmt(p, 0)).(*ast.BinaryExpression)
	if bin.Operator != token.Plus {
		t.Fatalf("top operator = %v; want +", bin.Operator)
	}
	right := bin.Right.MustBinary()
	if right.Operator != token.Multiply {
		t.Errorf("right operator = %v; want *", right.Operator)
	}
}

func TestPrecedenceComparisonOverLogical(t *testing.T) {
	p := mustParse(t, "var x = a < b && c > d")
	bin := initializerExpr(firstStmt(p, 0)).(*ast.BinaryExpression)
	if bin.Operator != token.LogicalAnd {
		t.Fatalf("top operator = %v; want &&", bin.Operator)
	}
	left := bin.Left.MustBinary()
	if left.Operator != token.Less {
		t.Errorf("left operator = %v; want <", left.Operator)
	}
	right := bin.Right.MustBinary()
	if right.Operator != token.Greater {
		t.Errorf("right operator = %v; want >", right.Operator)
	}
}

func TestPrecedenceOrOverAnd(t *testing.T) {
	p := mustParse(t, "var x = a || b && c")
	bin := initializerExpr(firstStmt(p, 0)).(*ast.BinaryExpression)
	if bin.Operator != token.LogicalOr {
		t.Fatalf("top operator = %v; want ||", bin.Operator)
	}
	right := bin.Right.MustBinary()
	if right.Operator != token.LogicalAnd {
		t.Errorf("right operator = %v; want &&", right.Operator)
	}
}

func TestPrecedenceTernaryOverAssignment(t *testing.T) {
	p := mustParse(t, "x = a ? b : c")
	assign := exprOf(firstStmt(p, 0)).(*ast.AssignExpression)
	cond, ok := assign.Right.Conditional()
	if !ok {
		t.Fatalf("rhs kind = %v; want Conditional", assign.Right.Kind())
	}
	if id := cond.Test.MustIdent(); id.Name != "a" {
		t.Errorf("test = %q; want a", id.Name)
	}
}

func TestPrecedenceUnaryOverBinary(t *testing.T) {
	p := mustParse(t, "var x = !a && b")
	bin := initializerExpr(firstStmt(p, 0)).(*ast.BinaryExpression)
	if bin.Operator != token.LogicalAnd {
		t.Fatalf("top operator = %v; want &&", bin.Operator)
	}
	unary, ok := bin.Left.Unary()
	if !ok {
		t.Fatalf("left kind = %v; want Unary", bin.Left.Kind())
	}
	if unary.Operator != token.Not {
		t.Errorf("unary operator = %v; want !", unary.Operator)
	}
}

func TestPrecedenceGrouping(t *testing.T) {
	p := mustParse(t, "var x = (a + b) * c")
	bin := initializerExpr(firstStmt(p, 0)).(*ast.BinaryExpression)
	if bin.Operator != token.Multiply {
		t.Fatalf("top operator = %v; want *", bin.Operator)
	}
	left := bin.Left.MustBinary()
	if left.Operator != token.Plus {
		t.Errorf("grouped operator = %v; want +", left.Operator)
	}
}

func TestPrecedenceBitwiseChain(t *testing.T) {
	// a | b ^ c & d  =>  a | (b ^ (c & d))
	p := mustParse(t, "var x = a | b ^ c & d")
	or := initializerExpr(firstStmt(p, 0)).(*ast.BinaryExpression)
	if or.Operator != token.Or {
		t.Fatalf("top = %v; want |", or.Operator)
	}
	xor := or.Right.MustBinary()
	if xor.Operator != token.ExclusiveOr {
		t.Fatalf("right = %v; want ^", xor.Operator)
	}
	and := xor.Right.MustBinary()
	if and.Operator != token.And {
		t.Errorf("inner = %v; want &", and.Operator)
	}
}

func TestPrecedenceNullishCoalescing(t *testing.T) {
	p := mustParse(t, "var x = a ?? b ?? c")
	outer := initializerExpr(firstStmt(p, 0)).(*ast.BinaryExpression)
	if outer.Operator != token.Coalesce {
		t.Fatalf("top = %v; want ??", outer.Operator)
	}
	inner := outer.Left.MustBinary()
	if inner.Operator != token.Coalesce {
		t.Errorf("left = %v; want ??", inner.Operator)
	}
}

func TestPrecedenceExponentiationRightAssociative(t *testing.T) {
	p := mustParse(t, "var x = a ** b ** c")
	outer := initializerExpr(firstStmt(p, 0)).(*ast.BinaryExpression)
	if outer.Operator != token.Exponent {
		t.Fatalf("top = %v; want **", outer.Operator)
	}
	inner, ok := outer.Right.Binary()
	if !ok {
		t.Fatalf("right = %v; want Binary", outer.Right.Kind())
	}
	if inner.Operator != token.Exponent {
		t.Errorf("right op = %v; want **", inner.Operator)
	}
}

// ===========================================================================
// ROUND-TRIP TESTS — broader coverage
// ===========================================================================

func TestRoundTripStatements(t *testing.T) {
	tests := []struct{ in, want string }{
		{"throw new Error('bad')", "throw new Error('bad');"},
		{"debugger", "debugger;"},
		{"while (true) break", "while (true) break;"},
		{"while (true) continue", "while (true) continue;"},
		{"do {} while (x)", "do {} while(x);"},
		{"with (obj) {}", "with (obj) {}"},
	}
	for _, tt := range tests {
		assertRoundTrip(t, tt.in, tt.want)
	}
}

func TestRoundTripExpressions(t *testing.T) {
	tests := []struct{ in, want string }{
		{"typeof x", "typeof x;"},
		{"void 0", "void 0;"},
		{"delete obj.x", "delete obj.x;"},
		{"new Foo(1, 2)", "new Foo(1, 2);"},
		{"a instanceof b", "a instanceof b;"},
		{"a in b", "a in b;"},
	}
	for _, tt := range tests {
		assertRoundTrip(t, tt.in, tt.want)
	}
}

func TestRoundTripDeclarations(t *testing.T) {
	tests := []struct{ in, want string }{
		{"var x", "var x;"},
		{"let x = 1", "let x = 1;"},
		{"const x = 1", "const x = 1;"},
		{"var a = 1, b = 2, c = 3", "var a = 1, b = 2, c = 3;"},
	}
	for _, tt := range tests {
		assertRoundTrip(t, tt.in, tt.want)
	}
}

// ===========================================================================
// ASYNC / AWAIT TESTS
// ===========================================================================

func TestAsyncAwaitSyntax(t *testing.T) {
	cases := []string{
		"async function f() { await fetch(url) }",
		"async function f() { const x = await promise }",

		"var f = async () => await 1",
		"var f = async (x) => await x",
		"class A { async method() { await this.load() } }",
		"var obj = { async method() { await 1 } }",
	}
	for _, code := range cases {
		mustParse(t, code)
	}
}

// ===========================================================================
// MEMBER EXPRESSION / CALL CHAIN TESTS
// ===========================================================================

func TestMemberCallChainAST(t *testing.T) {
	// a.b.c(1).d[2] — top-level is computed member [2], whose object is dot .d,
	// whose object is the call a.b.c(1).
	p := mustParse(t, "a.b.c(1).d[2]")
	top, ok := exprOf(firstStmt(p, 0)).(*ast.MemberExpression)
	if !ok {
		t.Fatalf("top = %T; want *MemberExpression", exprOf(firstStmt(p, 0)))
	}
	dot, ok := top.Object.Member()
	if !ok {
		t.Fatalf("top.object = %v; want Member (.d)", top.Object.Kind())
	}
	call, ok := dot.Object.Call()
	if !ok {
		t.Fatalf("dot.object = %v; want Call", dot.Object.Kind())
	}
	if got := len(call.ArgumentList); got != 1 {
		t.Errorf("call args = %d; want 1", got)
	}
}

func TestOptionalChainSyntax(t *testing.T) {
	cases := []string{
		"a?.b",
		"a?.b?.c",
		"a?.()",
		"a?.b()",
		"a?.[0]",
		"a?.b?.[0]?.()",
		"(a?.b).c",
	}
	for _, code := range cases {
		mustParse(t, code)
	}
}

// ===========================================================================
// DESTRUCTURING PATTERN TESTS
// ===========================================================================

func TestDestructuringAssignment(t *testing.T) {
	cases := []string{
		"[a, b] = [1, 2]",
		"[a, ...rest] = arr",
		"({ a, b } = obj)",
		"({ a: x, b: y } = obj)",
		"[a, [b, c]] = nested",
		"({ a: { b } } = deep)",
		"[a = 1, b = 2] = arr",
		"({ a = 1, b = 2 } = obj)",
	}
	for _, code := range cases {
		mustParse(t, code)
	}
}

func TestFunctionDefaultParameters(t *testing.T) {
	cases := []string{
		"function f(a = 1) {}",
		"function f(a = 1, b = 2) {}",
		"function f(a, b = a + 1) {}",
		"function f({ x = 0, y = 0 } = {}) {}",
		"function f([a, b] = [1, 2]) {}",
		"var f = (x = 1) => x",
	}
	for _, code := range cases {
		mustParse(t, code)
	}
}

func TestRestParameters(t *testing.T) {
	cases := []string{
		"function f(...args) {}",
		"function f(a, b, ...rest) {}",
		"var f = (...args) => args",
	}
	for _, code := range cases {
		mustParse(t, code)
	}
}

// ===========================================================================
// FOR-IN / FOR-OF TESTS
// ===========================================================================

func TestForInStatementAST(t *testing.T) {
	p := mustParse(t, "for (var k in obj) {}")
	forIn := firstStmt(p, 0).(*ast.ForInStatement)
	if forIn.Source == nil {
		t.Fatal("source is nil")
	}
	if id := forIn.Source.MustIdent(); id.Name != "obj" {
		t.Errorf("source = %q; want obj", id.Name)
	}
}

func TestForOfStatementAST(t *testing.T) {
	p := mustParse(t, "for (const x of arr) {}")
	forOf := firstStmt(p, 0).(*ast.ForOfStatement)
	if forOf.Source == nil {
		t.Fatal("source is nil")
	}
	if id := forOf.Source.MustIdent(); id.Name != "arr" {
		t.Errorf("source = %q; want arr", id.Name)
	}
}

// ===========================================================================
// ASSIGNMENT OPERATORS
// ===========================================================================

func TestAssignmentOperatorsAST(t *testing.T) {
	ops := []struct {
		code string
		tok  token.Token
	}{
		{"x = 1", token.Assign},
		{"x += 1", token.Plus},
		{"x -= 1", token.Minus},
		{"x *= 1", token.Multiply},
		{"x /= 1", token.Slash},
		{"x %= 1", token.Remainder},
		{"x **= 1", token.Exponent},
		{"x <<= 1", token.ShiftLeft},
		{"x >>= 1", token.ShiftRight},
		{"x >>>= 1", token.UnsignedShiftRight},
		{"x &= 1", token.And},
		{"x |= 1", token.Or},
		{"x ^= 1", token.ExclusiveOr},
		{"x &&= 1", token.LogicalAnd},
		{"x ||= 1", token.LogicalOr},
		{"x ??= 1", token.Coalesce},
	}
	for _, tt := range ops {
		p := mustParse(t, tt.code)
		assign := exprOf(firstStmt(p, 0)).(*ast.AssignExpression)
		if assign.Operator != tt.tok {
			t.Errorf("%s: operator = %v; want %v", tt.code, assign.Operator, tt.tok)
		}
	}
}

// ===========================================================================
// UNARY / UPDATE EXPRESSION TESTS
// ===========================================================================

func TestUnaryExpressionsAST(t *testing.T) {
	tests := []struct {
		code string
		op   token.Token
	}{
		{"!x", token.Not},
		{"~x", token.BitwiseNot},
		{"+x", token.Plus},
		{"-x", token.Minus},
		{"typeof x", token.Typeof},
		{"void x", token.Void},
		{"delete x", token.Delete},
	}
	for _, tt := range tests {
		p := mustParse(t, tt.code)
		unary := exprOf(firstStmt(p, 0)).(*ast.UnaryExpression)
		if unary.Operator != tt.op {
			t.Errorf("%s: op = %v; want %v", tt.code, unary.Operator, tt.op)
		}
	}
}

func TestUpdateExpressionsAST(t *testing.T) {
	tests := []struct {
		code    string
		op      token.Token
		postfix bool
	}{
		{"++x", token.Increment, false},
		{"--x", token.Decrement, false},
		{"x++", token.Increment, true},
		{"x--", token.Decrement, true},
	}
	for _, tt := range tests {
		p := mustParse(t, tt.code)
		upd := exprOf(firstStmt(p, 0)).(*ast.UpdateExpression)
		if upd.Operator != tt.op {
			t.Errorf("%s: op = %v; want %v", tt.code, upd.Operator, tt.op)
		}
		if upd.Postfix != tt.postfix {
			t.Errorf("%s: postfix = %v; want %v", tt.code, upd.Postfix, tt.postfix)
		}
	}
}

// ===========================================================================
// DO-WHILE / WHILE / LABELLED STATEMENT TESTS
// ===========================================================================

func TestDoWhileAST(t *testing.T) {
	p := mustParse(t, "do { x++ } while (x < 10)")
	dw := firstStmt(p, 0).(*ast.DoWhileStatement)
	if dw.Test == nil {
		t.Fatal("test is nil")
	}
	bin := dw.Test.MustBinary()
	if bin.Operator != token.Less {
		t.Errorf("test op = %v; want <", bin.Operator)
	}
}

func TestWhileAST(t *testing.T) {
	p := mustParse(t, "while (x > 0) { x-- }")
	w := firstStmt(p, 0).(*ast.WhileStatement)
	bin := w.Test.MustBinary()
	if bin.Operator != token.Greater {
		t.Errorf("test op = %v; want >", bin.Operator)
	}
}

func TestLabelledBreakContinue(t *testing.T) {
	p := mustParse(t, "loop: for (;;) { break loop; }")
	labelled := firstStmt(p, 0).(*ast.LabelledStatement)
	if labelled.Label.Name != "loop" {
		t.Errorf("label = %q; want loop", labelled.Label.Name)
	}
	forStmt := labelled.Statement.MustFor()
	body := forStmt.Body.MustBlock()
	brk := body.List[0].MustBreak()
	if brk.Label == nil || brk.Label.Name != "loop" {
		t.Errorf("break label = %v; want loop", brk.Label)
	}
}

// ===========================================================================
// PARSE ERROR TESTS
// ===========================================================================

func TestParseErrors(t *testing.T) {
	cases := []string{
		"var",
		"function",
		"if",
		"if ()",
		"for (;;",
		"switch",
		"class {",
		"(1 +)",
		"var x = {,}",
		"x.%",
		"let [",
		"let {",
		"=> x",
	}
	for _, code := range cases {
		_, err := parser.ParseFile(code)
		if err == nil {
			t.Errorf("expected parse error for: %s", code)
		}
	}
}

// ===========================================================================
// EDGE CASE SYNTAX TESTS
// ===========================================================================

func TestEdgeCaseSyntax(t *testing.T) {
	cases := []string{
		"",
		";",
		";;;;;;",
		"(((((1)))))",
		"a, b, c",
		"1 + 2 + 3 + 4 + 5",
		"a.b.c.d.e.f",
		"a()()()()",
		"new new Foo()",
		"new Foo.Bar()",
		"a[0][1][2]",
		"void typeof delete x",
		"+'1'",
		"-'1'",
		"~0",
		"!![].length",
		"0, 1, 2",
		"x = y = z = 1",
		"true ? 1 : false ? 2 : 3",
		`"use strict"`,
		"var x = /regex/ + 1",
	}
	for _, code := range cases {
		mustParse(t, code)
	}
}

func TestComputedPropertyKey(t *testing.T) {
	p := mustParse(t, "var o = { [1 + 2]: 'three' }")
	obj := initializerExpr(firstStmt(p, 0)).(*ast.ObjectLiteral)
	if got := len(obj.Value); got != 1 {
		t.Fatalf("property count = %d; want 1", got)
	}
	pk := obj.Value[0].MustKeyed()
	if !pk.Computed {
		t.Error("computed = false; want true")
	}
	bin := pk.Key.MustBinary()
	if bin.Operator != token.Plus {
		t.Errorf("key op = %v; want +", bin.Operator)
	}
}

func TestNewExpressionAST(t *testing.T) {
	p := mustParse(t, "new Foo(1, 2)")
	newExpr := exprOf(firstStmt(p, 0)).(*ast.NewExpression)
	if id := newExpr.Callee.MustIdent(); id.Name != "Foo" {
		t.Errorf("callee = %q; want Foo", id.Name)
	}
	if got := len(newExpr.ArgumentList); got != 2 {
		t.Fatalf("arg count = %d; want 2", got)
	}
}

func TestThrowStatementAST(t *testing.T) {
	p := mustParse(t, "throw new Error('msg')")
	thr := firstStmt(p, 0).(*ast.ThrowStatement)
	if thr.Argument == nil {
		t.Fatal("argument is nil")
	}
	if !thr.Argument.IsNew() {
		t.Errorf("argument = %v; want New", thr.Argument.Kind())
	}
}

func TestEmptyStatementAST(t *testing.T) {
	p := mustParse(t, ";;;")
	if got := len(p.Body); got != 3 {
		t.Fatalf("body = %d; want 3", got)
	}
	for i := range p.Body {
		if !p.Body[i].IsEmpty() {
			t.Errorf("body[%d] = %v; want Empty", i, p.Body[i].Kind())
		}
	}
}

func TestYieldExpressionAST(t *testing.T) {
	p := mustParse(t, "function* g() { yield 1; yield* other(); }")
	body := firstStmt(p, 0).(*ast.FunctionDeclaration).Function.Body
	if got := len(body.List); got != 2 {
		t.Fatalf("body = %d; want 2", got)
	}

	es1 := body.List[0].MustExpression()
	y1 := es1.Expression.MustYield()
	if y1.Delegate {
		t.Error("yield 1 should not be delegate")
	}
	if y1.Argument == nil {
		t.Error("yield 1 should have argument")
	}

	es2 := body.List[1].MustExpression()
	y2 := es2.Expression.MustYield()
	if !y2.Delegate {
		t.Error("yield* should be delegate")
	}
}
