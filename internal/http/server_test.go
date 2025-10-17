package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/shivanshkc/gravastar/internal/logger"
)

// TestServer_Start checks if the HTTP server starts correctly with all the valid parameters.
func TestServer_Start(t *testing.T) {
	// Start the server with mock dependencies.
	server := mockServerStart(t)
	defer func() { _ = server.httpServer.Shutdown(context.Background()) }()

	// Dummy request with a path that does not exist. We will expect 404.
	reqURI := fmt.Sprintf("http://%s/not-existent-path", "localhost:8080")
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, reqURI, nil)

	// Execute request.
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Errorf("unexpected error in http request execution: %v", err)
		return
	}

	// Cleanup.
	defer func() { _ = resp.Body.Close() }()

	// Verify status code.
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status code to be: %d, but got: %d", http.StatusNotFound, resp.StatusCode)
		return
	}
}

// mockServerStart creates a *Server instance using a mock logger.
// It sleeps for a second to give the server some time to boot up.
func mockServerStart(t *testing.T) *Server {
	// Server dependencies.
	logger.Init(io.Discard, "info", true)

	// Instantiate the server to be tested.
	server, err := NewServer("localhost:8080", ".", nil)
	if err != nil {
		t.Fatal("failed to create server:", err)
		return nil
	}

	// Start the server without blocking.
	go func() { _ = server.Start(context.Background()) }()

	// Wait for the server to start.
	time.Sleep(time.Second)
	return server
}
