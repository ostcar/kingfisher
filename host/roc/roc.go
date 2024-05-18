package roc

/*
#cgo LDFLAGS: -L.. -lapp
#include "./host.h"
*/
import "C"

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"os"
	"sync"
	"unsafe"
)

// Roc holds the connection to roc.
type Roc struct {
	mu sync.RWMutex

	model unsafe.Pointer
}

// New initializes the connection to roc.
func New(encodedModel []byte, reader io.Reader) (*Roc, error) {
	var decodeArg C.struct_DecodeArg
	decodeArg.discriminant = 1 // Init
	if encodedModel != nil {
		decodeArg.discriminant = 0 // Exist
		rocList := NewRocList(encodedModel)
		decodeArg.payload = *(*[24]byte)(unsafe.Pointer(&rocList))
	}

	var mayModel C.struct_ResultModel
	C.roc__mainForHost_0_caller(&decodeArg, nil, &mayModel)

	var model unsafe.Pointer
	switch mayModel.disciminant {
	case 1: // Ok
		model = *(*unsafe.Pointer)(unsafe.Pointer(&mayModel.payload))
	case 0: // Err
		msg := (*(*RocStr)(unsafe.Pointer(&mayModel.payload))).String()
		return nil, fmt.Errorf("decoding model: Roc returned: %v", msg)
	default:
		return nil, fmt.Errorf("decoding model got invalid data")
	}

	r := Roc{model: model}

	decoder := json.NewDecoder(reader)
	for {
		var request Request
		if err := decoder.Decode(&request); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decoding data: %w", err)
		}

		responseModel, err := r.runWriteRequest(request)
		if err != nil {
			return nil, fmt.Errorf("rerun request: %w", err)
		}
		r.model = unsafe.Pointer(responseModel.model)
	}

	return &r, nil
}

// DumpModel returns a []byte reprsentation of the model.
func (r *Roc) DumpModel() []byte {
	var rocBytes C.struct_RocList
	C.roc__mainForHost_1_caller(&r.model, nil, &rocBytes)
	return RocList[byte](rocBytes).List()
}

func setRefCountToTwo(ptr unsafe.Pointer) {
	refcountPtr := unsafe.Add(ptr, -8)
	refCountSlice := unsafe.Slice((*uint)(refcountPtr), 1)
	refCountSlice[0] = 9223372036854775809
}

// Request represents an http request
type Request struct {
	Body    []byte              `json:"body"`
	Method  string              `json:"method"`
	Header  map[string][]string `json:"headers"`
	URL     string              `json:"url"`
	Timeout uint64              `json:"timeout"`
}

// RequestFromHTTP creates a Request object from an http.Request.
func RequestFromHTTP(r *http.Request) (Request, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return Request{}, fmt.Errorf("read request body: %w", err)
	}

	return Request{
		Body:    body,
		Method:  r.Method,
		Header:  r.Header,
		URL:     r.URL.String(),
		Timeout: 0, // What is a request timeout?
	}, nil
}

func (r Request) getHeader(key string) string {
	return textproto.MIMEHeader(r.Header).Get(key)
}

// ReadRequest handles a read request.
func (r *Roc) ReadRequest(request Request) (Response, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rocRequest, err := convertRequest(request)
	if err != nil {
		return Response{}, fmt.Errorf("convert request: %w", err)
	}

	// TODO: check the refcount of the response and deallocate it if necessary.
	var response C.struct_Response
	setRefCountToTwo(r.model)
	C.roc__mainForHost_2_caller(&rocRequest, &r.model, nil, &response)

	return Response{
		Status:  int(response.status),
		Headers: toGoHeaders(RocList[C.struct_Header](response.headers)),
		Body:    RocStr(response.body).String(),
	}, nil
}

// WriteRequest handles a write request.
func (r *Roc) WriteRequest(request Request, db io.Writer) (Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	responseModel, err := r.runWriteRequest(request)
	if err != nil {
		return Response{}, err
	}

	if err := json.NewEncoder(db).Encode(request); err != nil {
		return Response{}, fmt.Errorf("encode request: %w", err)
	}

	response := Response{
		Status:  int(responseModel.response.status),
		Headers: toGoHeaders(RocList[C.struct_Header](responseModel.response.headers)),
		Body:    RocStr(responseModel.response.body).String(),
	}

	r.model = unsafe.Pointer(responseModel.model)

	return response, nil
}

func (r *Roc) runWriteRequest(request Request) (C.struct_ResponseModel, error) {
	rocRequest, err := convertRequest(request)
	if err != nil {
		return C.struct_ResponseModel{}, fmt.Errorf("convert request: %w", err)
	}

	// TODO: check the refcount of the response and deallocate it if necessary.
	var responseModel C.struct_ResponseModel
	setRefCountToTwo(r.model)
	C.roc__mainForHost_3_caller(&rocRequest, &r.model, nil, &responseModel)
	return responseModel, nil
}

func convertRequest(r Request) (C.struct_Request, error) {
	contentType := r.getHeader("Content-type")
	if contentType == "" {
		contentType = "text/plain"
	}

	var rocRequest C.struct_Request
	rocRequest.body = NewRocList(r.Body).C()
	rocRequest.mimeType = NewRocStr(contentType).C()
	rocRequest.methodEnum = convertMethod(r.Method)
	rocRequest.headers = toRocHeader(r.Header).C()
	rocRequest.url = NewRocStr(r.URL).C()
	// TODO: What is a request timeout?
	rocRequest.timeout = C.struct_RequestTimeout{discriminant: 0}
	return rocRequest, nil
}

func toRocHeader(goHeader map[string][]string) RocList[C.struct_Header] {
	// This is only the correct len, if each header-name unique. This should be most of the time.
	headers := make([]C.struct_Header, 0, len(goHeader))
	for name, valueList := range goHeader {
		for _, value := range valueList {
			h := C.struct_Header{
				name:  NewRocStr(name).C(),
				value: NewRocStr(value).C(),
			}
			headers = append(headers, h)
		}
	}

	return NewRocList(headers)
}

func toGoHeaders(rocHeaders RocList[C.struct_Header]) []Header {
	headerList := rocHeaders.List()

	goHeader := make([]Header, len(headerList))
	for i, header := range headerList {
		goHeader[i] = Header{Name: RocStr(header.name).String(), Value: RocStr(header.value).String()}
	}

	return goHeader
}

func convertMethod(method string) C.uchar {
	switch method {
	case http.MethodConnect:
		return 0
	case http.MethodDelete:
		return 1
	case http.MethodGet:
		return 2
	case http.MethodHead:
		return 3
	case http.MethodOptions:
		return 4
	case http.MethodPatch:
		return 5
	case http.MethodPost:
		return 6
	case http.MethodPut:
		return 7
	case http.MethodTrace:
		return 8
	default:
		panic("invalid method")
	}
}

// Header represents one http header.
type Header struct {
	Name  string
	Value string
}

// Response represents a http response.
type Response struct {
	Status  int
	Headers []Header
	Body    string
}

const is64Bit = uint64(^uintptr(0)) == ^uint64(0)

//export roc_alloc
func roc_alloc(size C.ulong, alignment int) unsafe.Pointer {
	_ = alignment
	return C.malloc(size)
}

//export roc_realloc
func roc_realloc(ptr unsafe.Pointer, newSize, _ C.ulong, alignment int) unsafe.Pointer {
	_ = alignment
	return C.realloc(ptr, newSize)
}

//export roc_dealloc
func roc_dealloc(ptr unsafe.Pointer, alignment int) {
	_ = alignment
	C.free(ptr)
}

//export roc_panic
func roc_panic(msg *RocStr, tagID C.uint) {
	panic(msg.String())
}

//export roc_dbg
func roc_dbg(loc *RocStr, msg *RocStr, src *RocStr) {
	if src.String() == msg.String() {
		fmt.Fprintf(os.Stderr, "[%s] {%s}\n", loc, msg)
	} else {
		fmt.Fprintf(os.Stderr, "[%s] {%s} = {%s}\n", loc, src, msg)
	}
}
