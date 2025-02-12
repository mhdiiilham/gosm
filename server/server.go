package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Server represents a network server with an IP address, port, and listener.
type Server struct {
	ip       string
	port     string
	listener net.Listener
}

// New initializes and returns a new server instance with the specified port.
// It creates a TCP listener and extracts the assigned IP and port.
// Returns an error if the listener creation fails.
func New(port string) (*Server, error) {
	addr := fmt.Sprintf(":%s", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener on %s: %w", addr, err)
	}

	return &Server{
		ip:       listener.Addr().(*net.TCPAddr).IP.String(),
		port:     strconv.Itoa(listener.Addr().(*net.TCPAddr).Port),
		listener: listener,
	}, nil
}

// ServeHTTP starts an HTTP server and listens for incoming requests.
// It gracefully shuts down the server when the context is canceled.
func (s *Server) ServeHTTP(ctx context.Context, srv *http.Server) error {
	errChan := make(chan error, 1)
	go func() {
		<-ctx.Done()

		shutDownCtx, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()

		logrus.Info("server received signal to shutdown")
		errChan <- srv.Shutdown(shutDownCtx)
	}()

	if err := srv.Serve(s.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to serve: %w", err)
	}

	if err := <-errChan; err != nil {
		return fmt.Errorf("failed to shutdown: %w", err)
	}

	return nil
}

// ServeHTTPHandler is a convenience function to serve an HTTP handler with a new HTTP server instance.
// It delegates to ServeHTTP using a new http.Server instance.
func (s *Server) ServeHTTPHandler(ctx context.Context, handler http.Handler) error {
	return s.ServeHTTP(ctx, &http.Server{
		Handler: handler,
	})
}

// ServeGRPC starts a gRPC server and listens for incoming requests.
// It gracefully stops the gRPC server when the context is canceled.
func (s *Server) ServeGRPC(ctx context.Context, srv *grpc.Server) error {
	errChan := make(chan error, 1)
	go func() {
		<-ctx.Done()

		logrus.Info("server.Server: context closed")
		logrus.Info("server.Server: shutting down")
		srv.GracefulStop()
	}()

	if err := srv.Serve(s.listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("failed to serve: %v", err)
	}

	logrus.Info("server.Server: serving stopped")

	select {
	case err := <-errChan:
		return fmt.Errorf("failed to shutdown: %v", err)
	default:
		return nil
	}
}
