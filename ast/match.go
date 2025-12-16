package ast

import (
	"reflect"
)

type Any struct{}

func AnyNode() *Any { return &Any{} }

type Capture struct {
	Out any
}

func CaptureNode(out any) *Capture {
	return &Capture{Out: out}
}

func Match(actual, pattern any) bool {
	return matchValue(reflect.ValueOf(actual), reflect.ValueOf(pattern))
}

var (
	anyType          = reflect.TypeOf(Any{})
	captureType      = reflect.TypeOf(Capture{})
	idxType          = reflect.TypeOf(Idx(0))
	scopeContextType = reflect.TypeOf(ScopeContext(0))
)

func matchValue(actual, pattern reflect.Value) bool {
	for pattern.Kind() == reflect.Interface {
		if pattern.IsNil() {
			return true
		}
		pattern = pattern.Elem()
	}
	for actual.Kind() == reflect.Interface {
		if actual.IsNil() {
			return false
		}
		actual = actual.Elem()
	}

	if pattern.Kind() == reflect.Pointer {
		if pattern.IsNil() {
			return actual.Kind() == reflect.Pointer && actual.IsNil()
		}

		switch pattern.Elem().Type() {
		case anyType:
			return true
		case captureType:
			return bindCapture(actual, pattern.Interface().(*Capture).Out)
		}

		if actual.Kind() != reflect.Pointer {
			return false
		}

		return matchValue(actual.Elem(), pattern.Elem())
	}

	if actual.Type() != pattern.Type() {
		return false
	}

	switch pattern.Kind() {
	case reflect.Struct:
		t := pattern.Type()
		for i := 0; i < pattern.NumField(); i++ {
			sf := t.Field(i)
			if sf.PkgPath != "" {
				continue
			}
			if sf.Type == idxType || sf.Type == scopeContextType {
				continue
			}
			if !matchValue(actual.Field(i), pattern.Field(i)) {
				return false
			}
		}
		return true
	case reflect.Slice, reflect.Array:
		if actual.Len() != pattern.Len() {
			return false
		}
		for i := 0; i < pattern.Len(); i++ {
			if !matchValue(actual.Index(i), pattern.Index(i)) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(actual.Interface(), pattern.Interface())
	}
}

func bindCapture(actual reflect.Value, out any) bool {
	outV := reflect.ValueOf(out)
	if outV.Kind() != reflect.Pointer || outV.IsNil() {
		return false
	}
	dst := outV.Elem()
	if dst.Kind() != reflect.Pointer {
		return false
	}

	dstType := dst.Type()

	if actual.Type().AssignableTo(dstType) {
		dst.Set(actual)
		return true
	}

	if actual.Kind() != reflect.Pointer && dstType.Kind() == reflect.Pointer && actual.CanAddr() {
		if addr := actual.Addr(); addr.Type().AssignableTo(dstType) {
			dst.Set(addr)
			return true
		}
	}

	return false
}

func (*Any) Idx0() Idx                 { return 0 }
func (*Any) Idx1() Idx                 { return 0 }
func (*Any) VisitWith(Visitor)         {}
func (*Any) VisitChildrenWith(Visitor) {}
func (*Any) _expr()                    {}
func (*Any) _stmt()                    {}

func (*Capture) Idx0() Idx                 { return 0 }
func (*Capture) Idx1() Idx                 { return 0 }
func (*Capture) VisitWith(Visitor)         {}
func (*Capture) VisitChildrenWith(Visitor) {}
func (*Capture) _expr()                    {}
func (*Capture) _stmt()                    {}
