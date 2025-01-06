package roc

/*
#include "./roc_std.h"
*/
import "C"
import (
	"unsafe"
)

type RocList[t any] C.struct_RocList

func NewRocList[t any](list []t) RocList[t] {
	if len(list) == 0 {
		return RocList[t]{}
	}

	var rocList RocList[t]
	var zero t
	typeSize := int(unsafe.Sizeof(zero))
	neededBytes := len(list) * typeSize

	var ptr unsafe.Pointer
	if rocList.ElementsAreRefcounted() {
		originalLenPointer := roc_alloc(C.size_t(neededBytes)+2*intBytes, intBytes)
		ptr = unsafe.Add(originalLenPointer, 2*intBytes)
		setRefCountToOne(ptr)
	} else {
		// When elements are not refcounted, then we can use the normal of allocation
		ptr = allocForRoc(neededBytes)
	}

	rocList.len = C.size_t(len(list))
	rocList.capacity = rocList.len
	rocList.bytes = (*C.char)(unsafe.Pointer(ptr))

	dataSlice := unsafe.Slice((*t)(ptr), len(list))
	copy(dataSlice, list)

	return rocList
}

func (r RocList[t]) List() []t {
	ptr := (*t)(unsafe.Pointer(r.bytes))
	return unsafe.Slice(ptr, r.len)
}

func (r RocList[t]) C() C.struct_RocList {
	return C.struct_RocList(r)
}

func (r *RocList[t]) CPtr() *C.struct_RocList {
	return (*C.struct_RocList)(r)
}

func (r *RocList[t]) IsSeamless() bool {
	return int(r.capacity) < 0
}

func (r *RocList[t]) ElementsAreRefcounted() bool {
	var zero t
	_, elementsAreRefcounted := any(zero).(decRefer)
	return elementsAreRefcounted
}

func (r RocList[t]) DecRef() {
	ptr := unsafe.Pointer(r.bytes)
	if ptr == nil {
		return
	}

	pointerToStart := unsafe.Pointer(r.bytes)
	if r.IsSeamless() {
		pointerToStart = unsafe.Pointer(uintptr(r.capacity << 1))
	}
	pointerToRefCount := unsafe.Add(pointerToStart, -intBytes)
	refCountValue := *(*uint)(pointerToRefCount)

	if refCountValue > refcountOne {
		refCountValue -= 1
		return
	}

	if !r.ElementsAreRefcounted() {
		// If the elements are not refcounted, the memory can just be cleared.
		roc_dealloc(pointerToRefCount, 0)
		return
	}

	pointerToOriginalLen := unsafe.Add(pointerToRefCount, -intBytes)

	len := uint(r.len)
	if r.IsSeamless() {
		len = *(*uint)(pointerToOriginalLen)
	}

	elements := unsafe.Slice((*t)(pointerToStart), len)

	for _, e := range elements {
		any(e).(decRefer).DecRef()
	}

	// TODO: This is not correct. It depends on alignment
	// This is the correct way to calculate the start of the allocated bytes. Same in in if conditoin above:
	// const required_space: usize = if (elements_refcounted) (2 * ptr_width) else ptr_width;
	// const extra_bytes = @max(required_space, element_alignment);
	roc_dealloc(pointerToOriginalLen, 0)
}

type decRefer interface {
	DecRef()
}
