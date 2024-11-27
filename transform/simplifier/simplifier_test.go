package simplifier_test

import (
	"strings"
	"testing"

	"github.com/t14raptor/go-fast/generator"
	"github.com/t14raptor/go-fast/parser"
	"github.com/t14raptor/go-fast/transform/simplifier"
)

func simplify(in string) (string, error) {
	p, err := parser.ParseFile(in)
	if err != nil {
		return "", err
	}
	simplifier.Simplify(p, true)
	return generator.Generate(p), nil
}

func TestJSFuck(t *testing.T) {
	var tests = []struct {
		in   string
		want string
	}{
		{in: "![]", want: "false"},
		{in: "!![]", want: "true"},
		{in: "[][[]]", want: "undefined"},
		{in: "+[![]]", want: "NaN"},
		{in: "+[]", want: "0"},
		{in: "+!+[]", want: "1"},
		{in: "!+[]+!+[]", want: "2"},
		{in: "[+!+[]]+[+[]]", want: "\"10\""},
		{in: "(![]+[])[+!+[]]+(![]+[])[!+[]+!+[]]+(!![]+[])[!+[]+!+[]+!+[]]+(!![]+[])[+!+[]]+(!![]+[])[+[]]+([][(![]+[])[+[]]+(![]+[])[!+[]+!+[]]+(![]+[])[+!+[]]+(!![]+[])[+[]]]+[])[+!+[]+[!+[]+!+[]+!+[]]]+[+!+[]]+([+[]]+![]+[][(![]+[])[+[]]+(![]+[])[!+[]+!+[]]+(![]+[])[+!+[]]+(!![]+[])[+[]]])[!+[]+!+[]+[+[]]]", want: "\"alert(1)\""},
	}

	for _, test := range tests {
		got, err := simplify(test.in)
		got = strings.TrimSuffix(strings.TrimSpace(got), ";")
		if err != nil {
			t.Errorf("simplify('%s') failed: %v", test.in, err)
			continue
		}
		if got != test.want {
			t.Errorf("simplify('%s') = '%s'; want '%s'", test.in, got, test.want)
		}
	}
}
