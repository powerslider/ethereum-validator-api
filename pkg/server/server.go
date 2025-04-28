package server

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"
	"github.com/powerslider/ethereum-validator-api/pkg/config"
)

// Server represents an HTTP server.
type Server struct {
	srv *http.Server
}

// NewServer constructs new HTTP server with the provided muxer.
func NewServer(
	config *config.Config,
	muxer *mux.Router,
) *Server {
	server := &http.Server{
		Addr:              net.JoinHostPort(config.ServerHost, strconv.Itoa(config.ServerPort)),
		ReadHeaderTimeout: 10 * time.Second,
		Handler:           muxer,
	}

	return &Server{
		srv: server,
	}
}

// Start starts the HTTP server.
func (s *Server) Start(errChan chan error) {
	log.Printf("[Start] HTTP server is starting on %s:\n", s.srv.Addr)

	err := s.srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		errChan <- pkgerrors.WithStack(err)
	}
}

// Stop stops the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	log.Println("[Shutdown] HTTP srv is shutting down...")

	shutdownCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	err := s.srv.Shutdown(shutdownCtx)
	if err != nil {
		return pkgerrors.Wrap(pkgerrors.WithStack(err), "shutdown server")
	}

	return nil
}

// Run manages the HTTP server lifecycle on start and on shutdown.
func (s *Server) Run(ctx context.Context) error {
	errChan := make(chan error, 1) // Buffered, avoid possible blocking.

	go s.Start(errChan)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// Stop capturing signals after done.
	defer signal.Stop(sigs)

	select {
	case sig := <-sigs:
		log.Printf("[Signal] Caught OS signal: %srv, shutting down...\n", sig)
		return s.Stop(ctx)
	case err := <-errChan:
		return pkgerrors.Wrap(pkgerrors.WithStack(err), "server error")
	}
}
