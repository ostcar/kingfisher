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

	rocList.len = C.ulong(len(list))
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

type Freer interface {
	Free()
}

func (r RocList[t]) Free() {
	ptr := unsafe.Pointer(r.bytes)
	if ptr == nil {
		return
	}

	for _, e := range r.List() {
		hasFree, ok := any(e).(Freer)
		if !ok {
			break
		}
		hasFree.Free()
	}

	freeForRoc(ptr)
}
