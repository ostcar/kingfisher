package roc

/*
#cgo LDFLAGS: -L.. -lapp
#include "./host.h"
*/
import "C"

import (
	"fmt"
	"os"
	"unsafe"
)

const refcount_one = 1 << 63
const is64Bit = uint64(^uintptr(0)) == ^uint64(0)
const intSize = 32 << (^uint(0) >> 63)
const intBytes = intSize / 8

// allocForRoc allocates memory. Prefixes that memory with a refcounter set to
// one.
func allocForRoc(size int) unsafe.Pointer {
	refCountPtr := roc_alloc(C.ulong(size)+intBytes, intBytes)
	refCountSlice := unsafe.Slice((*uint)(refCountPtr), 1)
	refCountSlice[0] = refcount_one
	return unsafe.Add(refCountPtr, intBytes)
}

// freeForRoc frees the memory with its refcounter.
func freeForRoc(ptr unsafe.Pointer) {
	refcountPtr := unsafe.Add(ptr, -intBytes)
	roc_dealloc(refcountPtr, 0)
}

// decRefCount reduces the refcounter by one.
//
// If the refcounter gets 0, the memory is freed.
func decRefCount(ptr unsafe.Pointer) {
	refcountPtr := unsafe.Add(ptr, -intBytes)
	refCountSlice := unsafe.Slice((*uint64)(refcountPtr), 1)

	switch refCountSlice[0] {
	case refcount_one:
		freeForRoc(ptr)
	case 0:
		// Data is static. Nothing to do
	default:
		refCountSlice[0] -= 1
	}
}

func setRefCountToInfinity(ptr unsafe.Pointer) {
	// Setting the refcount to 0 tells roc, not to modify it.
	refcountPtr := unsafe.Add(ptr, -8)
	refCountSlice := unsafe.Slice((*uint)(refcountPtr), 1)
	refCountSlice[0] = 0
}

func setRefCountToOne(ptr unsafe.Pointer) {
	refcountPtr := unsafe.Add(ptr, -8)
	refCountSlice := unsafe.Slice((*uint)(refcountPtr), 1)
	refCountSlice[0] = refcount_one
}

//export roc_alloc
func roc_alloc(size C.ulong, alignment int) unsafe.Pointer {
	_ = alignment
	return C.malloc(size)
}

//export roc_realloc
func roc_realloc(ptr unsafe.Pointer, newSize, _ C.ulong, alignment int) unsafe.Pointer {
	_ = alignment
	return C.realloc(ptr, newSize)
}

//export roc_dealloc
func roc_dealloc(ptr unsafe.Pointer, alignment int) {
	_ = alignment
	C.free(ptr)
}

//export roc_panic
func roc_panic(msg *RocStr, tagID C.uint) {
	panic(msg.String())
}

//export roc_dbg
func roc_dbg(loc *RocStr, msg *RocStr, src *RocStr) {
	if src.String() == msg.String() {
		fmt.Fprintf(os.Stderr, "[%s] {%s}\n", loc, msg)
	} else {
		fmt.Fprintf(os.Stderr, "[%s] {%s} = {%s}\n", loc, src, msg)
	}
}
