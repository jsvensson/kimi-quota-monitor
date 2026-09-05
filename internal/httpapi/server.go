// Package httpapi serves the latest quota payload over HTTP.
package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// Server serves the latest quota payload at GET /quota.
type Server struct {
	addr string
	// payload holds the latest payload as []byte; nil until the first poll.
	payload atomic.Value
}

// NewServer returns a Server that listens on addr.
func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

// SetPayload stores data as the payload served by the handler.
func (s *Server) SetPayload(data []byte) {
	s.payload.Store(data)
}

// Handler returns the HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/quota", s.handleQuota)
	return mux
}

// handleQuota serves the latest payload, or 503 if no payload exists yet.
func (s *Server) handleQuota(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	data, ok := s.payload.Load().([]byte)
	if !ok || data == nil {
		http.Error(w, "no payload available yet", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

// Run starts the HTTP server and shuts it down when ctx is done.
func (s *Server) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:              s.addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		return <-errCh
	}
}
