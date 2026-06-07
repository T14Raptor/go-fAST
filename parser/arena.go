package parser

import (
	"unsafe"
)

// arena is a typed bump allocator that hands out pointers into
// pre-allocated slices of T. When a chunk fills up, a new chunk is
// allocated at 1.5x the previous size.
//
// The arena uses lazy initialization: the first chunk is not allocated
// until the first make() / makeSlice() call. The sentinel for "needs
// init" is index == len at zero (a.a == nil). newArena sets index = len
// = startLen so the first make() lands in the slow path and allocates a
// chunk of the requested initial size.
type arena[T any] struct {
	elementSize uintptr

	a     unsafe.Pointer
	len   uintptr // initial chunk size before first alloc; current cap after
	index uintptr // initially equals len (sentinel: forces resize on first use)
}

func newArena[T any](startLen int) arena[T] {
	var t T
	return arena[T]{
		elementSize: unsafe.Sizeof(t),
		len:         uintptr(startLen),
		index:       uintptr(startLen), // sentinel — first make() triggers resize
		// a is nil; resize allocates the first chunk on demand
	}
}

func (a *arena[T]) make() *T {
	if a.index == a.len {
		a.resize()
	}
	n := (*T)(unsafe.Add(a.a, a.index*a.elementSize))
	a.index++
	return n
}

//go:noinline
func (a *arena[T]) resize() {
	if a.a != nil {
		a.len += a.len >> 1 // 1.5x growth, integer math
	}
	// First alloc: keep the initial a.len value (set by newArena).
	a.a = unsafe.Pointer(&make([]T, a.len)[0])
	a.index = 0
}

// makeSlice allocates n contiguous elements from the arena and returns a
// slice whose backing array lives in arena memory. If the current chunk
// doesn't have enough room, a new chunk is allocated that is large enough.
func (a *arena[T]) makeSlice(n int) []T {
	if n == 0 {
		return nil
	}
	un := uintptr(n)
	if a.index+un > a.len {
		a.growForSlice(un)
	}
	start := unsafe.Add(a.a, a.index*a.elementSize)
	a.index += un
	return unsafe.Slice((*T)(start), n)
}

// growForSlice allocates a new chunk large enough to hold at least minElems
// contiguous elements. Adds 50% headroom when the slice exceeds the grown
// chunk so subsequent small allocs don't immediately trigger another resize.
// Kept out-of-line so the fast path in makeSlice has no write barriers or
// float conversions.
//
//go:noinline
func (a *arena[T]) growForSlice(minElems uintptr) {
	var newLen uintptr
	if a.a != nil {
		newLen = a.len + a.len>>1 // 1.5x growth, integer math
	} else {
		newLen = a.len // first alloc: use the initial size
	}
	if newLen < minElems {
		newLen = minElems + minElems>>1 // headroom for future allocs
	}
	a.len = newLen
	a.a = unsafe.Pointer(&make([]T, newLen)[0])
	a.index = 0
}
