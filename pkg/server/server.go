package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func LaunchBlockingServer(addr string, hfn http.Handler) {
	if len(addr) == 0 {
		addr = "0.0.0.0:7777"
	}
	srv := &http.Server{
		Addr:    addr,
		Handler: hfn,
	}

	sigChan := make(chan os.Signal, 1)
	go signalHandler(sigChan, srv)

	if err := srv.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			// TODO: change this for a log
			fmt.Printf("error %s\nSHUTTING DOWN", err.Error())
		}
	}
}

func LaunchBackgroundServer(addr string, hfn http.Handler) *http.Server {
	if len(addr) == 0 {
		addr = "0.0.0.0:7777"
	}
	srv := &http.Server{
		Addr:    addr,
		Handler: hfn,
	}

	sigChan := make(chan os.Signal, 1)
	go signalHandler(sigChan, srv)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				// TODO: change this for a log
				fmt.Printf("error %s\nSHUTTING DOWN", err.Error())
			}
		}
	}()
	return srv
}

func signalHandler(sc chan os.Signal, srv *http.Server) {
	signal.Notify(sc, os.Interrupt)
	<-sc
	fmt.Printf("shutdown signal received\n")
	ShutdownServer(srv)
}

func ShutdownServer(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
