package parser_test

import (
	"testing"

	"github.com/t14raptor/go-fast/parser"
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
