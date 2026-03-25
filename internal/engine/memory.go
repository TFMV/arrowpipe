package engine

import (
    "unsafe"

    "github.com/apache/arrow-go/v18/arrow/memory"
)

// AlignedAllocator is a memory allocator that ensures memory is aligned to a specific boundary.
// This is crucial for performance with certain hardware instructions (e.g., SIMD).
type AlignedAllocator struct {
    pool memory.Allocator
    alignment int
}

// NewAlignedAllocator creates a new AlignedAllocator.
func NewAlignedAllocator(pool memory.Allocator, alignment int) *AlignedAllocator {
    return &AlignedAllocator{pool, alignment}
}

// Allocate allocates a slice of memory with the specified alignment.
func (a *AlignedAllocator) Allocate(size int) []byte {
    buf := a.pool.Allocate(size + a.alignment)
    offset := int(uintptr(unsafe.Pointer(&buf[0])) & uintptr(a.alignment-1))
    return buf[offset : offset+size]
}

// Reallocate reallocates a slice of memory.
func (a *AlignedAllocator) Reallocate(size int, b []byte) []byte {
    return a.pool.Reallocate(size, b)
}

// Free frees a slice of memory.
func (a *AlignedAllocator) Free(b []byte) {
    a.pool.Free(b)
}
