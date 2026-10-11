package internalhttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

type Logger interface {
	Info(msg string)
}

type Application interface{}

func NewServer(logger Logger, _ Application, host string, port int) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", helloHandler)

	return &Server{
		httpServer: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", host, port),
			Handler:           loggingMiddleware(logger, mux),
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	err := s.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func helloHandler(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("hello world\n"))
}
