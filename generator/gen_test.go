package generator

import (
	"fmt"
	"testing"

	"github.com/t14raptor/go-fast/parser"
)

func TestGen(t *testing.T) {
	p, err := parser.ParseFile(`0.0["toString"]()`)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(Generate(p))
}
