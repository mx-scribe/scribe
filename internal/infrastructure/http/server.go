package http

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/mx-scribe/scribe/internal/infrastructure/http/handlers"
	"github.com/mx-scribe/scribe/internal/infrastructure/persistence/sqlite"
)

// Server represents the HTTP server.
type Server struct {
	router   *chi.Mux
	server   *http.Server
	db       *sqlite.Database
	staticFS fs.FS
	sseHub   *handlers.SSEHub
}

// NewServer creates a new HTTP server.
func NewServer(db *sqlite.Database) *Server {
	s := &Server{
		router: chi.NewRouter(),
		db:     db,
		sseHub: handlers.NewSSEHub(),
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// SSEHub returns the SSE hub for broadcasting events.
func (s *Server) SSEHub() *handlers.SSEHub {
	return s.sseHub
}

// Start starts the HTTP server with graceful shutdown.
func (s *Server) Start(port int) error {
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		fmt.Printf("SCRIBE server starting on http://localhost:%d\n", port)
		serverErrors <- s.server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-shutdown:
		fmt.Printf("\nReceived %v signal, shutting down...\n", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.server.Shutdown(ctx); err != nil {
			s.server.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}

// Router returns the chi router for testing.
func (s *Server) Router() *chi.Mux {
	return s.router
}

// DB returns the database for handlers.
func (s *Server) DB() *sqlite.Database {
	return s.db
}
