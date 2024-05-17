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

		refCountPtr := roc_alloc(C.ulong(len(encodedModel)+8), 1)
		refCountSlice := unsafe.Slice((*uint)(refCountPtr), 1)
		refCountSlice[0] = 9223372036854775808
		startPtr := unsafe.Add(refCountPtr, 8)

		dataSlice := unsafe.Slice((*byte)(startPtr), len(encodedModel))
		copy(dataSlice, encodedModel)

		var rocList C.struct_RocList
		rocList.len = C.ulong(len(encodedModel))
		rocList.capacity = rocList.len
		rocList.bytes = (*C.char)(unsafe.Pointer(startPtr))
		decodeArg.payload = *(*[24]byte)(unsafe.Pointer(&rocList))
	}

	var mayModel C.struct_ResultModel
	C.roc__mainForHost_0_caller(&decodeArg, nil, &mayModel)

	var model unsafe.Pointer
	switch mayModel.disciminant {
	case 1: // Ok
		model = *(*unsafe.Pointer)(unsafe.Pointer(&mayModel.payload))
	case 0: // Err
		msg := rocStrRead(*(*C.struct_RocStr)(unsafe.Pointer(&mayModel.payload)))
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

	len := rocBytes.len
	ptr := (*byte)(unsafe.Pointer(rocBytes.bytes))
	bytes := unsafe.Slice(ptr, len)
	return bytes
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
		Headers: toGoHeaders(response.headers),
		Body:    rocStrRead(response.body),
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
		Headers: toGoHeaders(responseModel.response.headers),
		Body:    rocStrRead(responseModel.response.body),
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
	var requestBody C.struct_RequestBody
	requestBody.discriminant = 1
	if len(r.Body) > 0 {
		contentType := r.getHeader("Content-type")
		if contentType == "" {
			contentType = "text/plain"
		}

		requestBody.discriminant = 0
		var bodyMimetype C.struct_BodyMimeType
		bodyMimetype.mimeType = rocStrFromStr(contentType)
		bodyMimetype.body = rocStrFromStr(string(r.Body))
		requestBody.payload = *(*[48]byte)(unsafe.Pointer(&bodyMimetype))
	}

	var rocRequest C.struct_Request
	rocRequest.body = requestBody
	rocRequest.methodEnum = convertMethod(r.Method)
	rocRequest.headers = toRocHeader(r.Header)
	rocRequest.url = rocStrFromStr(r.URL)
	// TODO: What is a request timeout?
	rocRequest.timeout = C.struct_RequestTimeout{discriminant: 0}
	return rocRequest, nil
}

func toRocHeader(goHeader map[string][]string) C.struct_RocList {
	// This is only the correct len, if each header-name unique. This should be most of the time.
	headers := make([]C.struct_Header, 0, len(goHeader))
	for name, valueList := range goHeader {
		for _, value := range valueList {
			h := C.struct_Header{
				name:  rocStrFromStr(name),
				value: rocStrFromStr(value),
			}
			headers = append(headers, h)
		}
	}

	var header C.struct_Header
	elementSize := int(unsafe.Sizeof(header))
	fullSize := elementSize*len(headers) + 8

	refCountPtr := roc_alloc(C.ulong(fullSize), 8)
	refCountSlice := unsafe.Slice((*uint)(refCountPtr), 1)
	refCountSlice[0] = 9223372036854775808
	startPtr := unsafe.Add(refCountPtr, 8)

	rocStrList := make([]C.struct_Header, len(headers))
	copy(rocStrList, headers)

	dataSlice := unsafe.Slice((*C.struct_Header)(startPtr), len(rocStrList))
	copy(dataSlice, rocStrList)

	var rocList C.struct_RocList
	rocList.len = C.ulong(len(headers))
	rocList.capacity = rocList.len
	rocList.bytes = (*C.char)(unsafe.Pointer(startPtr))

	return rocList
}

func toGoHeaders(rocHeaders C.struct_RocList) []Header {
	len := rocHeaders.len
	ptr := (*C.struct_Header)(unsafe.Pointer(rocHeaders.bytes))
	headerList := unsafe.Slice(ptr, len)

	goHeader := make([]Header, len)
	for i, header := range headerList {
		goHeader[i] = Header{Name: rocStrRead(header.name), Value: rocStrRead(header.value)}
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

func rocListBytes(rocList C.struct_RocList) []byte {
	len := rocList.len
	ptr := (*byte)(unsafe.Pointer(rocList.bytes))
	return unsafe.Slice(ptr, len)
}

func rocStrFromStr(str string) C.struct_RocStr {
	// TODO: 8 only works for 64bit. Use the correct size.
	refCountPtr := roc_alloc(C.ulong(len(str)+8), 8)
	refCountSlice := unsafe.Slice((*uint)(refCountPtr), 1)
	refCountSlice[0] = 9223372036854775808 // TODO: calculate this number from the lowest int
	startPtr := unsafe.Add(refCountPtr, 8)

	var rocStr C.struct_RocStr
	rocStr.len = C.ulong(len(str))
	rocStr.capacity = rocStr.len
	rocStr.bytes = (*C.char)(unsafe.Pointer(startPtr))

	dataSlice := unsafe.Slice((*byte)(startPtr), len(str))
	copy(dataSlice, []byte(str))

	return rocStr
}

func rocStrRead(rocStr C.struct_RocStr) string {
	if int(rocStr.capacity) < 0 {
		// Small string
		ptr := (*byte)(unsafe.Pointer(&rocStr))

		byteLen := 12
		if is64Bit {
			byteLen = 24
		}

		shortStr := unsafe.String(ptr, byteLen)
		len := shortStr[byteLen-1] ^ 128
		return shortStr[:len]
	}

	// Remove the bit for seamless string
	len := (uint(rocStr.len) << 1) >> 1
	ptr := (*byte)(unsafe.Pointer(rocStr.bytes))
	return unsafe.String(ptr, len)
}

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
func roc_panic(msg *C.struct_RocStr, tagID C.uint) {
	panic(fmt.Sprint(rocStrRead(*msg)))
}

//export roc_dbg
func roc_dbg(loc *C.struct_RocStr, msg *C.struct_RocStr, src *C.struct_RocStr) {
	locStr := rocStrRead(*loc)
	msgStr := rocStrRead(*msg)
	srcStr := rocStrRead(*src)

	if srcStr == msgStr {
		fmt.Fprintf(os.Stderr, "[%s] {%s}\n", locStr, msgStr)
	} else {
		fmt.Fprintf(os.Stderr, "[%s] {%s} = {%s}\n", locStr, srcStr, msgStr)
	}
}
