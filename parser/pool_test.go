package parser

import (
	"reflect"
	"testing"
)

// putParser must not leave the last parse's nodes in the scratch buffers'
// backing arrays, or the pooled parser keeps that AST alive.
func TestPutParserClearsScratchBuffers(t *testing.T) {
	src := `var a = [1, [2, 3], {x: 4, y: [5]}], {p, q: [r]} = o;
class C { m() { return f(a, ...b, c); } static n = 1; }
function g(x, {y}, [z]) { if (x) { y(); z(); } }`
	p := getParser(src, Options{})
	if _, err := p.parse(); err != nil {
		t.Fatal(err)
	}
	putParser(p)
	assertCleared(t, "exprBuf", p.exprBuf)
	assertCleared(t, "stmtBuf", p.stmtBuf)
	assertCleared(t, "propBuf", p.propBuf)
	assertCleared(t, "elemBuf", p.elemBuf)
	assertCleared(t, "declBuf", p.declBuf)
	assertCleared(t, "patBuf", p.patBuf)
	assertCleared(t, "patPropBuf", p.patPropBuf)
}

func assertCleared[T any](t *testing.T, name string, b []T) {
	t.Helper()
	if len(b) != 0 {
		t.Errorf("%s: len %d after putParser, want 0", name, len(b))
	}
	for i, v := range b[:cap(b)] {
		if !reflect.ValueOf(&v).Elem().IsZero() {
			t.Errorf("%s[%d] still holds a node after putParser", name, i)
			return
		}
	}
}
