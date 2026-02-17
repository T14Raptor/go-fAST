package parser

import (
	"unsafe"
)

// miniArena is a typed bump allocator that hands out pointers into
// pre-allocated slices of T. When a chunk fills up, a new chunk is
// allocated at 1.5x the previous size.
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

// makeSlice allocates n contiguous elements from the arena and returns a
// slice whose backing array lives in arena memory. If the current chunk
// doesn't have enough room, a new chunk is allocated that is large enough.
func (a *miniArena[T]) makeSlice(n int) []T {
	if n == 0 {
		return nil
	}
	un := uintptr(n)
	if a.index+un > a.len {
		// Need a new chunk that fits at least n elements.
		newLen := uintptr(float64(a.len) * 1.5)
		if newLen < un {
			newLen = un
		}
		a.len = newLen
		a.a = unsafe.Pointer(&make([]T, newLen)[0])
		a.index = 0
	}
	start := unsafe.Add(a.a, a.index*a.elementSize)
	a.index += un
	if a.index == a.len {
		a.resize()
	}
	return unsafe.Slice((*T)(start), n)
}
