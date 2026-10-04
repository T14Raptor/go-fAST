package ext_test

import (
	"fmt"
	"testing"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/ast/ext"
	"github.com/t14raptor/go-fast/parser"
	"github.com/t14raptor/go-fast/resolver"
)

func TestMayHaveSideEffectsStmtChecksLexicalInitializers(t *testing.T) {
	program, err := parser.Parse("let x = foo();")
	if err != nil {
		t.Fatal(err)
	}

	if !ext.MayHaveSideEffectsStmt(program.Body[0]) {
		t.Fatal("let initializer call should be side-effectful")
	}

	program, err = parser.Parse("const x = 1;")
	if err != nil {
		t.Fatal(err)
	}

	if ext.MayHaveSideEffectsStmt(program.Body[0]) {
		t.Fatal("pure const initializer should not be side-effectful")
	}

	program, err = parser.Parse("const {[foo()]: x} = {}; ")
	if err != nil {
		t.Fatal(err)
	}

	if !ext.MayHaveSideEffectsStmt(program.Body[0]) {
		t.Fatal("destructuring patterns should be treated as side-effectful")
	}
}

// IsGlobalRefTo reads the resolved context: a global anywhere in the program
// is unresolved, and a declaration of the name, even at the top level, is not
// the global.
func TestIsGlobalRefToAfterResolve(t *testing.T) {
	for _, tt := range []struct {
		src, name string
		want      bool
	}{
		{"function f() { return undefined; }", "undefined", true},
		{"function f(undefined) { return undefined; }", "undefined", false},
		{"function f() { var undefined; return undefined; }", "undefined", false},
		{"function f() { return Math; }", "Math", true},
		{"var Math = {}; function f() { return Math; }", "Math", false},
	} {
		program, err := parser.Parse(tt.src)
		if err != nil {
			t.Fatal(err)
		}
		resolver.Resolve(program)
		fn := program.Body[len(program.Body)-1].MustFuncDecl()
		body := fn.Function.Body.List
		ret := body[len(body)-1].MustReturn()
		if got := ext.IsGlobalRefTo(ret.Argument, tt.name); got != tt.want {
			t.Errorf("%s: IsGlobalRefTo(%s) = %v, want %v", tt.src, tt.name, got, tt.want)
		}
	}
}

// Reading a global that may not exist can throw a ReferenceError.
func TestReadingAnUndeclaredGlobalMayHaveSideEffects(t *testing.T) {
	for src, want := range map[string]bool{
		"maybeUndeclared;": true,
		"Math;":            false,
		"var x; x;":        false,
	} {
		program, err := parser.Parse(src)
		if err != nil {
			t.Fatal(err)
		}
		resolver.Resolve(program)
		expr := program.Body[len(program.Body)-1].MustExpression().Expression
		if got := ext.MayHaveSideEffects(expr); got != want {
			t.Errorf("%s: MayHaveSideEffects = %v, want %v", src, got, want)
		}
	}
}

// A declared undefined, NaN or Math is the program's own binding, not the
// global: casting it is unknown.
func TestCastsOnlyFoldGlobals(t *testing.T) {
	for _, tt := range []struct {
		src        string
		wantNumber bool
		wantString bool
	}{
		{"function f() { return undefined; }", true, true},
		{"function f(undefined) { return undefined; }", false, false},
		{"function f() { return NaN; }", true, true},
		{"function f(NaN) { return NaN; }", false, false},
		{"function f(Math) { return Math; }", false, false},
	} {
		program, err := parser.Parse(tt.src)
		if err != nil {
			t.Fatal(err)
		}
		resolver.Resolve(program)
		body := program.Body[0].MustFuncDecl().Function.Body.List
		arg := body[len(body)-1].MustReturn().Argument
		if n, _ := ext.CastToNumber(arg); n.Known() != tt.wantNumber {
			t.Errorf("%s: CastToNumber known = %v, want %v", tt.src, n.Known(), tt.wantNumber)
		}
		if s := ext.AsPureString(arg); s.Known() != tt.wantString {
			t.Errorf("%s: AsPureString known = %v, want %v", tt.src, s.Known(), tt.wantString)
		}
	}
}

// initializer parses `x = (src);` and returns the right-hand side, resolved.
func initializer(t *testing.T, src string, opts parser.Options) *ast.Expression {
	t.Helper()
	program, err := parser.ParseWithOptions("x = ("+src+");", opts)
	if err != nil {
		t.Fatal(err)
	}
	resolver.Resolve(program)
	return program.Body[0].MustExpression().Expression.MustAssign().Right
}

// The helpers analyze values, so parentheses kept by the parser must not
// change their answers.
func TestHelpersLookThroughParens(t *testing.T) {
	for _, src := range []string{
		"1", "(1)", "'a' + (b)", "(!0)", "void (0)", "[(1), (null)]", "(foo())",
		"(a), 1", "typeof (x)", "(Math).abs", "('abc').length", "(null)", "(undefined)",
		"(1) / (0)", "((NaN))", "(1) - (1)", "('a')", "(a) || (1)", "new (Date)()",
		"({ a: (1) }).a",
	} {
		plain := initializer(t, src, parser.Options{})
		kept := initializer(t, src, parser.Options{PreserveParens: true})
		if !kept.IsParen() {
			t.Fatalf("%s: expected a preserved paren, got %s", src, kept.Kind())
		}
		describe := func(e *ast.Expression) string {
			b, bp := ext.CastToBool(e)
			n, np := ext.CastToNumber(e)
			return fmt.Sprintf("bool=%v/%v num=%v/%v str=%v type=%v string=%v array=%v void=%v pure=%v effects=%v",
				b, bp, n, np, ext.AsPureString(e), ext.GetType(e), ext.IsString(e), ext.IsArrayLiteral(e),
				ext.IsVoid(e), ext.IsPureCallee(e), ext.MayHaveSideEffects(e))
		}
		if got, want := describe(kept), describe(plain); got != want {
			t.Errorf("%s:\n  with parens: %s\n  without:     %s", src, got, want)
		}

		var plainEffects, keptEffects []ast.Expression
		ext.ExtractSideEffectsTo(&plainEffects, plain)
		ext.ExtractSideEffectsTo(&keptEffects, kept)
		if len(keptEffects) != len(plainEffects) {
			t.Errorf("%s: %d side effects with parens, %d without", src, len(keptEffects), len(plainEffects))
		}
	}
}
