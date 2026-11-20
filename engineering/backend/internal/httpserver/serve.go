// Package httpserver holds what the API and MCP servers share to run safely:
// a graceful shutdown on SIGTERM, and server settings that keep slow or
// stalled clients from tying the server up.
package httpserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

const (
	// ReadHeaderTimeout bounds how long a client may take to send its request
	// headers, so connections that trickle in a byte at a time cannot hold the
	// server's connections open (Slowloris).
	ReadHeaderTimeout = 10 * time.Second

	// ShutdownGracePeriod is how long in-flight requests get to finish after
	// SIGTERM. Keep it below the platform's kill deadline (Kubernetes
	// terminationGracePeriodSeconds, 30s by default), or the process is killed
	// before this expires.
	ShutdownGracePeriod = 25 * time.Second

	// The settings below are for NewServer, an ordinary request and response
	// API. The MCP server sets its own: its tool calls run for minutes and its
	// SSE streams stay open, so it cannot use a read or write timeout.
	readTimeout  = 30 * time.Second
	writeTimeout = 60 * time.Second
	idleTimeout  = 2 * time.Minute
)

// NewServer returns an http.Server for an ordinary request and response API,
// with timeouts on every stage of a request: reading the headers and the body,
// writing the response, and sitting idle between requests. Without them a slow
// or stalled client can hold a connection open indefinitely.
func NewServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: ReadHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

// ServeUntilDone runs srv on ln until it fails or ctx is cancelled. On cancel
// it shuts down gracefully, waiting up to grace for in-flight requests; if they
// do not finish in time it closes their connections and reports an error. A
// clean, completed shutdown returns nil.
//
// onShutdown, if set, runs concurrently with the HTTP drain, and ServeUntilDone
// does not return until it has finished, under the same grace deadline. It is
// how the MCP server ends the long-lived SSE streams that Shutdown would
// otherwise wait on, and waits for work that outlives its HTTP request: a tool
// whose client disconnected keeps running with no connection left for Shutdown
// to wait on, so without this wait the process could exit under it. The
// callback gets the grace deadline as drainCtx and must return once it ends.
func ServeUntilDone(ctx context.Context, srv *http.Server, ln net.Listener, grace time.Duration, onShutdown func(drainCtx context.Context)) error {
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve(ln) }()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	log.Printf("HTTP server shutting down, waiting up to %s for in-flight requests", grace)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), grace)
	defer cancel()
	var drained chan struct{}
	if onShutdown != nil {
		drained = make(chan struct{})
		go func() {
			defer close(drained)
			onShutdown(shutdownCtx)
		}()
	}

	err := srv.Shutdown(shutdownCtx)
	if err == nil && drained != nil {
		// Shutdown returns as soon as no HTTP connection is left, which can be
		// before the callback's work is done. Wait for it, under the same
		// deadline: when the deadline passes, report the work as cut off.
		select {
		case <-drained:
		case <-shutdownCtx.Done():
			err = shutdownCtx.Err()
		}
	}
	if err != nil {
		srv.Close()
		return fmt.Errorf("graceful shutdown did not finish, in-flight work was cut off: %w", err)
	}
	return nil
}
