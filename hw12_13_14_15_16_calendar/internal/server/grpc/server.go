package internalgrpc

import (
	"context"
	"fmt"
	"net"

	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/server/grpc/pb"
	"google.golang.org/grpc"
)

type Server struct {
	addr string
	srv  *grpc.Server
}

func NewServer(logger Logger, app Application, host string, port int) *Server {
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor(logger)))
	pb.RegisterEventServiceServer(grpcServer, newEventService(app))

	return &Server{
		addr: fmt.Sprintf("%s:%d", host, port),
		srv:  grpcServer,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	return s.serve(listener)
}

func (s *Server) serve(listener net.Listener) error {
	return s.srv.Serve(listener)
}

func (s *Server) Stop(ctx context.Context) error {
	stopped := make(chan struct{})
	go func() {
		s.srv.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		return nil
	case <-ctx.Done():
		s.srv.Stop()
		return ctx.Err()
	}
}
