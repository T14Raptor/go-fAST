package simplifier_test

import (
	"fmt"
	"testing"
)

func TestCustom(t *testing.T) {
	in := `
    var hlF;
    hlF = this[typeof LL()[A9(1)] === [] + [][[]] ? LL()[A9(11)].apply(null, [709, 266]) : LL()[A9(3)].apply(null, [71, 650])] = wEF(pgF), Wq.pop(), hlF;`
	out, err := simplify(in)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(out)
}
