package ext_test

import (
	"testing"

	"github.com/t14raptor/go-fast/ast/ext"
	"github.com/t14raptor/go-fast/parser"
)

func TestMayHaveSideEffectsStmtChecksLexicalInitializers(t *testing.T) {
	program, err := parser.ParseFile("let x = foo();")
	if err != nil {
		t.Fatal(err)
	}

	if !ext.MayHaveSideEffectsStmt(program.Body[0]) {
		t.Fatal("let initializer call should be side-effectful")
	}

	program, err = parser.ParseFile("const x = 1;")
	if err != nil {
		t.Fatal(err)
	}

	if ext.MayHaveSideEffectsStmt(program.Body[0]) {
		t.Fatal("pure const initializer should not be side-effectful")
	}

	program, err = parser.ParseFile("const {[foo()]: x} = {}; ")
	if err != nil {
		t.Fatal(err)
	}

	if !ext.MayHaveSideEffectsStmt(program.Body[0]) {
		t.Fatal("destructuring patterns should be treated as side-effectful")
	}
}
