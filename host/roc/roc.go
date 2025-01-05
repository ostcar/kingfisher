package roc

import "C"

import (
	"bytes"
	"fmt"
	"io"
	"iter"
	"net/http"
	"sync"
	"unsafe"
)

// Roc holds the connection to roc.
type Roc struct {
	mu sync.RWMutex

	model unsafe.Pointer
}

// New initializes the connection to roc.
func New(eventReader iter.Seq2[[]byte, error]) (*Roc, error) {
	initModel := rocCallInitModel()

	var events []RocList[byte]
	for event, err := range eventReader {
		if err != nil {
			return nil, fmt.Errorf("can not read event: %w", err)
		}

		events = append(events, NewRocList(event))
	}

	rocEvents := NewRocList(events)

	model, err, success := rocCallUpdateModel(initModel, rocEvents).result()
	if !success {
		return nil, fmt.Errorf("can not run events: %v", err)
	}

	setRefCountToInfinity(model)

	r := Roc{model: model}

	return &r, nil
}

func (r *Roc) HanldeRequest(w http.ResponseWriter, req *http.Request, eventWriter func(event ...[]byte) error) error {
	rocRequest, err := convertRequest(req)
	if err != nil {
		return fmt.Errorf("convert request: %w", err)
	}

	var rocResponse RocResponse
	if isWriteRequest(req.Method) {
		rocResponse, err = r.handleWriteRequest(rocRequest, eventWriter)
	} else {
		rocResponse, err = r.handleReadRequest(rocRequest)
	}
	if err != nil {
		return fmt.Errorf("handle request returned %v", err)
	}
	defer rocResponse.DecRef()

	writeResponse(w, rocResponse)

	return nil
}

func isWriteRequest(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func (r *Roc) handleWriteRequest(rocRequest RocRequest, eventWriter func(event ...[]byte) error) (RocResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	currentEvents = [][]byte{}
	response, rocErr, success := rocCallHandleRequest(rocRequest, r.model).result()
	if !success {
		return RocResponse{}, fmt.Errorf("calling Roc for write: %v", rocErr)
	}

	if len(currentEvents) > 0 {
		eventSlice := make([]RocList[byte], len(currentEvents))
		for i, event := range currentEvents {
			eventSlice[i] = NewRocList(event)
		}

		var events = NewRocList(eventSlice)

		newModel, failMsg, success := rocCallUpdateModel(r.model, events).result()
		if !success {
			return RocResponse{}, fmt.Errorf("update model: %s", failMsg)
		}

		setRefCountToInfinity(newModel)

		if err := eventWriter(currentEvents...); err != nil {
			return RocResponse{}, fmt.Errorf("save event: %w", err)
		}

		r.model = newModel
	}

	return response, nil
}

func (r *Roc) handleReadRequest(rocRequest RocRequest) (RocResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	response, rocErr, success := rocCallHandleRequest(rocRequest, r.model).result()
	if !success {
		return RocResponse{}, fmt.Errorf("calling Roc for read: %v", rocErr)
	}

	return response, nil
}

func writeResponse(w http.ResponseWriter, r RocResponse) {
	for _, header := range r.Headers().List() {
		w.Header().Add(RocStr(header.name).String(), RocStr(header.value).String())
	}

	w.WriteHeader(int(r.status))
	w.Write(bytes.Clone((RocList[byte](r.body).List())))
}

func convertRequest(r *http.Request) (RocRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return RocRequest{}, fmt.Errorf("read request body: %w", err)
	}

	var rocRequest RocRequest
	rocRequest.body = NewRocList(body).C()
	rocRequest.methodEnum = C.uchar(convertMethod(r.Method))
	rocRequest.headers = toRocHeader(r.Header).C()
	rocRequest.url = NewRocStr(r.URL.String()).C()
	return rocRequest, nil
}

func toRocHeader(goHeader map[string][]string) RocList[RocHeader] {
	// This is only the correct len, if each header-name unique. This should be most of the time.
	headers := make([]RocHeader, 0, len(goHeader))
	for name, valueList := range goHeader {
		for _, value := range valueList {
			headers = append(headers, RocHeader{name: NewRocStr(name).C(), value: NewRocStr(value).C()})
		}
	}

	return NewRocList(headers)
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

var currentEvents [][]byte // This has to be global as long as effects don't get a hidden argument.

//export roc_fx_save_event
func roc_fx_save_event(event *C.struct_RocList) {
	// TODO: Do I have to deref the value? Should I make a copy?
	buf := make([]byte, event.len)
	copy(buf, RocList[byte](*event).List())

	currentEvents = append(currentEvents, buf)
}

// //export roc_fx_stderrLine
// func roc_fx_stderrLine(msg *RocStr) RocResultVoidString {
// 	// TODO: use stderr
// 	fmt.Println(*msg)
// 	return RocResultVoidString{
// 		disciminant: 1,
// 	}
// }
