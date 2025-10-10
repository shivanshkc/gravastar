package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/shivanshkc/gravastar/internal/logger"
)

// AccessLogger middleware handles access logging.
func (m Middleware) AccessLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// This will be used to calculate the total request execution time.
		start := time.Now()

		// Add an ID to the request to uniquely identify it.
		newCtx := logger.AddContextValue(ctx, "request_id", uuid.NewString())

		// Respect the correlation ID sent by the user.
		correlationID := r.Header.Get("X-Correlation-Id")
		if correlationID == "" {
			// If not sent, generate own.
			newCtx = logger.AddContextValue(newCtx, "correlation_id", uuid.New().String())
		}

		// Update the request with the new context.
		*r = *r.WithContext(newCtx)

		// Embedding the writer into the custom-writer to persist status-code for logging.
		cw := &responseWriterWithCode{ResponseWriter: w}

		// Request entry log.
		slog.InfoContext(ctx, "request received", "url", r.URL.String(), "method", r.Method)
		// Release control to the next middleware or handler.
		next.ServeHTTP(cw, r)
		// Request exit log.
		slog.InfoContext(ctx, "request completed", "latency", time.Since(start), "status", cw.statusCode)
	})
}

// responseWriterWithCode is a wrapper for http.ResponseWriter for persisting statusCode.
type responseWriterWithCode struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseWriterWithCode) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
