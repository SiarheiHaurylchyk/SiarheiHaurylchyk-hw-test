package internalhttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage"
)

type Server struct {
	httpServer *http.Server
}

type Logger interface {
	Info(msg string)
}

type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, event storage.Event) error
	DeleteEvent(ctx context.Context, eventID string) error
	ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForWeek(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForMonth(ctx context.Context, date time.Time) ([]storage.Event, error)
}

func NewServer(logger Logger, app Application, host string, port int) *Server {
	events := &eventHandler{app: app}
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("POST /events", events.create)
	mux.HandleFunc("PUT /events/{id}", events.update)
	mux.HandleFunc("DELETE /events/{id}", events.delete)
	mux.HandleFunc("GET /events/day", events.listDay)
	mux.HandleFunc("GET /events/week", events.listWeek)
	mux.HandleFunc("GET /events/month", events.listMonth)

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
