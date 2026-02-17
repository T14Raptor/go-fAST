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

// firstStmt returns the Stmt inside the i-th top-level statement.
func firstStmt(p *ast.Program, i int) ast.Stmt {
	return p.Body[i].Stmt
}

// exprOf extracts the inner Expr from an ExpressionStatement.
func exprOf(s ast.Stmt) ast.Expr {
	return s.(*ast.ExpressionStatement).Expression.Expr
}

// initializerExpr extracts the initializer expression from the first
// VariableDeclarator of a VariableDeclaration statement.
func initializerExpr(s ast.Stmt) ast.Expr {
	return s.(*ast.VariableDeclaration).List[0].Initializer.Expr
}

// bodyOf extracts the BlockStatement body from a FunctionDeclaration.
func bodyOf(s ast.Stmt) *ast.BlockStatement {
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

	if _, ok := arr.Value[0].Expr.(*ast.NumberLiteral); !ok {
		t.Errorf("arr[0] type = %T; want *NumberLiteral", arr.Value[0].Expr)
	}
	if n := arr.Value[0].Expr.(*ast.NumberLiteral); n.Value != 1 {
		t.Errorf("arr[0] value = %v; want 1", n.Value)
	}
	if _, ok := arr.Value[1].Expr.(*ast.StringLiteral); !ok {
		t.Errorf("arr[1] type = %T; want *StringLiteral", arr.Value[1].Expr)
	}
	if s := arr.Value[1].Expr.(*ast.StringLiteral); s.Value != "two" {
		t.Errorf("arr[1] value = %q; want \"two\"", s.Value)
	}
	if _, ok := arr.Value[2].Expr.(*ast.BooleanLiteral); !ok {
		t.Errorf("arr[2] type = %T; want *BooleanLiteral", arr.Value[2].Expr)
	}
	if _, ok := arr.Value[3].Expr.(*ast.NullLiteral); !ok {
		t.Errorf("arr[3] type = %T; want *NullLiteral", arr.Value[3].Expr)
	}
}

func TestArrayLiteralElisionsAST(t *testing.T) {
	p := mustParse(t, "var a = [1,,2,,3]")
	arr := initializerExpr(firstStmt(p, 0)).(*ast.ArrayLiteral)

	if got := len(arr.Value); got != 5 {
		t.Fatalf("array length = %d; want 5", got)
	}
	// Positions 1 and 3 should be elisions (Expression with nil Expr).
	if arr.Value[1].Expr != nil {
		t.Errorf("arr[1] = %T; want nil (elision)", arr.Value[1].Expr)
	}
	if arr.Value[3].Expr != nil {
		t.Errorf("arr[3] = %T; want nil (elision)", arr.Value[3].Expr)
	}
	// Positions 0, 2, 4 should be NumberLiterals.
	for _, i := range []int{0, 2, 4} {
		if _, ok := arr.Value[i].Expr.(*ast.NumberLiteral); !ok {
			t.Errorf("arr[%d] type = %T; want *NumberLiteral", i, arr.Value[i].Expr)
		}
	}
}

func TestArrayLiteralSpreadAST(t *testing.T) {
	p := mustParse(t, "var a = [1, ...b, 2]")
	arr := initializerExpr(firstStmt(p, 0)).(*ast.ArrayLiteral)

	if got := len(arr.Value); got != 3 {
		t.Fatalf("array length = %d; want 3", got)
	}
	if _, ok := arr.Value[0].Expr.(*ast.NumberLiteral); !ok {
		t.Errorf("arr[0] type = %T; want *NumberLiteral", arr.Value[0].Expr)
	}
	spread, ok := arr.Value[1].Expr.(*ast.SpreadElement)
	if !ok {
		t.Fatalf("arr[1] type = %T; want *SpreadElement", arr.Value[1].Expr)
	}
	if id, ok := spread.Expression.Expr.(*ast.Identifier); !ok || id.Name != "b" {
		t.Errorf("spread target = %v; want identifier 'b'", spread.Expression.Expr)
	}
	if _, ok := arr.Value[2].Expr.(*ast.NumberLiteral); !ok {
		t.Errorf("arr[2] type = %T; want *NumberLiteral", arr.Value[2].Expr)
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
		num, ok := arr.Value[idx].Expr.(*ast.NumberLiteral)
		if !ok {
			t.Errorf("arr[%d] type = %T; want *NumberLiteral", idx, arr.Value[idx].Expr)
			continue
		}
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
	if _, ok := call.ArgumentList[0].Expr.(*ast.NumberLiteral); !ok {
		t.Errorf("arg[0] type = %T; want *NumberLiteral", call.ArgumentList[0].Expr)
	}
	if _, ok := call.ArgumentList[1].Expr.(*ast.StringLiteral); !ok {
		t.Errorf("arg[1] type = %T; want *StringLiteral", call.ArgumentList[1].Expr)
	}
	if _, ok := call.ArgumentList[2].Expr.(*ast.BooleanLiteral); !ok {
		t.Errorf("arg[2] type = %T; want *BooleanLiteral", call.ArgumentList[2].Expr)
	}
}

func TestArgumentListSpreadAST(t *testing.T) {
	p := mustParse(t, "f(1, ...args)")
	call := exprOf(firstStmt(p, 0)).(*ast.CallExpression)

	if got := len(call.ArgumentList); got != 2 {
		t.Fatalf("arg count = %d; want 2", got)
	}
	if _, ok := call.ArgumentList[0].Expr.(*ast.NumberLiteral); !ok {
		t.Errorf("arg[0] type = %T; want *NumberLiteral", call.ArgumentList[0].Expr)
	}
	if _, ok := call.ArgumentList[1].Expr.(*ast.SpreadElement); !ok {
		t.Errorf("arg[1] type = %T; want *SpreadElement", call.ArgumentList[1].Expr)
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

	inner1 := outer.ArgumentList[0].Expr.(*ast.CallExpression)
	if got := len(inner1.ArgumentList); got != 2 {
		t.Errorf("inner1 arg count = %d; want 2", got)
	}

	inner2 := outer.ArgumentList[1].Expr.(*ast.CallExpression)
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
		num, ok := seq.Sequence[i].Expr.(*ast.NumberLiteral)
		if !ok {
			t.Errorf("seq[%d] type = %T; want *NumberLiteral", i, seq.Sequence[i].Expr)
			continue
		}
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
	n1 := seq.Sequence[0].Expr.(*ast.NumberLiteral)
	n2 := seq.Sequence[1].Expr.(*ast.NumberLiteral)
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
	id1, ok := tmpl.Expressions[0].Expr.(*ast.Identifier)
	if !ok {
		t.Fatalf("expr[0] type = %T; want *Identifier", tmpl.Expressions[0].Expr)
	}
	if id1.Name != "name" {
		t.Errorf("expr[0] name = %q; want \"name\"", id1.Name)
	}

	id2, ok := tmpl.Expressions[1].Expr.(*ast.Identifier)
	if !ok {
		t.Fatalf("expr[1] type = %T; want *Identifier", tmpl.Expressions[1].Expr)
	}
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
		id, ok := tmpl.Expressions[i].Expr.(*ast.Identifier)
		if !ok {
			t.Errorf("expr[%d] type = %T; want *Identifier", i, tmpl.Expressions[i].Expr)
			continue
		}
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
	inner, ok := outer.Expressions[0].Expr.(*ast.TemplateLiteral)
	if !ok {
		t.Fatalf("outer.expr[0] type = %T; want *TemplateLiteral", outer.Expressions[0].Expr)
	}
	if got := len(inner.Expressions); got != 1 {
		t.Fatalf("inner expression count = %d; want 1", got)
	}
	id, ok := inner.Expressions[0].Expr.(*ast.Identifier)
	if !ok {
		t.Fatalf("inner.expr[0] type = %T; want *Identifier", inner.Expressions[0].Expr)
	}
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
	tagId, ok := tmpl.Tag.Expr.(*ast.Identifier)
	if !ok {
		t.Fatalf("tag type = %T; want *Identifier", tmpl.Tag.Expr)
	}
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
	re, ok := call.ArgumentList[0].Expr.(*ast.RegExpLiteral)
	if !ok {
		t.Fatalf("arg[0] type = %T; want *RegExpLiteral", call.ArgumentList[0].Expr)
	}
	if re.Pattern != "foo" || re.Flags != "g" {
		t.Errorf("regex = /%s/%s; want /foo/g", re.Pattern, re.Flags)
	}
	str, ok := call.ArgumentList[1].Expr.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("arg[1] type = %T; want *StringLiteral", call.ArgumentList[1].Expr)
	}
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
		if _, ok := body.List[i].Stmt.(*ast.VariableDeclaration); !ok {
			t.Errorf("stmt[%d] type = %T; want *VariableDeclaration", i, body.List[i].Stmt)
		}
	}
	if _, ok := body.List[3].Stmt.(*ast.ReturnStatement); !ok {
		t.Errorf("stmt[3] type = %T; want *ReturnStatement", body.List[3].Stmt)
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
	ifStmt, ok := body.List[0].Stmt.(*ast.IfStatement)
	if !ok {
		t.Fatalf("stmt[0] type = %T; want *IfStatement", body.List[0].Stmt)
	}
	block, ok := ifStmt.Consequent.Stmt.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("if body type = %T; want *BlockStatement", ifStmt.Consequent.Stmt)
	}
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
	decl0 := body.List[0].Stmt.(*ast.VariableDeclaration)
	if decl0.List[0].Target.Target.(*ast.Identifier).Name != "x0" {
		t.Errorf("first decl name = %q; want \"x0\"", decl0.List[0].Target.Target.(*ast.Identifier).Name)
	}
	declN := body.List[n-1].Stmt.(*ast.VariableDeclaration)
	want := fmt.Sprintf("x%d", n-1)
	if declN.List[0].Target.Target.(*ast.Identifier).Name != want {
		t.Errorf("last decl name = %q; want %q", declN.List[0].Target.Target.(*ast.Identifier).Name, want)
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
	if _, ok := p.Body[0].Stmt.(*ast.VariableDeclaration); !ok {
		t.Errorf("body[0] type = %T; want *VariableDeclaration", p.Body[0].Stmt)
	}
	if _, ok := p.Body[1].Stmt.(*ast.FunctionDeclaration); !ok {
		t.Errorf("body[1] type = %T; want *FunctionDeclaration", p.Body[1].Stmt)
	}
	if _, ok := p.Body[2].Stmt.(*ast.ClassDeclaration); !ok {
		t.Errorf("body[2] type = %T; want *ClassDeclaration", p.Body[2].Stmt)
	}
	if _, ok := p.Body[3].Stmt.(*ast.VariableDeclaration); !ok {
		t.Errorf("body[3] type = %T; want *VariableDeclaration", p.Body[3].Stmt)
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
	es, ok := sw.Body[0].Consequent[0].Stmt.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("case1.stmt[0] type = %T; want *ExpressionStatement", sw.Body[0].Consequent[0].Stmt)
	}
	call, ok := es.Expression.Expr.(*ast.CallExpression)
	if !ok {
		t.Fatalf("case1.stmt[0].expr type = %T; want *CallExpression", es.Expression.Expr)
	}
	if id := call.Callee.Expr.(*ast.Identifier); id.Name != "a" {
		t.Errorf("case1.stmt[0] callee = %q; want \"a\"", id.Name)
	}
	// Last statement should be break.
	if _, ok := sw.Body[0].Consequent[3].Stmt.(*ast.BreakStatement); !ok {
		t.Errorf("case1.stmt[3] type = %T; want *BreakStatement", sw.Body[0].Consequent[3].Stmt)
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
	s1 := sw.Body[0].Test.Expr.(*ast.StringLiteral)
	if s1.Value != "a" {
		t.Errorf("case 0 test = %q; want \"a\"", s1.Value)
	}
	s2 := sw.Body[1].Test.Expr.(*ast.StringLiteral)
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
	p1 := obj.Value[0].Prop.(*ast.PropertyKeyed)
	if id := p1.Key.Expr.(*ast.StringLiteral); id.Value != "a" {
		t.Errorf("prop[0] key = %q; want \"a\"", id.Value)
	}
	if n := p1.Value.Expr.(*ast.NumberLiteral); n.Value != 1 {
		t.Errorf("prop[0] value = %v; want 1", n.Value)
	}

	// Second property: b: 'two'
	p2 := obj.Value[1].Prop.(*ast.PropertyKeyed)
	if s := p2.Value.Expr.(*ast.StringLiteral); s.Value != "two" {
		t.Errorf("prop[1] value = %q; want \"two\"", s.Value)
	}

	// Third property: c: true
	p3 := obj.Value[2].Prop.(*ast.PropertyKeyed)
	if b := p3.Value.Expr.(*ast.BooleanLiteral); !b.Value {
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
		ps, ok := obj.Value[i].Prop.(*ast.PropertyShort)
		if !ok {
			t.Errorf("prop[%d] type = %T; want *PropertyShort", i, obj.Value[i].Prop)
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
	if _, ok := obj.Value[0].Prop.(*ast.PropertyKeyed); !ok {
		t.Errorf("prop[0] type = %T; want *PropertyKeyed", obj.Value[0].Prop)
	}
	spread, ok := obj.Value[1].Prop.(*ast.SpreadElement)
	if !ok {
		t.Fatalf("prop[1] type = %T; want *SpreadElement", obj.Value[1].Prop)
	}
	if id := spread.Expression.Expr.(*ast.Identifier); id.Name != "b" {
		t.Errorf("spread target = %q; want \"b\"", id.Name)
	}
	if _, ok := obj.Value[2].Prop.(*ast.PropertyKeyed); !ok {
		t.Errorf("prop[2] type = %T; want *PropertyKeyed", obj.Value[2].Prop)
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
	bin, ok := forStmt.Test.Expr.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("test type = %T; want *BinaryExpression", forStmt.Test.Expr)
	}
	if bin.Operator != token.Less {
		t.Errorf("test op = %v; want <", bin.Operator)
	}

	// Update should be an update expression: i++
	upd, ok := forStmt.Update.Expr.(*ast.UpdateExpression)
	if !ok {
		t.Fatalf("update type = %T; want *UpdateExpression", forStmt.Update.Expr)
	}
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
	if forStmt.Test != nil && forStmt.Test.Expr != nil {
		t.Errorf("test should be nil or have nil Expr")
	}
	if forStmt.Update != nil && forStmt.Update.Expr != nil {
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
		id := decl.List[i].Target.Target.(*ast.Identifier)
		if id.Name != names[i] {
			t.Errorf("decl[%d] name = %q; want %q", i, id.Name, names[i])
		}
		num := decl.List[i].Initializer.Expr.(*ast.NumberLiteral)
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
	if id := ifStmt.Test.Expr.(*ast.Identifier); id.Name != "a" {
		t.Errorf("if test = %q; want \"a\"", id.Name)
	}
	if ifStmt.Alternate == nil {
		t.Fatal("alternate is nil; want else-if chain")
	}

	// Alternate should be another IfStatement (else if).
	elseIf, ok := ifStmt.Alternate.Stmt.(*ast.IfStatement)
	if !ok {
		t.Fatalf("alternate type = %T; want *IfStatement", ifStmt.Alternate.Stmt)
	}
	if id := elseIf.Test.Expr.(*ast.Identifier); id.Name != "b" {
		t.Errorf("else-if test = %q; want \"b\"", id.Name)
	}
	if elseIf.Alternate == nil {
		t.Fatal("else-if alternate is nil; want else block")
	}

	// Final else should be a BlockStatement.
	if _, ok := elseIf.Alternate.Stmt.(*ast.BlockStatement); !ok {
		t.Errorf("final else type = %T; want *BlockStatement", elseIf.Alternate.Stmt)
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
		id := arrow.ParameterList.List[i].Target.Target.(*ast.Identifier)
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

	if id := cond.Test.Expr.(*ast.Identifier); id.Name != "a" {
		t.Errorf("test = %q; want \"a\"", id.Name)
	}
	if n := cond.Consequent.Expr.(*ast.NumberLiteral); n.Value != 1 {
		t.Errorf("consequent = %v; want 1", n.Value)
	}
	if n := cond.Alternate.Expr.(*ast.NumberLiteral); n.Value != 2 {
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
	catchParam := tr.Catch.Parameter.Target.(*ast.Identifier)
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

	inner0 := outer.Value[0].Expr.(*ast.ArrayLiteral)
	if got := len(inner0.Value); got != 2 {
		t.Errorf("inner0 length = %d; want 2", got)
	}

	inner1 := outer.Value[1].Expr.(*ast.ArrayLiteral)
	if got := len(inner1.Value); got != 2 {
		t.Errorf("inner1 length = %d; want 2", got)
	}

	// inner1[1] should be [4, [5]]
	nested := inner1.Value[1].Expr.(*ast.ArrayLiteral)
	if got := len(nested.Value); got != 2 {
		t.Errorf("nested length = %d; want 2", got)
	}
	deepest := nested.Value[1].Expr.(*ast.ArrayLiteral)
	if got := len(deepest.Value); got != 1 {
		t.Errorf("deepest length = %d; want 1", got)
	}
	if n := deepest.Value[0].Expr.(*ast.NumberLiteral); n.Value != 5 {
		t.Errorf("deepest value = %v; want 5", n.Value)
	}
}

func TestDeeplyNestedCallsAST(t *testing.T) {
	p := mustParse(t, "f(g(1, 2), h(3, i(4, 5)))")
	f := exprOf(firstStmt(p, 0)).(*ast.CallExpression)

	if got := len(f.ArgumentList); got != 2 {
		t.Fatalf("f args = %d; want 2", got)
	}

	g := f.ArgumentList[0].Expr.(*ast.CallExpression)
	if got := len(g.ArgumentList); got != 2 {
		t.Errorf("g args = %d; want 2", got)
	}

	h := f.ArgumentList[1].Expr.(*ast.CallExpression)
	if got := len(h.ArgumentList); got != 2 {
		t.Errorf("h args = %d; want 2", got)
	}

	i := h.ArgumentList[1].Expr.(*ast.CallExpression)
	if got := len(i.ArgumentList); got != 2 {
		t.Errorf("i args = %d; want 2", got)
	}
	if n := i.ArgumentList[0].Expr.(*ast.NumberLiteral); n.Value != 4 {
		t.Errorf("i arg[0] = %v; want 4", n.Value)
	}
	if n := i.ArgumentList[1].Expr.(*ast.NumberLiteral); n.Value != 5 {
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

	ifBlock := body.List[0].Stmt.(*ast.IfStatement).Consequent.Stmt.(*ast.BlockStatement)
	if got := len(ifBlock.List); got != 1 {
		t.Fatalf("if body = %d; want 1", got)
	}

	whileBlock := ifBlock.List[0].Stmt.(*ast.WhileStatement).Body.Stmt.(*ast.BlockStatement)
	if got := len(whileBlock.List); got != 1 {
		t.Fatalf("while body = %d; want 1", got)
	}

	forBlock := whileBlock.List[0].Stmt.(*ast.ForStatement).Body.Stmt.(*ast.BlockStatement)
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
	arr := body.List[0].Stmt.(*ast.VariableDeclaration).List[0].Initializer.Expr.(*ast.ArrayLiteral)
	if got := len(arr.Value); got != 3 {
		t.Fatalf("array length = %d; want 3", got)
	}
	call := arr.Value[1].Expr.(*ast.CallExpression)
	if got := len(call.ArgumentList); got != 2 {
		t.Errorf("g() args = %d; want 2", got)
	}
	innerArr := arr.Value[2].Expr.(*ast.ArrayLiteral)
	if got := len(innerArr.Value); got != 2 {
		t.Errorf("inner array length = %d; want 2", got)
	}

	// Second statement: switch
	sw := body.List[1].Stmt.(*ast.SwitchStatement)
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
	retStmt := sw.Body[1].Consequent[0].Stmt.(*ast.ReturnStatement)
	seq := retStmt.Argument.Expr.(*ast.SequenceExpression)
	if got := len(seq.Sequence); got != 3 {
		t.Errorf("sequence length = %d; want 3", got)
	}

	// Check the return in case 1 calls h with 2 args (array and template).
	retStmt1 := sw.Body[0].Consequent[0].Stmt.(*ast.ReturnStatement)
	hCall := retStmt1.Argument.Expr.(*ast.CallExpression)
	if got := len(hCall.ArgumentList); got != 2 {
		t.Errorf("h() args = %d; want 2", got)
	}
	// First arg is [...a]
	spreadArr := hCall.ArgumentList[0].Expr.(*ast.ArrayLiteral)
	if got := len(spreadArr.Value); got != 1 {
		t.Errorf("[...a] length = %d; want 1", got)
	}
	if _, ok := spreadArr.Value[0].Expr.(*ast.SpreadElement); !ok {
		t.Errorf("[...a][0] type = %T; want *SpreadElement", spreadArr.Value[0].Expr)
	}
	// Second arg is template literal.
	tmpl := hCall.ArgumentList[1].Expr.(*ast.TemplateLiteral)
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
