package generator

import (
	"strings"
	"testing"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/parser"
)

func parseSource(src string) (ast.VisitableNode, error) {
	ast, err := parser.ParseFile(src)
	if err != nil {
		return nil, err
	}
	return ast, nil
}

func generateASTNoIndent(program ast.VisitableNode) string {
	output := Generate(program)
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(output, "\n", ""), "    ", ""), "'", "\"")
}

func TestSequenceExpressionInNewExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "sequence as single argument to new",
			input:    "new F6(((a = 1), 2));",
			expected: "new F6((a = 1, 2));",
		},
		{
			name:     "sequence as second argument to new",
			input:    "new F6(x, ((b = 2), 3));",
			expected: "new F6(x, (b = 2, 3));",
		},
		{
			name:     "sequence as third argument to new",
			input:    "new F6(x, y, ((c = 3), 4));",
			expected: "new F6(x, y, (c = 3, 4));",
		},
		{
			name:     "sequence with function literal in new",
			input:    "new F6(h, ((r = R), function (W) { return r++; }));",
			expected: "new F6(h, (r = R, function(W) {return r++;}));",
		},
		{
			name:     "sequence in regular function call (should work)",
			input:    "f(((d = 4), 5));",
			expected: "f((d = 4, 5));",
		},
		{
			name:     "sequence as second argument in regular call (should work)",
			input:    "f(x, ((e = 5), 6));",
			expected: "f(x, (e = 5, 6));",
		},
		{
			name:     "sequence in throw statement",
			input:    "throw ((a = 1), 2);",
			expected: "throw (a = 1, 2);",
		},
		{
			name:     "sequence in await expression",
			input:    "async function f() { await ((b = 2), 3); }",
			expected: "async function f() {await (b = 2, 3);}",
		},
		{
			name:     "sequence in return statement",
			input:    "function g() { return ((d = 4), 5); }",
			expected: "function g() {return (d = 4, 5);}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := parseSource(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse input: %v", err)
			}

			result := generateASTNoIndent(ctx)

			if result != tt.expected {
				t.Errorf("\nInput:    %s\nExpected: %s\nGot:      %s", tt.input, tt.expected, result)
			}
		})
	}
}
