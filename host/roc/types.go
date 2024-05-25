package roc

import "C"
import (
	"unsafe"
)

type RocDecodeArg C.struct_DecodeArg

func decodeArgInit() RocDecodeArg {
	return RocDecodeArg{
		discriminant: 1,
	}
}

func decodeArgExisting(list RocList[byte]) RocDecodeArg {
	return RocDecodeArg{
		discriminant: 0,
		payload:      *(*[24]byte)(unsafe.Pointer(&list)),
	}
}

func (r RocDecodeArg) C() C.struct_DecodeArg {
	return C.struct_DecodeArg(r)
}

func (r *RocDecodeArg) CPtr() *C.struct_DecodeArg {
	return (*C.struct_DecodeArg)(r)
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

func (r RocResponse) Free() {
	RocStr(r.body).Free()
	RocList[RocHeader](r.headers).Free()
}

type RocHeader C.struct_Header

func (r RocHeader) C() C.struct_Header {
	return C.struct_Header(r)
}

func (r RocHeader) Free() {
	RocStr(r.name).Free()
	RocList[byte](r.value).Free()
}

type RocResponseModel C.struct_ResponseModel

func (r RocResponseModel) Response() RocResponse {
	return *(*RocResponse)(unsafe.Pointer(&r.response))
}

func (r *RocResponseModel) CPtr() *C.struct_ResponseModel {
	return (*C.struct_ResponseModel)(r)
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
