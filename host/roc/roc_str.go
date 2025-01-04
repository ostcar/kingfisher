package roc

/*
#include "./roc_std.h"
*/
import "C"
import (
	"unsafe"
)

type RocStr C.struct_RocStr

func NewRocStr(str string) RocStr {
	ptr := allocForRoc(len(str))

	var rocStr RocStr
	rocStr.len = C.size_t(len(str))
	rocStr.capacity = rocStr.len
	rocStr.bytes = (*C.char)(unsafe.Pointer(ptr))

	dataSlice := unsafe.Slice((*byte)(ptr), len(str))
	copy(dataSlice, []byte(str))

	return rocStr
}

func (r RocStr) Small() bool {
	return int(r.capacity) < 0
}

func (r RocStr) String() string {
	if r.Small() {
		ptr := (*byte)(unsafe.Pointer(&r))

		byteLen := 12
		if is64Bit {
			byteLen = 24
		}

		shortStr := unsafe.String(ptr, byteLen)
		len := shortStr[byteLen-1] ^ 128
		return shortStr[:len]
	}

	// Remove the bit for seamless string
	len := (uint(r.len) << 1) >> 1
	ptr := (*byte)(unsafe.Pointer(r.bytes))
	return unsafe.String(ptr, len)
}

func (r RocStr) C() C.struct_RocStr {
	return C.struct_RocStr(r)
}

func (r *RocStr) CPtr() *C.struct_RocStr {
	return (*C.struct_RocStr)(r)
}

func (r RocStr) DecRef() {
	ptr := unsafe.Pointer(r.bytes)
	if r.Small() || ptr == nil {
		return
	}

	decRefCount(ptr)
}
