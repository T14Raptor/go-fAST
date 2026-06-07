package generator

import (
	"testing"

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
