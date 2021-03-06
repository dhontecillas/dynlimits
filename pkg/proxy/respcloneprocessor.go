package proxy

import (
	"net/http"
)

type RecordedResponseProcessorFn func(rwr *ResponseWriterRecorder)

type RecordedResponseProcessorHandler interface {
	ProcessRecordedResponse(rwr *ResponseWriterRecorder)
}

type ParallelHandler struct {
	handler             http.Handler
	parallelProcRunning bool
	parallelProc        chan *ResponseWriterRecorder
	respProcessor       RecordedResponseProcessorHandler
}

func NewParallelHandler(h http.Handler,
	recRespProcessor RecordedResponseProcessorHandler) (p *ParallelHandler) {

	return &ParallelHandler{
		handler:       h,
		respProcessor: recRespProcessor,
	}
}

func (p *ParallelHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	recRW := NewResponseWriterRecorder(req, p.parallelProc)
	dupRW := NewDupResponseWriter(rw, recRW)
	p.handler.ServeHTTP(dupRW, req)
	// send the recorded request / response to be processed
	p.parallelProc <- recRW
}

func (p *ParallelHandler) LaunchParallelProc() {
	if p.parallelProc != nil {
		return
	}

	p.parallelProc = make(chan *ResponseWriterRecorder)
	go func() {
		for {
			select {
			case r := <-p.parallelProc:
				p.respProcessor.ProcessRecordedResponse(r)
			}
		}
	}()
}
