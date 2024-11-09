package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/t14raptor/go-fast/generator"
	"github.com/t14raptor/go-fast/parser"
	"github.com/t14raptor/go-fast/simplifier"
)

func TestSimplifier(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("assets", "in.js"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := parser.ParseFile(string(data))
	if err != nil {
		t.Fatal(err)
	}
	simplifier.Simplify(program)

	out := generator.Generate(program)
	if err := os.WriteFile(filepath.Join("assets", "out.js"), []byte(out), 0644); err != nil {
		panic(err)
	}
}
