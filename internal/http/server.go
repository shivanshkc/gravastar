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
	"github.com/shivanshkc/gravastar/internal/utils/httputils"
)

// Server is the HTTP server of this application.
type Server struct {
	addr       string
	httpServer *http.Server
}

// NewServer returns a new Server instance. Use the Start method to start the server.
func NewServer(addr string, staticDir string, handler *handlers.Handler) (*Server, error) {
	root, err := httputils.NewFileSystemRoot(staticDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open the static directory: %w", err)
	}

	httpHandler := func() http.Handler {
		mux := http.NewServeMux()
		mw := middleware.Middleware{}

		mux.HandleFunc("POST /api/dots", handler.CreateDot)
		mux.HandleFunc("GET /api/dots", handler.ListDots)
		mux.HandleFunc("GET /api/conn", handler.GetConn)

		// Health check.
		mux.HandleFunc("GET /api", func(w http.ResponseWriter, r *http.Request) {
			httputils.Write(w, http.StatusOK, nil, map[string]any{"code": "OK"})
		})

		// Static file server.
		staticServer := http.FileServer(root)
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			staticServer.ServeHTTP(w, r)
		})

		return mw.CORS(mw.AccessLogger(mw.Recovery(mux)))
	}()

	httpServer := &http.Server{Addr: addr, ReadHeaderTimeout: time.Minute, Handler: httpHandler}
	return &Server{addr: addr, httpServer: httpServer}, nil
}

// Start sets up all the dependencies and routes on the server, and calls ListenAndServe on it.
func (s *Server) Start(ctx context.Context) error {
	// Create the HTTP server.

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
	slog.Info("starting http server", "addr", s.addr)
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error in ListenAndServe call: %w", err)
	}

	<-shutdownCompleteChan
	return nil
}
