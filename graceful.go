// Package graceful provides a server graceful draining helper.
package graceful

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Drain takes a http.Server and drains on certain os.Signals
// and returns a future to block the server until fully drained.
func Drain(srv *http.Server) <-chan struct{} {
	idleConns := make(chan struct{})
	go func() {
		q := make(chan os.Signal, 1)
		signal.Notify(q, syscall.SIGTERM, os.Interrupt)
		<-q
		log.Println("[graceful] Starting shutdown sequence")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("[graceful] Could not shutdown gracefully: %v", err)
		}
		close(idleConns)
	}()
	return idleConns
}

// DrainWithContext behaves the same as Drain but can be canceled with ctx.
func DrainWithContext(ctx context.Context, srv *http.Server) <-chan struct{} {
	idleConns := make(chan struct{})
	go func() {
		q := make(chan os.Signal, 1)
		signal.Notify(q, syscall.SIGTERM, os.Interrupt)

		// Wait for signal either from ctx or q
		select {
		case signal := <-q:
			log.Printf("[graceful] Shutting down due to signal: %v", signal)
		case <-ctx.Done():
			log.Printf("[graceful] Shutting down due to ctx: %v", ctx.Err())
		}

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("[graceful] Could not shutdown gracefully: %v", err)
		}
		close(idleConns)
	}()
	return idleConns
}
