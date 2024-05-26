package roc

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

	ptr := allocForRoc(len(list) * typeSize)

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

func (r RocList[t]) DecRef() {
	ptr := unsafe.Pointer(r.bytes)
	if ptr == nil {
		return
	}

	type decRefer interface {
		DecRef()
	}

	for _, e := range r.List() {
		hasDecRef, ok := any(e).(decRefer)
		if !ok {
			break
		}
		hasDecRef.DecRef()
	}

	decRefCount(ptr)
}
