package evaluator

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"
)

const builtinStringTrimWhitespace = "\u0009\u000A\u000B\u000C\u000D\u0020\u00A0\u1680\u180E\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200A\u2028\u2029\u202F\u205F\u3000\uFEFF"

var stringToNumberParseInteger = regexp.MustCompile(`^(?:0[xX])`)

func parseNumber(value string) float64 {
	value = strings.Trim(value, builtinStringTrimWhitespace)

	if value == "" {
		return 0
	}

	var parseFloat bool
	switch {
	case strings.ContainsRune(value, '.'):
		parseFloat = true
	case stringToNumberParseInteger.MatchString(value):
		parseFloat = false
	default:
		parseFloat = true
	}

	if parseFloat {
		number, err := strconv.ParseFloat(value, 64)
		if err != nil && !errors.Is(err, strconv.ErrRange) {
			return math.NaN()
		}
		return number
	}

	number, err := strconv.ParseInt(value, 0, 64)
	if err != nil {
		return math.NaN()
	}
	return float64(number)
}

type valueKind int

const (
	valueUndefined valueKind = iota
	valueNull
	valueNumber
	valueString
	valueBoolean
	valueObject
)

var (
	nullValue  = Value{kind: valueNull}
	falseValue = Value{kind: valueBoolean, value: false}
	trueValue  = Value{kind: valueBoolean, value: true}
)

// Value is the representation of a JavaScript value.
type Value struct {
	value interface{}
	kind  valueKind
}

var matchLeading0Exponent = regexp.MustCompile(`([eE][\+\-])0+([1-9])`) // 1e-07 => 1e-7

func floatToString(value float64, bitsize int) string {
	// TODO Fit to ECMA-262 9.8.1 specification
	if math.IsNaN(value) {
		return "NaN"
	} else if math.IsInf(value, 0) {
		if math.Signbit(value) {
			return "-Infinity"
		}
		return "Infinity"
	}
	exponent := math.Log10(math.Abs(value))
	if exponent >= 21 || exponent < -6 {
		return matchLeading0Exponent.ReplaceAllString(strconv.FormatFloat(value, 'g', -1, bitsize), "$1$2")
	}
	return strconv.FormatFloat(value, 'f', -1, bitsize)
}

func (v Value) string() string {
	if v.kind == valueString {
		switch value := v.value.(type) {
		case string:
			return value
		case []uint16:
			return string(utf16.Decode(value))
		}
	}
	if v.IsUndefined() {
		return "undefined"
	}
	if v.IsNull() {
		return "null"
	}
	switch value := v.value.(type) {
	case bool:
		return strconv.FormatBool(value)
	case int:
		return strconv.FormatInt(int64(value), 10)
	case int8:
		return strconv.FormatInt(int64(value), 10)
	case int16:
		return strconv.FormatInt(int64(value), 10)
	case int32:
		return strconv.FormatInt(int64(value), 10)
	case int64:
		return strconv.FormatInt(value, 10)
	case uint:
		return strconv.FormatUint(uint64(value), 10)
	case uint8:
		return strconv.FormatUint(uint64(value), 10)
	case uint16:
		return strconv.FormatUint(uint64(value), 10)
	case uint32:
		return strconv.FormatUint(uint64(value), 10)
	case uint64:
		return strconv.FormatUint(value, 10)
	case float32:
		if value == 0 {
			return "0" // Take care not to return -0
		}
		return floatToString(float64(value), 32)
	case float64:
		if value == 0 {
			return "0" // Take care not to return -0
		}
		return floatToString(value, 64)
	case []uint16:
		return string(utf16.Decode(value))
	case string:
		return value
	}
	panic(fmt.Errorf("%v.string( %T)", v.value, v.value))
}

func (v Value) float64() float64 {
	switch v.kind {
	case valueUndefined:
		return math.NaN()
	case valueNull:
		return 0
	}
	switch value := v.value.(type) {
	case bool:
		if value {
			return 1
		}
		return 0
	case int:
		return float64(value)
	case int8:
		return float64(value)
	case int16:
		return float64(value)
	case int32:
		return float64(value)
	case int64:
		return float64(value)
	case uint:
		return float64(value)
	case uint8:
		return float64(value)
	case uint16:
		return float64(value)
	case uint32:
		return float64(value)
	case uint64:
		return float64(value)
	case float64:
		return value
	case string:
		return parseNumber(value)
	}
	panic(fmt.Errorf("toFloat(%T)", v.value))
}

func (v Value) bool() bool {
	if v.kind == valueBoolean {
		return v.value.(bool)
	}
	if v.IsUndefined() || v.IsNull() {
		return false
	}
	switch value := v.value.(type) {
	case bool:
		return value
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(value).Int() != 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(value).Uint() != 0
	case float32:
		return value != 0
	case float64:
		if math.IsNaN(value) || value == 0 {
			return false
		}
		return true
	case string:
		return len(value) != 0
	case []uint16:
		return len(utf16.Decode(value)) != 0
	}
	panic(fmt.Sprintf("unexpected boolean type %T", v.value))
}

// IsBoolean will return true if value is a boolean (primitive).
func (v Value) IsBoolean() bool {
	return v.kind == valueBoolean
}

// IsNumber will return true if value is a number (primitive).
func (v Value) IsNumber() bool {
	return v.kind == valueNumber
}

// IsString will return true if value is a string (primitive).
func (v Value) IsString() bool {
	return v.kind == valueString
}

// IsNull will return true if the value is null, and false otherwise.
func (v Value) IsNull() bool {
	return v.kind == valueNull
}

// IsUndefined will return true if the value is undefined, and false otherwise.
func (v Value) IsUndefined() bool {
	return v.kind == valueUndefined
}

func toValue(value interface{}) Value {
	switch value := value.(type) {
	case Value:
		return value
	case bool:
		return Value{kind: valueBoolean, value: value}
	case int:
		return Value{kind: valueNumber, value: value}
	case int8:
		return Value{kind: valueNumber, value: value}
	case int16:
		return Value{kind: valueNumber, value: value}
	case int32:
		return Value{kind: valueNumber, value: value}
	case int64:
		return Value{kind: valueNumber, value: value}
	case uint:
		return Value{kind: valueNumber, value: value}
	case uint8:
		return Value{kind: valueNumber, value: value}
	case uint16:
		return Value{kind: valueNumber, value: value}
	case uint32:
		return Value{kind: valueNumber, value: value}
	case uint64:
		return Value{kind: valueNumber, value: value}
	case float32:
		return Value{kind: valueNumber, value: float64(value)}
	case float64:
		return Value{kind: valueNumber, value: value}
	case []uint16:
		return Value{kind: valueString, value: value}
	case string:
		return Value{kind: valueString, value: value}
	}
	panic(fmt.Errorf("toValue(%T)", value))
}

// ECMA 262: 9.5.
func toInt32(value Value) int32 {
	switch value := value.value.(type) {
	case int8:
		return int32(value)
	case int16:
		return int32(value)
	case int32:
		return value
	}

	floatValue := value.float64()
	if math.IsNaN(floatValue) || math.IsInf(floatValue, 0) || floatValue == 0 {
		return 0
	}

	// Convert to int64 before int32 to force correct wrapping.
	return int32(int64(floatValue))
}

func toUint32(value Value) uint32 {
	switch value := value.value.(type) {
	case int8:
		return uint32(value)
	case int16:
		return uint32(value)
	case uint8:
		return uint32(value)
	case uint16:
		return uint32(value)
	case uint32:
		return value
	}

	floatValue := value.float64()
	if math.IsNaN(floatValue) || math.IsInf(floatValue, 0) || floatValue == 0 {
		return 0
	}

	// Convert to int64 before uint32 to force correct wrapping.
	return uint32(int64(floatValue))
}

var (
	nan              float64 = math.NaN()
	positiveInfinity float64 = math.Inf(+1)
	negativeInfinity float64 = math.Inf(-1)
	positiveZero     float64 = 0
	negativeZero     float64 = math.Float64frombits(0 | (1 << 63))
)

// NaNValue will return a value representing NaN.
//
// It is equivalent to:
//
//	ToValue(math.NaN())
func NaNValue() Value {
	return Value{kind: valueNumber, value: nan}
}

func positiveInfinityValue() Value {
	return Value{kind: valueNumber, value: positiveInfinity}
}

func negativeInfinityValue() Value {
	return Value{kind: valueNumber, value: negativeInfinity}
}

func positiveZeroValue() Value {
	return Value{kind: valueNumber, value: positiveZero}
}

func negativeZeroValue() Value {
	return Value{kind: valueNumber, value: negativeZero}
}

func stringValue(value string) Value {
	return Value{
		kind:  valueString,
		value: value,
	}
}

func float64Value(value float64) Value {
	return Value{
		kind:  valueNumber,
		value: value,
	}
}

func boolValue(value bool) Value {
	return Value{
		kind:  valueBoolean,
		value: value,
	}
}

func int32Value(value int32) Value {
	return Value{
		kind:  valueNumber,
		value: value,
	}
}

func uint32Value(value uint32) Value {
	return Value{
		kind:  valueNumber,
		value: value,
	}
}
