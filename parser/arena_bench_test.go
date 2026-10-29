package parser

import (
	"strings"
	"testing"
)

const benchFunction = "function f(a,b,c){var d=a[b]^c;for(var i=0;i<d.length;i++){c=(c<<5)-c+d.charCodeAt(i)|0;if(c>0x7fff){c^=b[i%3]}else{c=c+'x'}}return [c,d,a.x.y(1,2,3)]}\n"

// BenchmarkParseBySize parses sources from a short expression up to a large
// script, so arena sizing can be judged across input sizes.
func BenchmarkParseBySize(b *testing.B) {
	for _, tc := range []struct{ name, src string }{
		{"expr", "(a[b]^c+0x1f*(d>>>3))"},
		{"function", benchFunction},
		{"8KB", strings.Repeat(benchFunction, 8<<10/len(benchFunction))},
		{"256KB", strings.Repeat(benchFunction, 256<<10/len(benchFunction))},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tc.src)))
			for b.Loop() {
				if _, err := Parse(tc.src); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
