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
	decodeArg := decodeArgInit()
	if encodedModel != nil {
		decodeArg = decodeArgExisting(NewRocList(encodedModel))
	}

	var mayModel ResultModel
	C.roc__mainForHost_0_caller(decodeArg.CPtr(), nil, mayModel.CPtr())

	var model unsafe.Pointer
	model, errStr, ok := mayModel.result()
	if !ok {
		return nil, fmt.Errorf("decoding model: Roc returned: %s", errStr)
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

	setRefCountToInfinity(r.model)

	return &r, nil
}

// DumpModel returns a []byte reprsentation of the model.
func (r *Roc) DumpModel() []byte {
	var rocBytes RocList[byte]
	C.roc__mainForHost_1_caller(&r.model, nil, rocBytes.CPtr())
	return RocList[byte](rocBytes).List()
}

func setRefCountToInfinity(ptr unsafe.Pointer) {
	// Setting the refcount to 0 tells roc, not to modify it.
	refcountPtr := unsafe.Add(ptr, -8)
	refCountSlice := unsafe.Slice((*uint)(refcountPtr), 1)
	refCountSlice[0] = 0
}

func setRefCountToOne(ptr unsafe.Pointer) {
	refcountPtr := unsafe.Add(ptr, -8)
	refCountSlice := unsafe.Slice((*uint)(refcountPtr), 1)
	refCountSlice[0] = 9223372036854775808
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
	var response RocResponse
	C.roc__mainForHost_2_caller(rocRequest.CPtr(), &r.model, nil, response.CPtr())

	return Response{
		Status:  int(response.status),
		Headers: toGoHeaders(response.Headers()),
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
		Headers: toGoHeaders(responseModel.Response().Headers()),
		Body:    RocStr(responseModel.response.body).String(),
	}

	r.model = unsafe.Pointer(responseModel.model)
	setRefCountToInfinity(r.model)

	return response, nil
}

func (r *Roc) runWriteRequest(request Request) (RocResponseModel, error) {
	rocRequest, err := convertRequest(request)
	if err != nil {
		return RocResponseModel{}, fmt.Errorf("convert request: %w", err)
	}

	var responseModel RocResponseModel
	setRefCountToOne(r.model)
	C.roc__mainForHost_3_caller(rocRequest.CPtr(), &r.model, nil, responseModel.CPtr())
	return responseModel, nil
}

func convertRequest(r Request) (RocRequest, error) {
	contentType := r.getHeader("Content-type")
	if contentType == "" {
		contentType = "text/plain"
	}

	var rocRequest RocRequest
	rocRequest.body = NewRocList(r.Body).C()
	rocRequest.mimeType = NewRocStr(contentType).C()
	rocRequest.methodEnum = C.uchar(convertMethod(r.Method))
	rocRequest.headers = toRocHeader(r.Header).C()
	rocRequest.url = NewRocStr(r.URL).C()
	// TODO: What is a request timeout?
	rocRequest.timeout = requestTimeoutNoTimeout().C()
	return rocRequest, nil
}

func toRocHeader(goHeader map[string][]string) RocList[RocHeader] {
	_ = C.struct_Header{} // TODO: This seems to be a go bug. If This line is removed, go generates a wrong c type.

	// This is only the correct len, if each header-name unique. This should be most of the time.
	headers := make([]RocHeader, 0, len(goHeader))
	for name, valueList := range goHeader {
		for _, value := range valueList {
			headers = append(headers, RocHeader{name: NewRocStr(name).C(), value: NewRocList([]byte(value)).C()})
		}
	}

	return NewRocList(headers)
}

func toGoHeaders(rocHeaders RocList[RocHeader]) []Header {
	headerList := rocHeaders.List()

	goHeader := make([]Header, len(headerList))
	for i, header := range headerList {
		goHeader[i] = Header{Name: RocStr(header.name).String(), Value: string(RocList[byte](header.value).List())}
	}

	return goHeader
}

func convertMethod(method string) byte {
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
