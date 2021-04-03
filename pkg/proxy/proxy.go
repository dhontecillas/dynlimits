package proxy

import (
	"bytes"
	"fmt"
	"net/http"
)

type ProxyHandler struct {
	scheme      string
	forwardAddr string
	client      *http.Client
}

func NewProxyHandler(scheme string, forwardAddr string) *ProxyHandler {
	// TODO: check that scheme is http or https
	return &ProxyHandler{
		scheme:      scheme,
		forwardAddr: forwardAddr,
		client:      &http.Client{},
	}
}

func (ph *ProxyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// just write the success header for now
	// c := context.Background()
	nr, err := http.NewRequestWithContext(req.Context(),
		req.Method, req.URL.String(), req.Body)
	if err != nil {
		// TODO: log the error
		fmt.Printf("error creating request: %s\n", err.Error())
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	// TODO: check the error
	nr.Host = ph.forwardAddr
	nr.URL.Scheme = ph.scheme
	nr.URL.Host = ph.forwardAddr
	nr.ContentLength = req.ContentLength

	for key, slc := range req.Header {
		nr.Header[key] = make([]string, len(slc))
		copy(nr.Header[key], slc)
	}

	res, err := ph.client.Do(nr)
	if err != nil {
		// TODO: log the error
		// fmt.Printf("error proxying the request: %s\n", err.Error())
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	dstH := rw.Header()
	for key, slc := range res.Header {
		dstH[key] = make([]string, len(slc))
		copy(dstH[key], slc)
	}
	rw.WriteHeader(res.StatusCode)

	var buf []byte
	if res.ContentLength >= 0 {
		buf = make([]byte, 0, res.ContentLength)
	}
	b := bytes.NewBuffer(buf)
	b.ReadFrom(res.Body)
	rw.Write(b.Bytes())
}
