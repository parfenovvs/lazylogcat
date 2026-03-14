package web

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/parfenovvs/lazylogcat/internal/config"
)

// Server is the HTTP server that serves the web UI and handles API/WebSocket
// requests.
type Server struct {
	httpServer *http.Server
	port       int
}

// NewServer creates a new web server bound to the given port. When demo
// is true, the server returns a fake device and generates synthetic log
// lines instead of requiring adb.
func NewServer(port int, cfg config.Config, demo bool) (*Server, error) {
	mux := http.NewServeMux()

	// REST endpoints
	mux.HandleFunc("GET /api/devices", handleDevices(demo))
	mux.HandleFunc("GET /api/config", handleConfig(cfg))

	// WebSocket
	mux.HandleFunc("GET /ws", handleWebSocket(cfg, demo))

	// Static files from embedded FS
	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return nil, fmt.Errorf("failed to create sub-filesystem: %w", err)
	}
	fileServer := http.FileServer(http.FS(staticSub))
	mux.Handle("/", fileServer)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return &Server{
		httpServer: srv,
		port:       port,
	}, nil
}

// Start begins listening and serving. It blocks until the server is shut down
// or encounters a fatal error.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}

	slog.Info("Web server started", "url", fmt.Sprintf("http://localhost:%d", s.port))

	if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the server with a timeout.
func (s *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	slog.Info("Shutting down web server...")
	return s.httpServer.Shutdown(ctx)
}

// Port returns the configured port.
func (s *Server) Port() int {
	return s.port
}
