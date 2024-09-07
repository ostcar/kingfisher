package roc

import "C"

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strings"
	"sync"
	"unsafe"
)

var currentEvents []RocStr // This has to be global as long as effects don't get a hidden argument.

// Roc holds the connection to roc.
type Roc struct {
	mu    sync.RWMutex
	model unsafe.Pointer
}

// New initializes the connection to roc.
func New(eventReader io.Reader) (*Roc, error) {
	buf := bufio.NewReader(eventReader)

	var events []RocStr
	for {
		str, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("reading from eventReader: %w", err)
		}

		events = append(events, NewRocStr(str[0:len(str)-1]))
	}

	rocEvents := NewRocList(events)

	model, err, success := rocCallUpdateModel(rocEvents, MaybeModelInit()).result()
	if !success {
		return nil, fmt.Errorf("can not update model: %s", err.String())
	}

	r := Roc{model: model}

	setRefCountToInfinity(r.model)

	return &r, nil
}

//export roc_fx_saveEvent
func roc_fx_saveEvent(event *RocStr) RocResultVoidString {
	currentEvents = append(currentEvents, *event)
	return RocResultVoidString{
		disciminant: 1,
	}
}

//export roc_fx_stderrLine
func roc_fx_stderrLine(msg *RocStr) RocResultVoidString {
	// TODO: use stderr
	fmt.Println(*msg)
	return RocResultVoidString{
		disciminant: 1,
	}
}

// Request represents an http request
type Request struct {
	Body    []byte              `json:"body"`
	Method  string              `json:"method"`
	Header  map[string][]string `json:"headers"`
	URL     string              `json:"url"`
	Timeout uint64              `json:"timeout"`
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

// TODO: can we use http.Request and http.Response now?
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

	response := rocCallRespond(rocRequest, &r.model).result()
	defer response.DecRef()

	return Response{
		Status:  int(response.status),
		Headers: toGoHeaders(response.Headers()),
		Body:    strings.Clone(RocStr(response.body).String()),
	}, nil
}

// WriteRequest handles a write request.
// TODO: this is nearly the same as ReadRequest
func (r *Roc) WriteRequest(request Request, db io.Writer) (Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rocRequest, err := convertRequest(request)
	if err != nil {
		return Response{}, fmt.Errorf("convert request: %w", err)
	}
	setRefCountToInfinity(unsafe.Pointer(&rocRequest.body.bytes))

	currentEvents = []RocStr{}
	response := rocCallRespond(rocRequest, &r.model).result()
	defer response.DecRef()

	var events = NewRocList(currentEvents)
	var existingModel = MaybeModelExisting(r.model)

	newModel, failMsg, success := rocCallUpdateModel(events, existingModel).result()
	if !success {
		return Response{}, fmt.Errorf("got invalid model: %s", failMsg)
	}

	for _, event := range currentEvents {
		fmt.Fprintln(db, event.String())
	}

	r.model = newModel

	return Response{
		Status:  int(response.status),
		Headers: toGoHeaders(response.Headers()),
		Body:    strings.Clone(string(RocList[byte](response.body).List())),
	}, nil

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
	// This is only the correct len, if each header-name unique. This should be most of the time.
	headers := make([]RocHeader, 0, len(goHeader))
	for name, valueList := range goHeader {
		for _, value := range valueList {
			headers = append(headers, RocHeader{name: NewRocStr(name).C(), value: NewRocStr(value).C()})
		}
	}

	return NewRocList(headers)
}

func toGoHeaders(rocHeaders RocList[RocHeader]) []Header {
	headerList := rocHeaders.List()

	goHeader := make([]Header, len(headerList))
	for i, header := range headerList {
		goHeader[i] = Header{
			Name:  strings.Clone(RocStr(header.name).String()),
			Value: strings.Clone(string(RocList[byte](header.value).List())),
		}
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
