package roc

/*
#cgo LDFLAGS: -L.. -lapp
#include "./app.h"
*/
import "C"
import (
	"unsafe"
)

func rocCallInitModel() unsafe.Pointer {
	var model unsafe.Pointer
	C.roc__init_model_for_host_1_exposed_generic(&model)
	return model
}

func rocCallUpdateModel(model unsafe.Pointer, events RocList[RocList[byte]]) ResultModel {
	var newModel ResultModel
	C.roc__update_model_for_host_1_exposed_generic(newModel.CPtr(), model, events.CPtr())
	return newModel
}

func rocCallHandleRequest(request RocRequest, model unsafe.Pointer) ResultResponse {
	var result ResultResponse
	C.roc__handle_request_for_host_1_exposed_generic(result.CPtr(), request.CPtr(), model)
	return result
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

type ResultResponse C.struct_ResultResponse

func (r *ResultResponse) CPtr() *C.struct_ResultResponse {
	return (*C.struct_ResultResponse)(r)
}

func (r ResultResponse) result() (RocResponse, RocStr, bool) {
	switch r.disciminant {
	case 1: // Ok
		return *(*RocResponse)(unsafe.Pointer(&r.payload)), RocStr{}, true
	case 0: // Err
		return RocResponse{}, (*(*RocStr)(unsafe.Pointer(&r.payload))), false
	default:
		panic("invalid disciminat")
	}
}

type RocResponse C.struct_Response

func (r RocResponse) Headers() RocList[RocHeader] {
	ctypeList := RocList[C.struct_Header](r.headers)
	return *(*RocList[RocHeader])(unsafe.Pointer(&ctypeList))
}

func (r RocResponse) DecRef() {
	RocStr(r.body).DecRef()
	RocList[RocHeader](r.headers).DecRef()
}

type RocHeader C.struct_Header

func (r RocHeader) DecRef() {
	RocStr(r.name).DecRef()
	RocStr(r.value).DecRef()
}

type RocRequest C.struct_Request

func (r *RocRequest) CPtr() *C.struct_Request {
	return (*C.struct_Request)(r)
}
