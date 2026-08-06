package parser_test

import (
	"testing"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/generator"
	"github.com/t14raptor/go-fast/parser"
	"github.com/t14raptor/go-fast/resolver"
)

type identCollector struct {
	ast.NoopVisitor
	name  string
	nodes []*ast.Identifier
}

func (c *identCollector) VisitIdentifier(n *ast.Identifier) {
	if n.Name == c.name {
		c.nodes = append(c.nodes, n)
	}
}

func collectIdents(prog *ast.Program, name string) []*ast.Identifier {
	c := &identCollector{name: name}
	c.V = c
	prog.VisitChildrenWith(c)
	return c.nodes
}

// parenCollector gathers every ParenthesizedExpression in a tree.
type parenCollector struct {
	ast.NoopVisitor
	nodes []*ast.ParenthesizedExpression
}

func (c *parenCollector) VisitParenthesizedExpression(n *ast.ParenthesizedExpression) {
	c.nodes = append(c.nodes, n)
	n.VisitChildrenWith(c.V)
}

func collectParens(prog *ast.Program) []*ast.ParenthesizedExpression {
	c := &parenCollector{}
	c.V = c
	prog.VisitChildrenWith(c)
	return c.nodes
}

func balanced(s string) bool {
	depth := 0
	for _, r := range s {
		switch r {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
		}
		if depth < 0 {
			return false
		}
	}
	return depth == 0
}

// By default grouping parentheses are discarded and the AST shape is unchanged:
// no ParenthesizedExpression node ever appears.
func TestPreserveParensDefaultOff(t *testing.T) {
	for _, src := range []string{`x = (a) || b;`, `y = a || (b);`, `z = ((a));`, `w = f((a + b));`} {
		prog, err := parser.Parse(src)
		if err != nil {
			t.Fatalf("parse %q: %v", src, err)
		}
		if got := collectParens(prog); len(got) != 0 {
			t.Errorf("%q: expected no ParenthesizedExpression by default, got %d", src, len(got))
		}
	}
}

// With PreserveParens, a parenthesized expression becomes a node whose
// Idx0/Idx1 span the parentheses, so the source slice is always balanced —
// unlike the inner node's own offsets, which stop inside the parentheses.
func TestPreserveParensBalancedSpan(t *testing.T) {
	cases := []struct {
		src  string
		want string // the exact parenthesized source the node should cover
	}{
		{`a || (b);`, `(b)`},
		{`(a) || b;`, `(a)`},
		{`a || (b || c);`, `(b || c)`},
		{`p["k"] || (a + b);`, `(a + b)`},
	}
	for _, tc := range cases {
		prog, err := parser.Parse(tc.src, parser.PreserveParens())
		if err != nil {
			t.Fatalf("parse %q: %v", tc.src, err)
		}
		parens := collectParens(prog)
		if len(parens) != 1 {
			t.Fatalf("%q: want 1 paren node, got %d", tc.src, len(parens))
		}
		n := parens[0]
		got := tc.src[n.Idx0():n.Idx1()]
		if got != tc.want {
			t.Errorf("%q: node span = %q, want %q", tc.src, got, tc.want)
		}
		if !balanced(got) {
			t.Errorf("%q: node span %q is not balanced", tc.src, got)
		}
	}
}

// Nested grouping nests the node: ((a)) -> Paren(Paren(Identifier)).
func TestPreserveParensNested(t *testing.T) {
	prog, err := parser.Parse(`x = ((a));`, parser.PreserveParens())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	parens := collectParens(prog)
	if len(parens) != 2 {
		t.Fatalf("want 2 nested paren nodes, got %d", len(parens))
	}
	inner, ok := parens[1].Expression.Identifier()
	if !ok {
		t.Fatalf("innermost paren should wrap an identifier, got kind %v", parens[1].Expression.Kind())
	}
	if inner.Name != "a" {
		t.Errorf("innermost identifier = %q, want %q", inner.Name, "a")
	}
}

// Only the single-expression form is wrapped. A comma sequence stays a
// SequenceExpression and empty parens stay an error node — no Paren wrapper.
func TestPreserveParensSequenceUnchanged(t *testing.T) {
	prog, err := parser.Parse(`x = (a, b);`, parser.PreserveParens())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := collectParens(prog); len(got) != 0 {
		t.Errorf("comma sequence should not be wrapped, got %d paren nodes", len(got))
	}
}

// The generator round-trips the node, re-emitting the parentheses it stands
// for. Redundant parentheses that the default path would drop are preserved.
func TestPreserveParensRoundTrip(t *testing.T) {
	src := `x = (a) + b;`

	def, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parse default: %v", err)
	}
	if got := generator.Generate(def); got != "x = a + b;\n" {
		t.Errorf("default generate = %q, want %q", got, "x = a + b;\n")
	}

	kept, err := parser.Parse(src, parser.PreserveParens())
	if err != nil {
		t.Fatalf("parse preserve: %v", err)
	}
	if got := generator.Generate(kept); got != "x = (a) + b;\n" {
		t.Errorf("preserve generate = %q, want %q", got, "x = (a) + b;\n")
	}
}

// The resolver descends through the node, so a reference wrapped in
// parentheses resolves to the same binding as an unwrapped one.
func TestPreserveParensResolvesThrough(t *testing.T) {
	prog, err := parser.Parse(`function f() { var a = 1; return a + (a); }`, parser.PreserveParens())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	resolver.Resolve(prog)

	idents := collectIdents(prog, "a")
	if len(idents) != 3 { // declarator + two references
		t.Fatalf("want 3 occurrences of 'a', got %d", len(idents))
	}
	ctx := idents[0].ScopeContext
	if ctx == ast.UnresolvedContext {
		t.Fatal("declaration of 'a' left unresolved")
	}
	for i, id := range idents {
		if id.ScopeContext != ctx {
			t.Errorf("occurrence %d resolved to a different binding than the declaration", i)
		}
	}
}
