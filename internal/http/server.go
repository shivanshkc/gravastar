package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/shivanshkc/gravastar/internal/handlers"
	"github.com/shivanshkc/gravastar/internal/middleware"
	"github.com/shivanshkc/gravastar/internal/utils/errutils"
	"github.com/shivanshkc/gravastar/internal/utils/httputils"
)

// Server is the HTTP server of this application.
type Server struct {
	Handler    *handlers.Handler
	httpServer *http.Server
}

// Start sets up all the dependencies and routes on the server, and calls ListenAndServe on it.
func (s *Server) Start(ctx context.Context, addr string) error {
	// Create the HTTP server.
	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           s.mux(),
		ReadHeaderTimeout: time.Minute,
	}

	// Channel to notify this thread of shut down completion.
	shutdownCompleteChan := make(chan struct{})

	// Cleanup goroutine.
	go func() {
		<-ctx.Done()

		// Shutdown server when the context expires.
		if err := s.httpServer.Shutdown(ctx); err != nil {
			slog.Error("error while shutting down http server: " + err.Error())
		}
		slog.Info("http server shut down complete")

		// Notify the main thread.
		shutdownCompleteChan <- struct{}{}
		close(shutdownCompleteChan)
	}()

	// Blocking call. Will release upon shut down (which happens upon context expiry).
	slog.Info("starting http server", "addr", addr)
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error in ListenAndServe call: %w", err)
	}

	<-shutdownCompleteChan
	return nil
}

// Mux returns the request multiplexer.
func (s *Server) mux() http.Handler {
	mux := http.NewServeMux()
	mw := middleware.Middleware{}

	mux.HandleFunc("POST /api/dots", s.Handler.CreateDot)
	mux.HandleFunc("GET /api/dots", s.Handler.ListDots)

	mux.HandleFunc("GET /api", func(w http.ResponseWriter, r *http.Request) {
		httputils.Write(w, http.StatusOK, nil, map[string]any{"code": "OK"})
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httputils.WriteErr(w, errutils.NotFound())
	})

	return mw.CORS(mw.AccessLogger(mw.Recovery(mux)))
}
