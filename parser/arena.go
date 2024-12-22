package parser

import (
	"unsafe"
)

type miniArena[T any] struct {
	elementSize uintptr

	a     unsafe.Pointer
	len   uintptr
	index uintptr
}

func newArena[T any](startLen int) *miniArena[T] {
	var t T
	return &miniArena[T]{
		elementSize: unsafe.Sizeof(t),
		len:         uintptr(startLen),
		a:           unsafe.Pointer(&make([]T, startLen)[0]),
	}
}

func (a *miniArena[T]) make() *T {
	n := (*T)(unsafe.Add(a.a, a.index*a.elementSize))
	if a.index++; a.index == a.len {
		a.resize()
	}

	return n
}

func (a *miniArena[T]) resize() {
	a.len = uintptr(float64(a.len) * 1.5)

	a.a = unsafe.Pointer(&make([]T, a.len)[0])
	a.index = 0
}
