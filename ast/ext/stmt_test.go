package ext_test

import (
	"testing"

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
