package roc

import "C"
import "unsafe"

type RocStr C.struct_RocStr

func NewRocStr(str string) RocStr {
	// TODO: 8 only works for 64bit. Use the correct size.
	refCountPtr := roc_alloc(C.ulong(len(str)+8), 8)
	refCountSlice := unsafe.Slice((*uint)(refCountPtr), 1)
	refCountSlice[0] = 9223372036854775808 // TODO: calculate this number from the lowest int
	startPtr := unsafe.Add(refCountPtr, 8)

	var rocStr RocStr
	rocStr.len = C.ulong(len(str))
	rocStr.capacity = rocStr.len
	rocStr.bytes = (*C.char)(unsafe.Pointer(startPtr))

	dataSlice := unsafe.Slice((*byte)(startPtr), len(str))
	copy(dataSlice, []byte(str))

	return rocStr
}

func (r RocStr) String() string {
	if int(r.capacity) < 0 {
		// Small string
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
