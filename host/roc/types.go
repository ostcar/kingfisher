package roc

/*
#cgo LDFLAGS: -L.. -lapp
#include "./host.h"
*/
import "C"
import (
	"unsafe"
)

func rocCallUpdateModel(events RocList[RocStr], model MaybeModel) ResultModel {
	var result ResultModel
	C.roc__mainForHost_2_caller(events.CPtr(), model.CPtr(), nil, result.CPtr())
	return result
}

func rocCallRespond(request RocRequest, model *unsafe.Pointer) ResultResponse {
	size := C.roc__mainForHost_0_result_size()
	capturePtr := roc_alloc(size, 0)
	defer roc_dealloc(capturePtr, 0)

	C.roc__mainForHost_0_caller(request.CPtr(), model, nil, capturePtr)

	var result ResultResponse
	C.roc__mainForHost_1_caller(nil, capturePtr, result.CPtr())
	return result
}

type ResultResponse C.struct_ResultResponse

func (r *ResultResponse) CPtr() *C.struct_ResultResponse {
	return (*C.struct_ResultResponse)(r)
}

func (r ResultResponse) result() RocResponse {
	switch r.disciminant {
	case 0: // Ok
		return (*(*RocResponse)(unsafe.Pointer(&r.payload)))
	default:
		panic("invalid disciminat")
	}
}

type MaybeModel C.struct_MaybeModel

func MaybeModelInit() MaybeModel {
	return MaybeModel{disciminant: 1}
}

func MaybeModelExisting(model unsafe.Pointer) MaybeModel {
	return MaybeModel{
		disciminant: 0,
		payload:     *(*[8]byte)(unsafe.Pointer(&model)),
	}
}

func (m *MaybeModel) CPtr() *C.struct_MaybeModel {
	return (*C.struct_MaybeModel)(m)
}

type ResultModel C.struct_ResultModel

func (r *ResultModel) CPtr() *C.struct_ResultModel {
	return (*C.struct_ResultModel)(r)
}

func (r ResultModel) result() (unsafe.Pointer, RocStr, bool) {
	switch r.disciminant {
	case 1: // Ok
		return *(*unsafe.Pointer)(unsafe.Pointer(&r.payload)), RocStr{}, true
	case 0: // Err
		return nil, (*(*RocStr)(unsafe.Pointer(&r.payload))), false
	default:
		panic("invalid disciminat")
	}
}

type RocResponse C.struct_Response

func (r *RocResponse) CPtr() *C.struct_Response {
	return (*C.struct_Response)(r)
}

func (r RocResponse) Headers() RocList[RocHeader] {
	ctypeList := RocList[C.struct_Header](r.headers)
	return *(*RocList[RocHeader])(unsafe.Pointer(&ctypeList))
}

func (r RocResponse) DecRef() {
	RocStr(r.body).DecRef()
	RocList[RocHeader](r.headers).DecRef()
}

type RocHeader C.struct_Header

func (r RocHeader) C() C.struct_Header {
	return C.struct_Header(r)
}

func (r RocHeader) DecRef() {
	RocStr(r.name).DecRef()
	RocList[byte](r.value).DecRef()
}

type RocRequest C.struct_Request

func (r *RocRequest) CPtr() *C.struct_Request {
	return (*C.struct_Request)(r)
}

type RocRequestTimeout C.struct_RequestTimeout

func requestTimeoutTimeoutMilliseconds(v int) RocRequestTimeout {
	return RocRequestTimeout{
		discriminant: 1,
		payload:      *(*[8]byte)(unsafe.Pointer(&v)),
	}
}

func requestTimeoutNoTimeout() RocRequestTimeout {
	return RocRequestTimeout{
		discriminant: 0,
	}
}

func (r RocRequestTimeout) C() C.struct_RequestTimeout {
	return C.struct_RequestTimeout(r)
}

type RocResultVoidVoid C.struct_ResultVoidVoid

type RocResultVoidString C.struct_ResultVoidStr
