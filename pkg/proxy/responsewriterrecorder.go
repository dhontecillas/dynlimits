package proxy

import (
	"bytes"
	"net/http"
)

type ResponseWriterRecorder struct {
	Req        *http.Request // reference to the original request
	StatusCode int
	Headers    http.Header
	Data       *bytes.Buffer
}

func NewResponseWriterRecorder(req *http.Request,
	onWriteFinishedChan chan<- *ResponseWriterRecorder) *ResponseWriterRecorder {
	// clones the request, so it can fiddle with it in a separate goroutine
	// when main processing completes
	clonedReq := req.Clone(req.Context())
	return &ResponseWriterRecorder{
		Req:        clonedReq,
		StatusCode: 0,
		Headers:    make(http.Header),
		Data:       new(bytes.Buffer),
	}
}

func (rwr *ResponseWriterRecorder) Header() http.Header {
	return rwr.Headers
}

func (rwr *ResponseWriterRecorder) Write(data []byte) (int, error) {
	// WriteHeader only writes the header if it has not been
	// previously written
	rwr.WriteHeader(http.StatusOK)
	// TODO: perhaps we can use a bytes.Buffer ?
	if rwr.Data == nil {
		rwr.Data = bytes.NewBuffer(data)
	} else {
		rwr.Data.Write(data)
	}
	return len(data), nil
}

func (rwr *ResponseWriterRecorder) WriteHeader(statusCode int) {
	if rwr.StatusCode != 0 {
		return
	}
	rwr.StatusCode = statusCode
}
