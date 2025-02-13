package ext

import "slices"

// Ref: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Object
var objectSymbols = []string{
	// Constructor
	"constructor",
	// Properties
	"__proto__",
	// Methods
	"__defineGetter__", "__defineSetter__", "__lookupGetter__", "__lookupSetter__", "hasOwnProperty", "isPrototypeOf",
	"propertyIsEnumerable", "toLocaleString", "toString", "valueOf",
	// removed, but kept in as these are often checked and polyfilled
	"watch", "unwatch",
}

// Ref: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array
var arraySymbols = []string{
	// Constructor
	"constructor",
	// Properties
	"length",
	// Methods
	"at", "concat", "copyWithin", "entries", "every", "fill", "filter", "find",
	"findIndex", "findLast", "findLastIndex", "flat", "flatMap", "forEach", "includes", "indexOf", "join",
	"keys", "lastIndexOf", "map", "pop", "push", "reduce", "reduceRight", "reverse", "shift",
	"slice", "some", "sort", "splice", "toLocaleString", "toReversed", "toSorted", "toSpliced", "toString",
	"unshift", "values", "with",
}

// Ref: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/String
var stringSymbols = []string{
	// Constructor
	"constructor",
	// Properties
	"length",
	// Methods
	"anchor", "at", "big", "blink", "bold", "charAt", "charCodeAt", "codePointAt", "concat",
	"endsWith", "fixed", "fontcolor", "fontsize", "includes", "indexOf", "isWellFormed", "italics", "lastIndexOf",
	"link", "localeCompare", "match", "matchAll", "normalize", "padEnd", "padStart", "repeat", "replace",
	"replaceAll", "search", "slice", "small", "split", "startsWith", "strike", "sub", "substr",
	"substring", "sup", "toLocaleLowerCase", "toLocaleUpperCase", "toLowerCase", "toString", "toUpperCase", "toWellFormed", "trim",
	"trimEnd", "trimStart", "valueOf",
}

// Ref: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number
var numberSymbols = []string{
	// Constructor
	"constructor",
	// Methods
	"toExponential", "toFixed", "toLocaleString", "toPrecision", "toString", "valueOf",
}

// Ref: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Boolean
var booleanSymbols = []string{
	// Constructor
	"constructor",
	// Methods
	"toString", "valueOf",
}

func IsObjectSymbol(sym string) bool {
	return slices.Contains(objectSymbols, sym)
}

func IsArraySymbol(sym string) bool {
	return slices.Contains(arraySymbols, sym) || IsObjectSymbol(sym)
}

func IsStringSymbol(sym string) bool {
	return slices.Contains(stringSymbols, sym) || IsObjectSymbol(sym)
}

func IsNumberSymbol(sym string) bool {
	return slices.Contains(numberSymbols, sym) || IsObjectSymbol(sym)
}

func IsBooleanSymbol(sym string) bool {
	return slices.Contains(booleanSymbols, sym) || IsObjectSymbol(sym)
}
