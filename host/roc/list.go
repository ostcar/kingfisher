package roc

import "C"
import "unsafe"

type RocList[t any] C.struct_RocList

func NewRocList[t any](list []t) RocList[t] {
	var zero t
	typeSize := int(unsafe.Sizeof(zero))

	// TODO: 8 only works for 64bit. Use the correct size.
	refCountPtr := roc_alloc(C.ulong(len(list)*typeSize+8), 8)
	refCountSlice := unsafe.Slice((*uint)(refCountPtr), 1)
	refCountSlice[0] = 9223372036854775808 // TODO: calculate this number from the lowest int
	startPtr := unsafe.Add(refCountPtr, 8)

	var rocList RocList[t]
	rocList.len = C.ulong(len(list))
	rocList.capacity = rocList.len
	rocList.bytes = (*C.char)(unsafe.Pointer(startPtr))

	dataSlice := unsafe.Slice((*t)(startPtr), len(list))
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
