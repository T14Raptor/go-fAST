package generator

import (
	"testing"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/parser"
	"github.com/t14raptor/go-fast/resolver"
)

func assertMinified(t *testing.T, input, want string) {
	t.Helper()

	p, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse input: %v", err)
	}

	got := GenerateMinified(p)
	if got != want {
		t.Fatalf("gen(%q) = %q; want %q", input, got, want)
	}
}

func TestMinifiedOperatorTokenBoundaries(t *testing.T) {
	assertMinified(t, `a + ++b;`, `a+ ++b;`)
	assertMinified(t, `a - --b;`, `a- --b;`)
	assertMinified(t, `a + +b;`, `a+ +b;`)
	assertMinified(t, `a - -b;`, `a- -b;`)
	assertMinified(t, `x = a / /b/.source;`, `x=a/ /b/.source;`)
	assertMinified(t, `x = a / /b/();`, `x=a/ /b/();`)
}

func TestMetaProperty(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{`function Foo(){new.target;}`, `function Foo(){new.target;}`},
		{`function Foo(){if(new.target){}}`, `function Foo(){if(new.target){}}`},
		{`function Foo(){let x=new.target;}`, `function Foo(){let x=new.target;}`},
	}
	for _, tt := range tests {
		p, err := parser.Parse(tt.in)
		if err != nil {
			t.Fatalf("Failed to parse input: %v", err)
		}

		got := GenerateMinified(p)
		if got != tt.want {
			t.Errorf("gen(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}

func TestMethodKindGetSetKeywords(t *testing.T) {
	assertMinified(t,
		`({get value(){return 1;},set value(next){this.next=next;}});`,
		`({get value(){return 1;},set value(next){this.next=next;}});`,
	)
}

func TestForInitializerForbidInRegressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "assignment rhs",
			input: "for (x = (a in b);;) {}",
			want:  "for(x=(a in b);;){}",
		},
		{
			name:  "sequence element",
			input: "for (x, (a in b);;) {}",
			want:  "for(x,(a in b);;){}",
		},
		{
			name:  "conditional test",
			input: "for (((a in b) ? c : d);;) {}",
			want:  "for((a in b)?c:d;;){}",
		},
		{
			name:  "conditional alternate",
			input: "for ((a ? b : (c in d));;) {}",
			want:  "for(a?b:(c in d);;){}",
		},
		{
			name:  "binary left subtree",
			input: "for (((a in b) && c);;) {}",
			want:  "for((a in b)&&c;;){}",
		},
		{
			name:  "binary right subtree",
			input: "for (a && (b in c);;) {}",
			want:  "for(a&&(b in c);;){}",
		},
		{
			name:  "wrapped conditional test clears forbid-in",
			input: "for (((a in b) ? c : d) * e;;) {}",
			want:  "for((a in b?c:d)*e;;){}",
		},
		{
			name:  "wrapped conditional alternate clears forbid-in",
			input: "for ((a ? b : (c in d)) * e;;) {}",
			want:  "for((a?b:c in d)*e;;){}",
		},
		{
			name:  "wrapped assignment clears forbid-in",
			input: "for (1 * (x = (a in b));;) {}",
			want:  "for(1*(x=a in b);;){}",
		},
		{
			name:  "nested wrapped sequence clears forbid-in",
			input: "for ((x, (a in b)) * c;;) {}",
			want:  "for((x,a in b)*c;;){}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMinified(t, tt.input, tt.want)
		})
	}
}

func TestBinaryExprNestedRightRegressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "binary right subtree",
			input: "c >> (d & e);",
			want:  "c>>(d&e);",
		},
		{
			name:  "conditional consequent binary right subtree",
			input: "a && b ? c >> (d & e) : f;",
			want:  "a&&b?c>>(d&e):f;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMinified(t, tt.input, tt.want)
		})
	}
}

func TestSequenceExpressionInNewExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "sequence as single argument to new",
			input:    "new F6(((a=1),2));",
			expected: "new F6((a=1,2));",
		},
		{
			name:     "sequence as second argument to new",
			input:    "new F6(x,((b=2),3));",
			expected: "new F6(x,(b=2,3));",
		},
		{
			name:     "sequence as third argument to new",
			input:    "new F6(x,y,((c=3),4));",
			expected: "new F6(x,y,(c=3,4));",
		},
		{
			name:     "sequence with function literal in new",
			input:    "new F6(h,((r=R),function(W){return r++;}));",
			expected: "new F6(h,(r=R,function(W){return r++;}));",
		},
		{
			name:     "sequence in regular function call (should work)",
			input:    "f(((d=4),5));",
			expected: "f((d=4,5));",
		},
		{
			name:     "sequence as second argument in regular call (should work)",
			input:    "f(x,((e=5),6));",
			expected: "f(x,(e=5,6));",
		},
		{
			name:     "sequence in throw statement",
			input:    "throw ((a=1),2);",
			expected: "throw (a=1,2);",
		},
		{
			name:     "sequence in await expression",
			input:    "async function f(){await ((b=2),3);}",
			expected: "async function f(){await (b=2,3);}",
		},
		{
			name:     "sequence in return statement",
			input:    "function g(){return ((d=4),5);}",
			expected: "function g(){return (d=4,5);}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse input: %v", err)
			}

			result := GenerateMinified(ctx)
			if result != tt.expected {
				t.Errorf("\nInput:    %s\nExpected: %s\nGot:      %s", tt.input, tt.expected, result)
			}
		})
	}
}

func TestPatternRoundTrip(t *testing.T) {
	cases := []struct{ in, want string }{
		{`var [a, , b = 1, ...c] = x;`, `var [a,,b=1,...c]=x;`},
		{`var {a, b: {c} = {}, ...r} = o;`, `var {a,b:{c}={},...r}=o;`},
		{`for (const [k, v] of m) {}`, `for(const [k,v] of m){}`},
		{`for ([x.y, z] of p) {}`, `for([x.y,z] of p){}`},
		{`try {} catch ({message}) {}`, `try{}catch({message}){}`},
		{`([x.y, z] = p);`, `([x.y,z]=p);`},
		{`function f([a] = [], {b} = {}, ...rest) {}`, `function f([a]=[],{b}={},...rest){}`},
		{`({a = 1} = o);`, `({a=1}=o);`},
		{`label: x = 1;`, `label:x=1;`},
		{`obj.x = 1;`, `obj.x=1;`},
	}
	for _, c := range cases {
		p, err := parser.Parse(c.in)
		if err != nil {
			t.Fatalf("parse(%q): %v", c.in, err)
		}
		got := GenerateMinified(p)
		if got != c.want {
			t.Errorf("gen(%q) = %q; want %q", c.in, got, c.want)
		}
	}
}

func TestPatternEdgeRoundTrip(t *testing.T) {
	ok := []struct{ in, want string }{
		{`var [a, ...[b, c]] = x;`, `var [a,...[b,c]]=x;`},     // array rest is a nested pattern
		{`function f(...[a, b]) {}`, `function f(...[a,b]){}`}, // param rest is a pattern
		{`var {a, ...rest} = o;`, `var {a,...rest}=o;`},        // object rest ident
		{`[a, , b] = c;`, `([a,,b]=c);`},                       // assignment with elision hole
		{`({a, b} = c);`, `({a,b}=c);`},                        // parenthesised object assign
		{`for ({a} of x) {}`, `for({a} of x){}`},               // for-of object pattern target
		{`for ([a] in x) {}`, `for([a] in x){}`},               // for-in array pattern target
		{`function f({a = 1, b: {c} = {}}) {}`, `function f({a=1,b:{c}={}}){}`},
		{`var {[k]: v = 1} = o;`, `var {[k]:v=1}=o;`}, // computed key + default
		{`let [a, b = a] = x;`, `let [a,b=a]=x;`},     // sibling default reference
		{`var [a = b.c] = o;`, `var [a=b.c]=o;`},      // default value may be a member
	}
	for _, c := range ok {
		p, err := parser.Parse(c.in)
		if err != nil {
			t.Errorf("parse(%q): %v", c.in, err)
			continue
		}
		resolver.Resolve(p) // must not panic
		if got := GenerateMinified(p); got != c.want {
			t.Errorf("gen(%q) = %q; want %q", c.in, got, c.want)
		}
	}

	bad := []string{
		`var {...{a}} = o;`, // object rest must be a simple target
		`var [a.b] = c;`,    // member in binding position
		`var {a: b.c} = o;`, // member value in binding position
	}
	for _, src := range bad {
		if _, err := parser.Parse(src); err == nil {
			t.Errorf("parse(%q): expected error, got nil", src)
		}
	}
}

func TestOptionalChainingMinified(t *testing.T) {
	assertMinified(t, `a?.b; a?.(b); a?.[b];`, `a?.b;a?.(b);a?.[b];`)
	assertMinified(t, `(function(){})?.();`, `(function(){})?.();`)
	assertMinified(t, `(function(){})?.x;`, `(function(){})?.x;`)
}

func TestLiteralMemberAndCallBasesMinified(t *testing.T) {
	assertMinified(t, `(function(){}).x;`, `(function(){}).x;`)
	assertMinified(t, `(class {}).x;`, `(class {}).x;`)
	assertMinified(t, `(class {})();`, `(class {})();`)
	assertMinified(t, `({[(a,b)]:1});`, `({[(a,b)]:1});`)
}

func TestComputedMemberSequenceMinified(t *testing.T) {
	assertMinified(t, `a[(b,c)]; a?.[(b,c)];`, `a[(b,c)];a?.[(b,c)];`)
}

func TestClassExtendsAndSuperMinified(t *testing.T) {
	assertMinified(t,
		`class A extends B { constructor(){ super(); super.x; } }`,
		`class A extends B{constructor(){super();super.x;}}`,
	)
	assertMinified(t, `class A extends (a, b) {}`, `class A extends (a,b){}`)
}

func TestGeneratorFunctionsAndObjectMethodsMinified(t *testing.T) {
	assertMinified(t, `function* g(){ yield 1; }`, `function* g(){yield 1;}`)
	assertMinified(t, `const g = function*(){ yield 1; };`, `const g=function*(){yield 1;};`)
	assertMinified(t, `async function* g(){ yield 1; }`, `async function* g(){yield 1;}`)
	assertMinified(t,
		`({ m(){ return super.x; }, *g(){ yield 1; } });`,
		`({m(){return super.x;},*g(){yield 1;}});`,
	)
}

func TestTemplateLiteralMinified(t *testing.T) {
	assertMinified(t, "tag`x${y}`;", "tag`x${y}`;")
	assertMinified(t, "`\\${x}`;", "`\\${x}`;")
	assertMinified(t, "`\\n`;", "`\\n`;")
	assertMinified(t, "(function(){})`x`;", "(function(){})`x`;")
	assertMinified(t, "(class {})`x`;", "(class {})`x`;")
	assertMinified(t, "({})`x`;", "({})`x`;")
}

func TestArrayHoles(t *testing.T) {
	assertMinified(t, `x = [,];`, `x=[,];`)
	assertMinified(t, `x = [,,];`, `x=[,,];`)
	assertMinified(t, `x = [1, 2, ,];`, `x=[1,2,,];`)
	assertMinified(t, `x = [1, , 2];`, `x=[1,,2];`)
	assertMinified(t, `x = [1, 2,];`, `x=[1,2];`)
	assertMinified(t, `[a, ,] = it;`, `([a,,]=it);`)
	assertMinified(t, `var [, ] = it;`, `var [,]=it;`)
	assertMinified(t, `var [a, , ...r] = it;`, `var [a,,...r]=it;`)

	// The printed array has the parsed one's length.
	for _, in := range []string{`x = [,];`, `x = [1, 2, ,];`, `x = [, , 3, ,];`} {
		p, err := parser.Parse(in)
		if err != nil {
			t.Fatal(err)
		}
		want := len(p.Body[0].MustExpression().Expression.MustAssign().Right.MustArrayLit().Value)
		out, err := parser.Parse(Generate(p))
		if err != nil {
			t.Fatalf("reparse %q: %v", Generate(p), err)
		}
		if got := len(out.Body[0].MustExpression().Expression.MustAssign().Right.MustArrayLit().Value); got != want {
			t.Errorf("%s printed as %q: %d elements, want %d", in, Generate(p), got, want)
		}
	}
}

// A string literal with no source text is quoted for JavaScript, and parses
// back to the same value.
func TestStringLiteralWithoutRaw(t *testing.T) {
	for _, tt := range []struct{ value, want string }{
		{"plain", `"plain"`},
		{`a"b\c`, `"a\"b\\c"`},
		{"\a\x00\x1f\x7f", `"\x07\x00\x1f\x7f"`},
		{"\n\r\t\b\f\v", `"\n\r\t\b\f\v"`},
		{"\u2028\u2029\ufeff", `"\u2028\u2029\ufeff"`},
		{"caf\u00e9 \U0001F600", "\"caf\u00e9 \U0001F600\""},
		{"\U000E0001", `"\udb40\udc01"`},
	} {
		lit := ast.NewStringLitExpr(&ast.StringLiteral{Value: tt.value})
		p, err := parser.Parse("x = 1;")
		if err != nil {
			t.Fatal(err)
		}
		p.Body[0].MustExpression().Expression.MustAssign().Right = &lit
		got := GenerateMinified(p)
		if want := "x=" + tt.want + ";"; got != want {
			t.Errorf("value %q: got %s, want %s", tt.value, got, want)
			continue
		}
		back, err := parser.Parse(got)
		if err != nil {
			t.Errorf("reparse %s: %v", got, err)
			continue
		}
		if v := back.Body[0].MustExpression().Expression.MustAssign().Right.MustStringLit().Value; v != tt.value {
			t.Errorf("%s parses back to %q, want %q", got, v, tt.value)
		}
	}
}

func TestPreservedParens(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		// Kept as written, even where precedence makes them redundant.
		{`(a + b) * c;`, `(a+b)*c;`},
		{`(a * b) + c;`, `(a*b)+c;`},
		{`((a));`, `((a));`},
		{`x = (1);`, `x=(1);`},
		// The parentheses already delimit their contents.
		{`(a, b);`, `(a,b);`},
		{`f((a, b));`, `f((a,b));`},
		{`({}).toString();`, `({}).toString();`},
		{`(function () {})();`, `(function(){})();`},
		{`() => ({});`, `()=>({});`},
		{`(5).toString();`, `(5).toString();`},
		{`(a?.b).c;`, `(a?.b).c;`},
		{`new (foo())();`, `new (foo())();`},
		{`(-x) ** 2;`, `(-x)**2;`},
		{`(a ?? b) || c;`, `(a??b)||c;`},
		{`for (x = (a in b);;) {}`, `for(x=(a in b);;){}`},
		{`(a)++;`, `(a)++;`},
		{`typeof (a);`, `typeof (a);`},
		// Assignment targets keep only the bare target.
		{`(a) = 1;`, `a=1;`},
		{`[(a), (b.c)] = d;`, `([a,b.c]=d);`},
		{`for ((a) of b) {}`, `for(a of b){}`},
	}
	for _, tt := range tests {
		p, err := parser.ParseWithOptions(tt.in, parser.Options{PreserveParens: true})
		if err != nil {
			t.Fatalf("Failed to parse %q: %v", tt.in, err)
		}
		if got := GenerateMinified(p); got != tt.want {
			t.Errorf("gen(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}

// A ParenthesizedExpression built by hand prints like a parsed one: its
// contents are delimited, so a sequence or a bare `in` needs no more parens.
func TestPreservedParensBuiltByHand(t *testing.T) {
	a := ast.NewIdentifierExpr(&ast.Identifier{Name: "a"})
	b := ast.NewIdentifierExpr(&ast.Identifier{Name: "b"})
	seq := ast.NewSequenceExpr(&ast.SequenceExpression{Sequence: ast.Expressions{a, b}})
	paren := ast.NewParenExpr(&ast.ParenthesizedExpression{Expression: &seq})
	call := ast.NewCallExpr(&ast.CallExpression{
		Callee:       &a,
		ArgumentList: ast.Expressions{paren},
	})
	stmt := ast.NewExpressionStmt(&ast.ExpressionStatement{Expression: &call})
	if got, want := GenerateMinified(&ast.Program{Body: ast.Statements{stmt}}), `a((a,b));`; got != want {
		t.Errorf("gen = %q; want %q", got, want)
	}
}
