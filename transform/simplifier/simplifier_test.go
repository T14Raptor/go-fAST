package simplifier_test

import (
	"regexp"
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

func fold(in, want string, t *testing.T) {
	got, err := simplify(in)
	if err != nil {
		t.Errorf("simplify('%s') failed: %v", in, err)
		return
	}
	got = strings.TrimSuffix(strings.TrimSpace(got), ";")
	got = regexp.MustCompile(`\s+`).ReplaceAllString(got, " ")
	if got != want {
		t.Errorf("simplify('%s') = '%s'; want '%s'", in, got, want)
	}
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
		fold(test.in, test.want, t)
	}
}

func TestUndefinedComparison(t *testing.T) {
	var tests = []struct {
		in   string
		want string
	}{
		{in: "undefined == undefined", want: "true"},
		{in: "undefined == null", want: "true"},
		{in: "undefined == void 0", want: "true"},

		{in: "undefined == 0", want: "false"},
		{in: "undefined == 1", want: "false"},
		{in: "undefined == 'hi'", want: "false"},
		{in: "undefined == true", want: "false"},
		{in: "undefined == false", want: "false"},

		{in: "undefined === undefined", want: "true"},
		{in: "undefined === null", want: "false"},
		{in: "undefined === void 0", want: "true"},

		{in: "undefined == this", want: "undefined == this"},
		{in: "undefined == x", want: "undefined == x"},

		{in: "undefined != undefined", want: "false"},
		{in: "undefined != null", want: "false"},
		{in: "undefined != void 0", want: "false"},

		{in: "undefined != 0", want: "true"},
		{in: "undefined != 1", want: "true"},
		{in: "undefined != 'hi'", want: "true"},
		{in: "undefined != true", want: "true"},
		{in: "undefined != false", want: "true"},

		{in: "undefined !== undefined", want: "false"},
		{in: "undefined !== null", want: "true"},
		{in: "undefined !== void 0", want: "false"},

		{in: "undefined != this", want: "undefined != this"},
		{in: "undefined != x", want: "undefined != x"},

		{in: "undefined < undefined", want: "false"},
		{in: "undefined > undefined", want: "false"},
		{in: "undefined <= undefined", want: "false"},
		{in: "undefined >= undefined", want: "false"},

		{in: "0 < undefined", want: "false"},
		{in: "true > undefined", want: "false"},
		{in: "'hi' >= undefined", want: "false"},
		{in: "null <= undefined", want: "false"},

		{in: "undefined < 0", want: "false"},
		{in: "undefined > true", want: "false"},
		{in: "undefined >= 'hi'", want: "false"},
		{in: "undefined <= null", want: "false"},

		{in: "null == undefined", want: "true"},
		{in: "0 == undefined", want: "false"},
		{in: "1 == undefined", want: "false"},
		{in: "'hi' == undefined", want: "false"},
		{in: "true == undefined", want: "false"},
		{in: "false == undefined", want: "false"},
		{in: "null === undefined", want: "false"},
		{in: "void 0 === undefined", want: "true"},

		{in: "undefined == NaN", want: "false"},
		{in: "NaN == undefined", want: "false"},
		{in: "undefined == Infinity", want: "false"},
		{in: "Infinity == undefined", want: "false"},
		{in: "undefined == -Infinity", want: "false"},
		{in: "-Infinity == undefined", want: "false"},
		{in: "({}) == undefined", want: "false"},
		{in: "undefined == ({})", want: "false"},
		{in: "([]) == undefined", want: "false"},
		{in: "undefined == ([])", want: "false"},
		{in: "(/a/g) == undefined", want: "false"},
		{in: "undefined == (/a/g)", want: "false"},
		{in: "(function(){}) == undefined", want: "false"},
		{in: "undefined == (function(){})", want: "false"},

		{in: "undefined != NaN", want: "true"},
		{in: "NaN != undefined", want: "true"},
		{in: "undefined != Infinity", want: "true"},
		{in: "Infinity != undefined", want: "true"},
		{in: "undefined != -Infinity", want: "true"},
		{in: "-Infinity != undefined", want: "true"},
		{in: "({}) != undefined", want: "true"},
		{in: "undefined != ({})", want: "true"},
		{in: "([]) != undefined", want: "true"},
		{in: "undefined != ([])", want: "true"},
		{in: "(/a/g) != undefined", want: "true"},
		{in: "undefined != (/a/g)", want: "true"},
		{in: "(function(){}) != undefined", want: "true"},
		{in: "undefined != (function(){})", want: "true"},

		{in: "this == undefined", want: "this == undefined"},
		{in: "x == undefined", want: "x == undefined"},

		{in: "'123' !== void 0", want: "true"},
		{in: "'123' === void 0", want: "false"},
		{in: "void 0 !== '123'", want: "true"},
		{in: "void 0 === '123'", want: "false"},

		{in: "'123' !== undefined", want: "true"},
		{in: "'123' === undefined", want: "false"},
		{in: "undefined !== '123'", want: "true"},
		{in: "undefined === '123'", want: "false"},

		{in: "1 !== void 0", want: "true"},
		{in: "1 === void 0", want: "false"},
		{in: "null !== void 0", want: "true"},
		{in: "null === void 0", want: "false"},
		{in: "undefined !== void 0", want: "false"},
		{in: "undefined === void 0", want: "true"},
	}

	for _, test := range tests {
		fold(test.in, test.want, t)
	}
}

func TestNullComparison(t *testing.T) {
	var tests = []struct {
		in   string
		want string
	}{
		{in: "null == undefined", want: "true"},
		{in: "null == null", want: "true"},
		{in: "null == void 0", want: "true"},

		{in: "null == 0", want: "false"},
		{in: "null == 1", want: "false"},
		{in: "null == 'hi'", want: "false"},
		{in: "null == true", want: "false"},
		{in: "null == false", want: "false"},

		{in: "null === undefined", want: "false"},
		{in: "null === null", want: "true"},
		{in: "null === void 0", want: "false"},
		{in: "null === x", want: "null === x"},

		{in: "null == this", want: "null == this"},
		{in: "null == x", want: "null == x"},

		{in: "null != undefined", want: "false"},
		{in: "null != null", want: "false"},
		{in: "null != void 0", want: "false"},

		{in: "null != 0", want: "true"},
		{in: "null != 1", want: "true"},
		{in: "null != 'hi'", want: "true"},
		{in: "null != true", want: "true"},
		{in: "null != false", want: "true"},

		{in: "null !== undefined", want: "true"},
		{in: "null !== void 0", want: "true"},
		{in: "null !== null", want: "false"},

		{in: "null != this", want: "null != this"},
		{in: "null != x", want: "null != x"},

		{in: "null < null", want: "false"},
		{in: "null > null", want: "false"},
		{in: "null >= null", want: "true"},
		{in: "null <= null", want: "true"},

		{in: "0 < null", want: "false"},
		{in: "0 > null", want: "false"},
		{in: "0 >= null", want: "true"},
		{in: "true > null", want: "true"},
		{in: "'hi' < null", want: "false"},
		{in: "'hi' >= null", want: "false"},

		{in: "null < 0", want: "false"},
		{in: "null > true", want: "false"},
		{in: "null < 'hi'", want: "false"},
		{in: "null >= 'hi'", want: "false"},

		{in: "null == NaN", want: "false"},
		{in: "NaN == null", want: "false"},
		{in: "null == Infinity", want: "false"},
		{in: "Infinity == null", want: "false"},
		{in: "null == -Infinity", want: "false"},
		{in: "-Infinity == null", want: "false"},
		{in: "({}) == null", want: "false"},
		{in: "null == ({})", want: "false"},
		{in: "([]) == null", want: "false"},
		{in: "null == ([])", want: "false"},
		{in: "(/a/g) == null", want: "false"},
		{in: "null == (/a/g)", want: "false"},
		{in: "(function(){}) == null", want: "false"},
		{in: "null == (function(){})", want: "false"},

		{in: "null != NaN", want: "true"},
		{in: "NaN != null", want: "true"},
		{in: "null != Infinity", want: "true"},
		{in: "Infinity != null", want: "true"},
		{in: "null != -Infinity", want: "true"},
		{in: "-Infinity != null", want: "true"},
		{in: "({}) != null", want: "true"},
		{in: "null != ({})", want: "true"},
		{in: "([]) != null", want: "true"},
		{in: "null != ([])", want: "true"},
		{in: "(/a/g) != null", want: "true"},
		{in: "null != (/a/g)", want: "true"},
		{in: "(function(){}) != null", want: "true"},
		{in: "null != (function(){})", want: "true"},

		{in: "this == null", want: "this == null"},
		{in: "x == null", want: "x == null"},
	}

	for _, test := range tests {
		fold(test.in, test.want, t)
	}
}

func TestBooleanBooleanComparison(t *testing.T) {
	fold("!x == !y", "!x == !y", t)
	fold("!x < !y", "!x < !y", t)
	fold("!x !== !y", "!x !== !y", t)

	fold("!x == !x", "!x == !x", t)   // foldable
	fold("!x < !x", "!x < !x", t)     // foldable
	fold("!x !== !x", "!x !== !x", t) // foldable
}

func TestBooleanNumberComparison(t *testing.T) {
	fold("!x == +y", "!x == +y", t)
	fold("!x <= +y", "!x <= +y", t)
	fold("!x !== +y", "true", t)
}

func TestNumberBooleanComparison(t *testing.T) {
	fold("+x == !y", "+x == !y", t)
	fold("+x <= !y", "+x <= !y", t)
	fold("+x === !y", "false", t)
}

func TestBooleanStringComparison(t *testing.T) {
	fold("!x == '' + y", "!x == '' + y", t)
	fold("!x <= '' + y", "!x <= '' + y", t)
	fold("!x !== '' + y", "true", t)
}

func TestStringBooleanComparison(t *testing.T) {
	fold("'' + x == !y", "'' + x == !y", t)
	fold("'' + x <= !y", "'' + x <= !y", t)
	fold("'' + x === !y", "false", t)
}

func TestNumberNumberComparison(t *testing.T) {
	fold("1 > 1", "false", t)
	fold("2 == 3", "false", t)
	fold("3.6 === 3.6", "true", t)
	fold("+x > +y", "+x > +y", t)
	fold("+x == +y", "+x == +y", t)
	fold("+x === +y", "+x === +y", t)
	fold("+x == +x", "+x == +x", t)
	fold("+x === +x", "+x === +x", t)

	fold("+x > +x", "+x > +x", t) // foldable
}

func TestStringStringComparison(t *testing.T) {
	fold("'a' < 'b'", "true", t)
	fold("'a' <= 'b'", "true", t)
	fold("'a' > 'b'", "false", t)
	fold("'a' >= 'b'", "false", t)
	fold("+'a' < +'b'", "false", t)
	fold("typeof a < 'a'", "typeof a < 'a'", t)
	fold("'a' >= typeof a", "'a' >= typeof a", t)
	fold("typeof a < typeof a", "false", t)
	fold("typeof a >= typeof a", "true", t)
	fold("typeof 3 > typeof 4", "false", t)
	fold("typeof function() {} < typeof function() {}", "false", t)
	fold("'a' == 'a'", "true", t)
	fold("'b' != 'a'", "true", t)
	fold("'undefined' == typeof a", "'undefined' == typeof a", t)
	fold("typeof a != 'number'", "typeof a != 'number'", t)
	fold("typeof a == typeof a", "true", t)
	fold("'a' === 'a'", "true", t)
	fold("'b' !== 'a'", "true", t)
	fold("typeof a === typeof a", "true", t)
	fold("typeof a !== typeof a", "false", t)
	fold("'' + x <= '' + y", "'' + x <= '' + y", t)
	fold("'' + x != '' + y", "'' + x != '' + y", t)
	fold("'' + x === '' + y", "'' + x === '' + y", t)

	fold("'' + x <= '' + x", "'' + x <= '' + x", t)   // potentially foldable
	fold("'' + x != '' + x", "'' + x != '' + x", t)   // potentially foldable
	fold("'' + x === '' + x", "'' + x === '' + x", t) // potentially foldable
}

func TestNumberStringComparison(t *testing.T) {
	fold("1 < '2'", "true", t)
	fold("2 > '1'", "true", t)
	fold("123 > '34'", "true", t)
	fold("NaN >= 'NaN'", "false", t)
	fold("1 == '2'", "false", t)
	fold("1 != '1'", "false", t)
	fold("NaN == 'NaN'", "false", t)
	fold("1 === '1'", "false", t)
	fold("1 !== '1'", "true", t)
	fold("+x > '' + y", "+x > '' + y", t)
	fold("+x == '' + y", "+x == '' + y", t)
	fold("+x !== '' + y", "true", t)
}

func TestStringNumberComparison(t *testing.T) {
	fold("'1' < 2", "true", t)
	fold("'2' > 1", "true", t)
	fold("'123' > 34", "true", t)
	fold("'NaN' < NaN", "false", t)
	fold("'1' == 2", "false", t)
	fold("'1' != 1", "false", t)
	fold("'NaN' == NaN", "false", t)
	fold("'1' === 1", "false", t)
	fold("'1' !== 1", "true", t)
	fold("'' + x < +y", "'' + x < +y", t)
	fold("'' + x == +y", "'' + x == +y", t)
	fold("'' + x === +y", "false", t)
}

func TestNaNComparison(t *testing.T) {
	fold("NaN < NaN", "false", t)
	fold("NaN >= NaN", "false", t)
	fold("NaN == NaN", "false", t)
	fold("NaN === NaN", "false", t)

	fold("NaN < null", "false", t)
	fold("null >= NaN", "false", t)
	fold("NaN == null", "false", t)
	fold("null != NaN", "true", t)
	fold("null === NaN", "false", t)

	fold("NaN < undefined", "false", t)
	fold("undefined >= NaN", "false", t)
	fold("NaN == undefined", "false", t)
	fold("undefined != NaN", "true", t)
	fold("undefined === NaN", "false", t)

	fold("NaN < x", "NaN < x", t)
	fold("x >= NaN", "x >= NaN", t)
	fold("NaN == x", "NaN == x", t)
	fold("x != NaN", "x != NaN", t)
	fold("NaN === x", "false", t)
	fold("x !== NaN", "true", t)
	fold("NaN == foo()", "NaN == foo()", t)
}

func TestObjectComparison(t *testing.T) {
	fold("!new Date()", "false", t)
	fold("!!new Date()", "true", t)

	fold("new Date() == null", "false", t)
	fold("new Date() == undefined", "false", t)
	fold("new Date() != null", "true", t)
	fold("new Date() != undefined", "true", t)
	fold("null == new Date()", "false", t)
	fold("undefined == new Date()", "false", t)
	fold("null != new Date()", "true", t)
	fold("undefined != new Date()", "true", t)
}

func TestUnaryOps(t *testing.T) {
	fold("!foo()", "!foo()", t)
	fold("~foo()", "~foo()", t)
	fold("-foo()", "-foo()", t)

	fold("a = !true", "a = false", t)
	fold("a = !10", "a = !10", t)
	fold("a = !false", "a = true", t)
	fold("a = !foo()", "a = !foo()", t)

	fold("a = -Infinity", "a = -Infinity", t)
	fold("a = -NaN", "a = NaN", t)
	fold("a = -foo()", "a = -foo()", t)
	fold("a = ~~0", "a = 0", t)
	fold("a = ~~10", "a = 10", t)
	fold("a = ~-7", "a = 6", t)

	fold("a = +true", "a = 1", t)
	fold("a = +10", "a = 10", t)
	fold("a = +false", "a = 0", t)
	fold("a = +foo()", "a = +foo()", t)
	fold("a = +f", "a = +f", t)

	fold("a = +0", "a = 0", t)
	fold("a = +Infinity", "a = Infinity", t)
	fold("a = +NaN", "a = NaN", t)
	fold("a = +-7", "a = -7", t)
	fold("a = +.5", "a = 0.5", t)

	fold("a = ~0xffffffff", "a = 0", t)
	fold("a = ~~0xffffffff", "a = -1", t)

	// Empty arrays
	fold("+[]", "0", t)
	fold("+[[]]", "0", t)
	fold("+[[[]]]", "0", t)

	// Arrays with one element
	fold("+[1]", "1", t)
	fold("+[[1]]", "1", t)
	fold("+[undefined]", "0", t)
	fold("+[null]", "0", t)
	fold("+[,]", "0", t)

	// Arrays with more than one element
	fold("+[1, 2]", "NaN", t)
	fold("+[[1], 2]", "NaN", t)
	fold("+[,1]", "NaN", t)
	fold("+[,,]", "NaN", t)
}

func TestUnaryOpsStringCompare(t *testing.T) {
	fold("a = -1", "a = -1", t)
	fold("a = ~0", "a = -1", t)
	fold("a = ~1", "a = -2", t)
	fold("a = ~101", "a = -102", t)
}

func TestFoldLogicalOp1(t *testing.T) {
	fold("x = true && x", "x = x", t)
	fold("x = [foo()] && x", "x = (foo(), x)", t)

	fold("x = false && x", "x = false", t)
	fold("x = true || x", "x = true", t)
	fold("x = false || x", "x = x", t)
	fold("x = 0 && x", "x = 0", t)
	fold("x = 3 || x", "x = 3", t)
	fold("x = false || 0", "x = 0", t)

	// Unfoldable cases
	fold("a = x && true", "a = x && true", t)
	fold("a = x && false", "a = x && false", t)
	fold("a = x || 3", "a = x || 3", t)
	fold("a = x || false", "a = x || false", t)
	fold("a = b ? c : x || false", "a = b ? c : x || false", t)
	fold("a = b ? x || false : c", "a = b ? x || false : c", t)
	fold("a = b ? c : x && true", "a = b ? c : x && true", t)
	fold("a = b ? x && true : c", "a = b ? x && true : c", t)

	// Folded but not here
	fold("a = x || false ? b : c", "a = x || false ? b : c", t)
	fold("a = x && true ? b : c", "a = x && true ? b : c", t)
}

func TestFoldLogicalOp2(t *testing.T) {
	fold("x = function(){} && x", "x = x", t)
	fold("x = true && function(){}", "x = function() {}", t)
	fold(
		"x = [(function(){alert(x)})()] && x",
		"x = ((function() { alert(x); })(), x)",
		t,
	)
}

func TestFoldBitwiseOp(t *testing.T) {
	fold("x = 1 & 1", "x = 1", t)
	fold("x = 1 & 2", "x = 0", t)
	fold("x = 3 & 1", "x = 1", t)
	fold("x = 3 & 3", "x = 3", t)

	fold("x = 1 | 1", "x = 1", t)
	fold("x = 1 | 2", "x = 3", t)
	fold("x = 3 | 1", "x = 3", t)
	fold("x = 3 | 3", "x = 3", t)

	fold("x = 1 ^ 1", "x = 0", t)
	fold("x = 1 ^ 2", "x = 3", t)
	fold("x = 3 ^ 1", "x = 2", t)
	fold("x = 3 ^ 3", "x = 0", t)

	fold("x = -1 & 0", "x = 0", t)
	fold("x = 0 & -1", "x = 0", t)
	fold("x = 1 & 4", "x = 0", t)
	fold("x = 2 & 3", "x = 2", t)

	// make sure we fold only when we are supposed to -- not when doing so would
	// lose information or when it is performed on nonsensical arguments.
	fold("x = 1 & 1.1", "x = 1", t)
	fold("x = 1.1 & 1", "x = 1", t)
	fold("x = 1 & 3000000000", "x = 0", t)
	fold("x = 3000000000 & 1", "x = 0", t)

	// Try some cases with | as well
	fold("x = 1 | 4", "x = 5", t)
	fold("x = 1 | 3", "x = 3", t)
	fold("x = 1 | 1.1", "x = 1", t)
	fold("x = 1 | 3E9", "x = 1 | 3E9", t)
	fold("x = 1 | 3000000001", "x = -1294967295", t)
	fold("x = 4294967295 | 0", "x = -1", t)
}

func TestIssue9256(t *testing.T) {
	// Returns -2 prior to fix (Number.MAX_VALUE)
	fold("1.7976931348623157e+308 << 1", "0", t)

	// Isn't changed prior to fix
	fold("1.7976931348623157e+308 << 1.7976931348623157e+308", "0", t)
	fold("1.7976931348623157e+308 >> 1.7976931348623157e+308", "0", t)

	// Panics prior to fix (Number.MIN_VALUE)
	fold("5e-324 >> 5e-324", "0", t)
	fold("5e-324 << 5e-324", "0", t)
	fold("5e-324 << 0", "0", t)
	fold("0 << 5e-324", "0", t)

	// Ensuring overflows are handled correctly
	fold("1 << 31", "-2147483648", t)
	fold("-8 >> 2", "-2", t)
	fold("-8 >>> 2", "1073741822", t)
}

func TestFoldingAdd(t *testing.T) {
	fold("x = null + true", "x = 1", t)
	fold("x = a + true", "x = a + true", t)

	fold("x = false + []", "x = \"false\"", t)
	fold("x = [] + true", "x = \"true\"", t)
	fold("NaN + []", "\"NaN\"", t)
}

func TestFoldBitwiseOpStringCompare(t *testing.T) {
	fold("x = -1 | 0", "x = -1", t)
}

func TestFoldBitShifts(t *testing.T) {
	fold("x = 1 << 0", "x = 1", t)
	fold("x = -1 << 0", "x = -1", t)
	fold("x = 1 << 1", "x = 2", t)
	fold("x = 3 << 1", "x = 6", t)
	fold("x = 1 << 8", "x = 256", t)

	fold("x = 1 >> 0", "x = 1", t)
	fold("x = -1 >> 0", "x = -1", t)
	fold("x = 1 >> 1", "x = 0", t)
	fold("x = 2 >> 1", "x = 1", t)
	fold("x = 5 >> 1", "x = 2", t)
	fold("x = 127 >> 3", "x = 15", t)
	fold("x = 3 >> 1", "x = 1", t)
	fold("x = 3 >> 2", "x = 0", t)
	fold("x = 10 >> 1", "x = 5", t)
	fold("x = 10 >> 2", "x = 2", t)
	fold("x = 10 >> 5", "x = 0", t)

	fold("x = 10 >>> 1", "x = 5", t)
	fold("x = 10 >>> 2", "x = 2", t)
	fold("x = 10 >>> 5", "x = 0", t)
	fold("x = -1 >>> 1", "x = 2147483647", t) // 0x7fffffff
	fold("x = -1 >>> 0", "x = 4294967295", t) // 0xffffffff
	fold("x = -2 >>> 0", "x = 4294967294", t) // 0xfffffffe
	fold("x = 0x90000000 >>> 28", "x = 9", t)

	fold("x = 0xffffffff << 0", "x = -1", t)
	fold("x = 0xffffffff << 4", "x = -16", t)
	fold("1 << 32", "1", t)
	fold("1 << -1", "-2147483648", t)
	fold("1 >> 32", "1", t)
}

func TestFoldBitShiftsStringCompare(t *testing.T) {
	fold("x = -1 << 1", "x = -2", t)
	fold("x = -1 << 8", "x = -256", t)
	fold("x = -1 >> 1", "x = -1", t)
	fold("x = -2 >> 1", "x = -1", t)
	fold("x = -1 >> 0", "x = -1", t)
}

func TestFoldArithmetic(t *testing.T) {
	fold("x = 10 + 20", "x = 30", t)
	fold("x = 2 / 4", "x = 0.5", t)
	fold("x = 2.25 * 3", "x = 6.75", t)
	fold("z = x * y", "z = x * y", t)
	fold("x = y * 5", "x = y * 5", t)
	fold("x = 1 / 0", "x = 1 / 0", t)
	fold("x = 3 % 2", "x = 1", t)
	fold("x = 3 % -2", "x = 1", t)
	fold("x = -1 % 3", "x = -1", t)
	fold("x = 1 % 0", "x = 1 % 0", t)
	fold("x = 2 ** 3", "x = 8", t)
	fold("x = 2 ** 55", "x = 2 ** 55", t)
	fold("x = 3 ** -1", "x = 3 ** -1", t)

	fold("x = y + 10 + 20", "x = y + 10 + 20", t)
	fold("x = y / 2 / 4", "x = y / 2 / 4", t)
	fold("x = y * 2.25 * 3", "x = y * 6.75", t)
	fold("z = x * y", "z = x * y", t)
	fold("x = y * 5", "x = y * 5", t)
	fold("x = y + (z * 24 * 60 * 60 * 1000)", "x = y + z * 86400000", t)

	fold("x = null * undefined", "x = NaN", t)
	fold("x = null * 1", "x = 0", t)
	fold("x = (null - 1) * 2", "x = -2", t)
	fold("x = (null + 1) * 2", "x = 2", t)
	fold("x = null ** 0", "x = 1", t)
}

func TestFoldArithmeticInfinity(t *testing.T) {
	fold("x = -Infinity - 2", "x = -Infinity", t)
	fold("x = Infinity - 2", "x = Infinity", t)
	fold("x = Infinity * 5", "x = Infinity", t)
	fold("x = Infinity ** 2", "x = Infinity", t)
	fold("x = Infinity ** -2", "x = 0", t)
}

func TestFoldArithmeticStringComp(t *testing.T) {
	fold("x = 10 - 20", "x = -10", t)
}

func TestFoldComparison(t *testing.T) {
	// Equality checks
	fold("x = 0 == 0", "x = true", t)
	fold("x = 1 == 2", "x = false", t)
	fold("x = 'abc' == 'def'", "x = false", t)
	fold("x = 'abc' == 'abc'", "x = true", t)
	fold("x = \"\" == ''", "x = true", t)
	fold("x = foo() == bar()", "x = foo() == bar()", t)

	// Inequality checks
	fold("x = 1 != 0", "x = true", t)
	fold("x = 'abc' != 'def'", "x = true", t)
	fold("x = 'a' != 'a'", "x = false", t)

	// Relational comparisons
	fold("x = 1 < 20", "x = true", t)
	fold("x = 3 < 3", "x = false", t)
	fold("x = 10 > 1.0", "x = true", t)
	fold("x = 10 > 10.25", "x = false", t)
	fold("x = y == y", "x = y == y", t) // Maybe foldable given type information
	fold("x = y < y", "x = false", t)
	fold("x = y > y", "x = false", t)
	fold("x = 1 <= 1", "x = true", t)
	fold("x = 1 <= 0", "x = false", t)
	fold("x = 0 >= 0", "x = true", t)
	fold("x = -1 >= 9", "x = false", t)

	// Boolean comparisons
	fold("x = true == true", "x = true", t)
	fold("x = false == false", "x = true", t)
	fold("x = false == null", "x = false", t)
	fold("x = false == true", "x = false", t)
	fold("x = true == null", "x = false", t)

	// Standalone equality checks
	fold("0 == 0", "true", t)
	fold("1 == 2", "false", t)
	fold("'abc' == 'def'", "false", t)
	fold("'abc' == 'abc'", "true", t)
	fold("\"\" == ''", "true", t)
	fold("foo() == bar()", "foo() == bar()", t)

	// Standalone inequality checks
	fold("1 != 0", "true", t)
	fold("'abc' != 'def'", "true", t)
	fold("'a' != 'a'", "false", t)

	// Standalone relational comparisons
	fold("1 < 20", "true", t)
	fold("3 < 3", "false", t)
	fold("10 > 1.0", "true", t)
	fold("10 > 10.25", "false", t)
	fold("x == x", "x == x", t)
	fold("x < x", "false", t)
	fold("x > x", "false", t)
	fold("1 <= 1", "true", t)
	fold("1 <= 0", "false", t)
	fold("0 >= 0", "true", t)
	fold("-1 >= 9", "false", t)

	// Boolean standalone comparisons
	fold("true == true", "true", t)
	fold("false == null", "false", t)
	fold("false == true", "false", t)
	fold("true == null", "false", t)
}

func TestFoldComparison2(t *testing.T) {
	// === and !== comparisons
	fold("x = 0 === 0", "x = true", t)
	fold("x = 1 === 2", "x = false", t)
	fold("x = 'abc' === 'def'", "x = false", t)
	fold("x = 'abc' === 'abc'", "x = true", t)
	fold("x = \"\" === ''", "x = true", t)
	fold("x = foo() === bar()", "x = foo() === bar()", t)

	fold("x = 1 !== 0", "x = true", t)
	fold("x = 'abc' !== 'def'", "x = true", t)
	fold("x = 'a' !== 'a'", "x = false", t)

	fold("x = y === y", "x = y === y", t)

	fold("x = true === true", "x = true", t)
	fold("x = false === false", "x = true", t)
	fold("x = false === null", "x = false", t)
	fold("x = false === true", "x = false", t)
	fold("x = true === null", "x = false", t)

	fold("0 === 0", "true", t)
	fold("1 === 2", "false", t)
	fold("'abc' === 'def'", "false", t)
	fold("'abc' === 'abc'", "true", t)
	fold("\"\" === ''", "true", t)
	fold("foo() === bar()", "foo() === bar()", t)

	fold("1 === '1'", "false", t)
	fold("1 === true", "false", t)
	fold("1 !== '1'", "true", t)
	fold("1 !== true", "true", t)

	fold("1 !== 0", "true", t)
	fold("'abc' !== 'def'", "true", t)
	fold("'a' !== 'a'", "false", t)

	fold("x === x", "x === x", t)

	fold("true === true", "true", t)
	fold("false === null", "false", t)
	fold("false === true", "false", t)
	fold("true === null", "false", t)
}

func TestFoldComparison3(t *testing.T) {
	fold("x = !1 == !0", "x = false", t)

	fold("x = !0 == !0", "x = true", t)
	fold("x = !1 == !1", "x = true", t)
	fold("x = !1 == null", "x = false", t)
	fold("x = !1 == !0", "x = false", t)
	fold("x = !0 == null", "x = false", t)

	fold("!0 == !0", "true", t)
	fold("!1 == null", "false", t)
	fold("!1 == !0", "false", t)
	fold("!0 == null", "false", t)

	fold("x = !0 === !0", "x = true", t)
	fold("x = !1 === !1", "x = true", t)
	fold("x = !1 === null", "x = false", t)
	fold("x = !1 === !0", "x = false", t)
	fold("x = !0 === null", "x = false", t)

	fold("!0 === !0", "true", t)
	fold("!1 === null", "false", t)
	fold("!1 === !0", "false", t)
	fold("!0 === null", "false", t)
}

func TestFoldComparison4(t *testing.T) {
	fold("[] == false", "[] == false", t)
	fold("[] == true", "[] == true", t)
	fold("[0] == false", "[0] == false", t)
	fold("[0] == true", "[0] == true", t)
	fold("[1] == false", "[1] == false", t)
	fold("[1] == true", "[1] == true", t)
	fold("({}) == false", "({}) == false", t)
	fold("({}) == true", "({}) == true", t)
}

func TestFoldGetElem1(t *testing.T) {
	fold("x = [,10][0]", "x = (0, void 0)", t)
	fold("x = [10, 20][0]", "x = (0, 10)", t)
	fold("x = [10, 20][1]", "x = (0, 20)", t)

	// Uncomment these if folding for out-of-bound indices is implemented
	// fold("x = [10, 20][-1]", "x = void 0;", t)
	// fold("x = [10, 20][2]", "x = void 0;", t)

	fold("x = [foo(), 0][1]", "x = (foo(), 0)", t)
	fold("x = [0, foo()][1]", "x = (0, foo())", t)
	// Uncomment if folding for first index with side effects is implemented
	// fold("x = [0, foo()][0]", "x = (foo(), 0);", t)

	// Uncomment if parser is fixed to be able to parse this
	// fold("for([1][0] in {});", "for([1][0] in {})", t)

	fold("x = 'string'[5]", "x = \"g\"", t)
	fold("x = 'string'[0]", "x = \"s\"", t)
	fold("x = 's'[0]", "x = \"s\"", t)
	fold("x = '\\uD83D\\uDCA9'[0]", "x = \"\\uD83D\"", t)
	fold("x = '\\uD83D\\uDCA9'[1]", "x = \"\\uDCA9\"", t)
	fold("x = \"💩\"[0]", "x = \"\\uD83D\"", t)
	fold("x = \"💩\"[1]", "x = \"\\uDCA9\"", t)
}

func TestFoldArrayLitSpreadGetElem(t *testing.T) {
	fold("x = [...[0    ]][0]", "x = (0, 0)", t)
	fold("x = [0, 1, ...[2, 3, 4]][3]", "x = (0, 3)", t)
	fold("x = [...[0, 1], 2, ...[3, 4]][3]", "x = (0, 3)", t)
	fold("x = [...[...[0, 1], 2, 3], 4][0]", "x = (0, 0)", t)
	fold("x = [...[...[0, 1], 2, 3], 4][3]", "x = (0, 3)", t)

	// Uncomment these cases if support for out-of-bound indices is added
	// fold("x = [...[]][100]", "x = void 0;", t)
	// fold("x = [...[0]][100]", "x = void 0;", t)
}

func TestDontFoldNonLiteralSpreadGetElem(t *testing.T) {
	fold("x = [...iter][0];", "x = [...iter][0]", t)
	fold("x = [0, 1, ...iter][2];", "x = [0, 1, ...iter][2]", t)
	fold("x = [0, 1, ...iter][0];", "x = [0, 1, ...iter][0]", t)
}

func TestFoldObjectSpread(t *testing.T) {
	fold("x = {...{}}", "x = {}", t)
	fold("x = {a, ...{}, b}", "x = { a, b }", t)
	fold("x = {...{a, b}, c, ...{d, e}}", "x = { a, b, c, d, e }", t)
	fold("x = {...{...{a}, b}, c}", "x = { a, b, c }", t)
	fold("({...{x}} = obj)", "({...{ x }} = obj)", t)
}

func TestFoldComplex(t *testing.T) {
	fold("x = (3 / 1.0) + (1 * 2)", "x = 5", t)
	fold("x = (1 == 1.0) && foo() && true", "x = foo() && true", t)
	fold("x = 'abc' + 5 + 10", "x = \"abc510\"", t)
}

func TestFoldArrayLength(t *testing.T) {
	// Can fold
	fold("x = [].length", "x = 0", t)
	fold("x = [1,2,3].length", "x = 3", t)
	fold("x = [a,b].length", "x = 2", t)

	// Not handled yet
	fold("x = [,,1].length", "x = 3", t)

	// Cannot fold
	fold("x = [foo(), 0].length", "x = [foo(), 0].length", t)
	fold("x = y.length", "x = y.length", t)
}

func TestFoldStringLength(t *testing.T) {
	// Can fold basic strings
	fold("x = ''.length", "x = 0", t)
	fold("x = '123'.length", "x = 3", t)

	// Test Unicode escapes are accounted for
	fold("x = '123\\u01dc'.length", "x = 4", t)
}

func TestFoldTypeof(t *testing.T) {
	fold("x = typeof 1", "x = \"number\"", t)
	fold("x = typeof 'foo'", "x = \"string\"", t)
	fold("x = typeof true", "x = \"boolean\"", t)
	fold("x = typeof false", "x = \"boolean\"", t)
	fold("x = typeof null", "x = \"object\"", t)
	fold("x = typeof undefined", "x = \"undefined\"", t)
	fold("x = typeof void 0", "x = \"undefined\"", t)
	fold("x = typeof []", "x = \"object\"", t)
	fold("x = typeof [1]", "x = \"object\"", t)
	fold("x = typeof [1,[]]", "x = \"object\"", t)
	fold("x = typeof {}", "x = \"object\"", t)
	fold("x = typeof function() {}", "x = \"function\"", t)

	fold("x = typeof [1, [foo()]]", "x = typeof [1, [foo()]]", t)
	fold("x = typeof { bathwater: baby() }", "x = typeof { bathwater: baby() }", t)
}

func TestFoldInstanceOf(t *testing.T) {
	// Non-object types are never instances of anything
	fold("64 instanceof Object", "false", t)
	fold("64 instanceof Number", "false", t)
	fold("'' instanceof Object", "false", t)
	fold("'' instanceof String", "false", t)
	fold("true instanceof Object", "false", t)
	fold("true instanceof Boolean", "false", t)
	fold("!0 instanceof Object", "false", t)
	fold("!0 instanceof Boolean", "false", t)
	fold("false instanceof Object", "false", t)
	fold("null instanceof Object", "false", t)
	fold("undefined instanceof Object", "false", t)
	fold("NaN instanceof Object", "false", t)
	fold("Infinity instanceof Object", "false", t)

	// Array and object literals are known to be objects
	fold("[] instanceof Object", "true", t)
	fold("({}) instanceof Object", "true", t)

	// These cases are foldable but not currently handled
	fold("new Foo() instanceof Object", "new Foo(), true", t)

	// These would require type information to fold
	fold("[] instanceof Foo", "[] instanceof Foo", t)
	fold("({}) instanceof Foo", "({}) instanceof Foo", t)

	fold("(function() {}) instanceof Object", "true", t)

	// An unknown value should never be folded
	fold("x instanceof Foo", "x instanceof Foo", t)
	fold("x instanceof Object", "x instanceof Object", t)
}

func TestDivision(t *testing.T) {
	// Ensure 1/3 does not expand to 0.333333
	fold("print(1 / 3)", "print(1 / 3)", t)

	// Prefer decimal form when strings are the same length
	fold("print(1 / 2)", "print(0.5)", t)
}

func TestAssignOpsEarly(t *testing.T) {
	fold("x = x + y", "x = x + y", t)
	fold("x = y + x", "x = y + x", t)
	fold("x = x * y", "x = x * y", t)
	fold("x = y * x", "x = y * x", t)
	fold("x.y = x.y + z", "x.y = x.y + z", t)
	fold("next().x = next().x + 1", "next().x = next().x + 1", t)

	fold("x = x - y", "x = x - y", t)
	fold("x = y - x", "x = y - x", t)
	fold("x = x | y", "x = x | y", t)
	fold("x = y | x", "x = y | x", t)
	fold("x = x * y", "x = x * y", t)
	fold("x = y * x", "x = y * x", t)
	fold("x = x ** y", "x = x ** y", t)
	fold("x = y ** 2", "x = y ** 2", t)
	fold("x.y = x.y + z", "x.y = x.y + z", t)
	fold("next().x = next().x + 1", "next().x = next().x + 1", t)
}

func TestFoldAdd(t *testing.T) {
	fold("x = false + 1", "x = 1", t)
	fold("x = true + 1", "x = 2", t)
	fold("x = 1 + false", "x = 1", t)
	fold("x = 1 + true", "x = 2", t)
}

func TestFoldLiteralNames(t *testing.T) {
	fold("NaN == NaN", "false", t)
	fold("Infinity == Infinity", "true", t)
	fold("Infinity == NaN", "false", t)
	fold("undefined == NaN", "false", t)
	fold("undefined == Infinity", "false", t)

	fold("Infinity >= Infinity", "true", t)
	fold("NaN >= NaN", "false", t)
}

func TestFoldLiteralsTypeMismatches(t *testing.T) {
	// Equality checks
	fold("true == true", "true", t)
	fold("true == false", "false", t)
	fold("true == null", "false", t)
	fold("false == null", "false", t)

	// Relational operators convert operands
	fold("null <= null", "true", t) // 0 = 0
	fold("null >= null", "true", t)
	fold("null > null", "false", t)
	fold("null < null", "false", t)

	fold("false >= null", "true", t) // 0 = 0
	fold("false <= null", "true", t)
	fold("false > null", "false", t)
	fold("false < null", "false", t)

	fold("true >= null", "true", t) // 1 > 0
	fold("true <= null", "false", t)
	fold("true > null", "true", t)
	fold("true < null", "false", t)

	fold("true >= false", "true", t) // 1 > 0
	fold("true <= false", "false", t)
	fold("true > false", "true", t)
	fold("true < false", "false", t)
}

func TestFoldSimpleArithmeticOp(t *testing.T) {
	fold("x * NaN", "x * NaN", t)
	fold("NaN / y", "NaN / y", t)
	fold("f(x) - 0", "f(x) - 0", t)
	fold("f(x) * 1", "f(x) * 1", t)
	fold("1 * f(x)", "1 * f(x)", t)
	fold("0 + a + b", "0 + a + b", t)
	fold("0 - a - b", "0 - a - b", t)
	fold("a + b - 0", "a + b - 0", t)
	fold("(1 + x) * NaN", "(1 + x) * NaN", t)

	fold("(1 + f(x)) * NaN", "(1 + f(x)) * NaN", t) // don't fold side-effects
}

func TestFoldArithmeticWithStrings(t *testing.T) {
	// Left side of expression is a string
	fold("'10' - 5", "5", t)
	fold("'4' / 2", "2", t)
	fold("'11' % 2", "1", t)
	fold("'10' ** 2", "100", t)
	fold("'Infinity' * 2", "Infinity", t)
	fold("'NaN' * 2", "NaN", t)

	// Right side of expression is a string
	fold("10 - '5'", "5", t)
	fold("4 / '2'", "2", t)
	fold("11 % '2'", "1", t)
	fold("10 ** '2'", "100", t)
	fold("2 * 'Infinity'", "Infinity", t)
	fold("2 * 'NaN'", "NaN", t)

	// Both sides are strings
	fold("'10' - '5'", "5", t)
	fold("'4' / '2'", "2", t)
	fold("'11' % '2'", "1", t)
	fold("'10' ** '2'", "100", t)
	fold("'Infinity' * '2'", "Infinity", t)
	fold("'NaN' * '2'", "NaN", t)
}

func TestNotFoldBackToTrueFalse(t *testing.T) {
	fold("!0", "!0", t) // fold_same
	fold("!1", "!1", t) // fold_same
	fold("!3", "!3", t) // fold_same
}

func TestFoldBangConstants(t *testing.T) {
	fold("1 + !0", "2", t)
	fold("1 + !1", "1", t)
	fold("'a ' + !1", "\"a false\"", t)
	fold("'a ' + !0", "\"a true\"", t)
}

func TestFoldMixed(t *testing.T) {
	fold("'' + [1]", "\"1\"", t)
	fold("false + []", "\"false\"", t)
}

func TestFoldVoid(t *testing.T) {
	fold("void 0", "void 0", t) // fold_same
	fold("void 1", "void 0", t)
	fold("void x", "void 0", t)
	fold("void x()", "void x()", t) // fold_same
}

func TestObjectLiteral(t *testing.T) {
	fold("(!{})", "false", t)
	fold("(!{a:1})", "false", t)
	fold("(!{a:foo()})", "foo(), false", t)
	fold("(!{'a':foo()})", "foo(), false", t)
}

func TestArrayLiteral(t *testing.T) {
	fold("(![])", "false", t)
	fold("(![1])", "false", t)
	fold("(![a])", "false", t)
	fold("foo(), false;", "foo(), false", t) // fold_same
}
