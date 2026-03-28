package generator_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/t14raptor/go-fast/generator"
	"github.com/t14raptor/go-fast/parser"
)

func gen(in string) (string, error) {
	p, err := parser.ParseFile(in)
	if err != nil {
		return "", err
	}
	out := generator.Generate(p)
	out = regexp.MustCompile(`\s+`).ReplaceAllString(out, " ")
	out = strings.TrimSpace(out)
	return out, nil
}

func TestMetaProperty(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{`function Foo() { new.target; }`, `function Foo() { new.target; }`},
		{`function Foo() { if (new.target) {} }`, `function Foo() { if (new.target) {} }`},
		{`function Foo() { let x = new.target; }`, `function Foo() { let x = new.target; }`},
	}
	for _, tt := range tests {
		got, err := gen(tt.in)
		if err != nil {
			t.Errorf("gen(%q) failed: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("gen(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}
