package deadcode_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/t14raptor/go-fast/generator"
	"github.com/t14raptor/go-fast/parser"
	"github.com/t14raptor/go-fast/transform/deadcode"
)

func dce(in string) (string, error) {
	p, err := parser.ParseFile(in)
	if err != nil {
		return "", err
	}
	deadcode.Eliminate(p, true)
	return generator.Generate(p), nil
}

func test(in, want string, t *testing.T) {
	got, err := dce(in)
	if err != nil {
		t.Errorf("dce('%s') failed: %v", in, err)
		return
	}
	got = regexp.MustCompile(`\s+`).ReplaceAllString(got, " ")
	got = strings.TrimSpace(got)
	if got != want {
		t.Errorf("dce('%s') = '%s'; want '%s'", in, got, want)
	}
}

func TestFunctionRemoval(t *testing.T) {
	test(`function a() { return "a"; } function b() { return "b"; }`, ``, t)
	test(`function a() { return "a"; } function b() { return "b"; } b();`, `function b() { return "b"; } b();`, t)
	test(`function a() { return "a"; } function b() { return a(); } b();`, `function a() { return "a"; } function b() { return a(); } b();`, t)
	test(`function a() { return "a"; } function b() { return "b"; } function c() { return c(); } a();`, `function a() { return "a"; } a();`, t)
}

func TestClassRemoval(t *testing.T) {
	test(`class A { constructor() {} } class B { constructor() {} }`, ``, t)
	test(`class A { constructor() {} } class B { constructor() {} } new B();`, `class B { constructor() {} } new B();`, t)
	test(`class A { constructor() {} } class B { constructor() { new A(); } } new B();`, `class A { constructor() {} } class B { constructor() { new A(); } } new B();`, t)
	test(`class A { constructor() {} } class B { constructor() {} } class C { constructor() { new C(); } } new A()`, `class A { constructor() {} } new A();`, t)

	test(`class A { b() { new B(); } } class B { a() { new A(); } } class C { a() { new A(); } } new B();`, `class A { b() { new B(); } } class B { a() { new A(); } } new B();`, t)
}

func TestVariableRemoval(t *testing.T) {
	test(`var a = 1; var b = 2;`, ``, t)
	test(`var a = 1; var b = 2; b;`, `var b = 2; b;`, t)
	test(`var a = 1; var b = a; b;`, `var a = 1; var b = a; b;`, t)

	test(`var a = 1, b = 2, c = 3; a;`, `var a = 1; a;`, t)
	test(`var a = 1, b = 2, c = 3; b;`, `var b = 2; b;`, t)

	test(`var a = 1; function b() { return a; } b();`, `var a = 1; function b() { return a; } b();`, t)
	test(`var a = 1; function b() { return a; } a;`, `var a = 1; a;`, t)

	test(`function a() { return 1; } var b = a();`, `function a() { return 1; } a();`, t)
	test(`class A { } var a = new A();`, `class A { } new A();`, t)
}

func TestAssignmentRemoval(t *testing.T) {
	test(`var a; a = 1;`, ``, t)
	test(`var a; a = 1; a;`, `var a; a = 1; a;`, t)

	test(`var a, b; a = 1; b = 2; a;`, `var a; a = 1; a;`, t)
	test(`var a, b; a = 1; b = 2; b;`, `var b; b = 2; b;`, t)
}

func TestEmptyStatement(t *testing.T) {
	test(`;`, ``, t)
	test(`1;`, `1;`, t)
	test(`1; ;`, `1;`, t)
	test(`1; 2;`, `1; 2;`, t)
	test(`1; ; 2;`, `1; 2;`, t)
}

func TestEmptyBlockStmt(t *testing.T) {
	test(`{ }`, ``, t)
	test(`{ {} a(); }`, `{ a(); }`, t)
}

func TestObject(t *testing.T) {
	test(`function a() {} var b = { a: a }; console.log(b.a);`, `function a() {} var b = { a: a }; console.log(b.a);`, t)
}
