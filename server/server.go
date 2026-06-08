package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"caos.one/caos/server/store"
)

// Server is the caos HTTP server.
type Server struct {
	mux     http.Handler
	httpSrv *http.Server
	store   *store.Store
	apiKey  string
}

// NewWithPort creates a caos server on the given port with a store rooted at rootDir.
// homePath is the redirect target for GET / (default "/data/d10b49b4").
// apiKey is the optional API key for write protection (empty = writes blocked).
func NewWithPort(rootDir string, port int, homePath string, apiKey string) *Server {
	s, err := store.Open(rootDir)
	if err != nil {
		slog.Error("Failed to open store", "error", err)
		os.Exit(1)
	}

	apiImpl := &apiImpl{store: s, homePath: homePath}
	mux := http.NewServeMux()
	HandlerWithOptions(apiImpl, StdHTTPServerOptions{
		BaseRouter: mux,
		Middlewares: []MiddlewareFunc{APIKeyAuth(apiKey)},
	})

	// Register trailing-slash variant for path autoindex.
	// The generated route GET /path/{addr} redirects to add /, and
	// GET /path/{addr}/ serves the autoindex page.
	mux.HandleFunc("GET /path/{addr}/", func(w http.ResponseWriter, r *http.Request) {
		apiImpl.GetPathAddr(w, r, r.PathValue("addr"))
	})

	srv := &Server{
		mux:    mux,
		store:  s,
		apiKey: apiKey,
	}

	srv.httpSrv = &http.Server{
		Addr:         ":" + strconv.Itoa(port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // 0 = no timeout (required for streaming responses)
		IdleTimeout:  0,
	}

	return srv
}

// Serve starts the HTTP server and blocks until shutdown.
func (s *Server) Serve() error {
	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		slog.Info("Shutting down caos server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpSrv.Shutdown(ctx); err != nil {
			slog.Error("Server forced to shutdown", "error", err)
		}

		if err := s.store.Close(); err != nil {
			slog.Error("Store close error", "error", err)
		}
	}()

	slog.Info("Starting caos server", "addr", s.httpSrv.Addr, "root", "?")
	return s.httpSrv.ListenAndServe()
}

// Shutdown gracefully stops the server and closes the store.
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		return err
	}
	return s.store.Close()
}
