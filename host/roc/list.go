package roc

import "C"
import (
	"unsafe"
)

type RocList[t any] C.struct_RocList

func NewRocList[t any](list []t) RocList[t] {
	var zero t
	typeSize := int(unsafe.Sizeof(zero))

	var rocList RocList[t]
	if len(list) > 0 {
		// TODO: 8 only works for 64bit. Use the correct size.
		refCountPtr := roc_alloc(C.ulong(len(list)*typeSize+8), 8)
		refCountSlice := unsafe.Slice((*uint)(refCountPtr), 1)
		refCountSlice[0] = refcount_one
		startPtr := unsafe.Add(refCountPtr, 8)

		rocList.len = C.ulong(len(list))
		rocList.capacity = rocList.len
		rocList.bytes = (*C.char)(unsafe.Pointer(startPtr))

		dataSlice := unsafe.Slice((*t)(startPtr), len(list))
		copy(dataSlice, list)
	}

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

	// TODO Fix for non 64 systems
	refCountPtr := unsafe.Add(ptr, -8)
	roc_dealloc(refCountPtr, 0)
}
