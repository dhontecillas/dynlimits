package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"sync/atomic"
)

type Counters struct {
	startsAt int64
	seconds  [60 * 60 * 24]int64
}

var globalCounters Counters

func main() {
	fmt.Println("starting request counter...")

	cntHandler := NewCounterHandler()
	launchServer(cntHandler)
}

func launchServer(hfn http.Handler) {
	srv := &http.Server{
		// TODO: load this from config:
		Addr:    "0.0.0.0:9876",
		Handler: hfn,
	}

	sigChan := make(chan os.Signal, 1)

	go signalHandler(sigChan, srv)

	if err := srv.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			fmt.Printf("error %s\nSHUTTING DOWN", err.Error())
		}
	}
}

func signalHandler(sc chan os.Signal, srv *http.Server) {
	signal.Notify(sc, os.Interrupt)
	<-sc
	fmt.Printf("\nshutdown signal received\n")
	shutdownServer(srv)
}

func shutdownServer(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

type CounterHandler struct {
}

func NewCounterHandler() *CounterHandler {
	globalCounters.startsAt = time.Now().Unix()
	return &CounterHandler{}
}

func (ch *CounterHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	timestamp := time.Now().Unix()

	buckedIdx := timestamp - globalCounters.startsAt
	curCnt := atomic.AddInt64(&globalCounters.seconds[buckedIdx], 1)
	// get the information from the previous second
	if buckedIdx != 0 {
		curCnt = atomic.LoadInt64(&globalCounters.seconds[buckedIdx-1])
	}
	res := fmt.Sprintf("{\"inst\": %d}", curCnt)
	rw.Write(([]byte)(res))
}
