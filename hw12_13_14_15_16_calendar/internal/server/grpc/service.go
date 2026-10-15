package internalgrpc

import (
	"context"
	"errors"
	"time"

	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/server/grpc/pb"
	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, event storage.Event) error
	DeleteEvent(ctx context.Context, eventID string) error
	ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForWeek(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForMonth(ctx context.Context, date time.Time) ([]storage.Event, error)
}

type eventService struct {
	pb.UnimplementedEventServiceServer
	app Application
}

func newEventService(app Application) *eventService {
	return &eventService{app: app}
}

func (s *eventService) CreateEvent(ctx context.Context, req *pb.CreateEventRequest) (*pb.CreateEventResponse, error) {
	event := eventFromPB(req.GetEvent())
	if err := s.app.CreateEvent(ctx, event); err != nil {
		return nil, statusFromError(err)
	}
	return &pb.CreateEventResponse{Event: eventToPB(event)}, nil
}

func (s *eventService) UpdateEvent(ctx context.Context, req *pb.UpdateEventRequest) (*pb.UpdateEventResponse, error) {
	event := eventFromPB(req.GetEvent())
	if req.GetId() != "" {
		event.ID = req.GetId()
	}
	if err := s.app.UpdateEvent(ctx, event); err != nil {
		return nil, statusFromError(err)
	}
	return &pb.UpdateEventResponse{Event: eventToPB(event)}, nil
}

func (s *eventService) DeleteEvent(ctx context.Context, req *pb.DeleteEventRequest) (*pb.DeleteEventResponse, error) {
	if err := s.app.DeleteEvent(ctx, req.GetId()); err != nil {
		return nil, statusFromError(err)
	}
	return &pb.DeleteEventResponse{}, nil
}

func (s *eventService) ListEventsForDay(
	ctx context.Context,
	req *pb.ListEventsRequest,
) (*pb.ListEventsResponse, error) {
	return s.listEvents(ctx, req.GetDate(), s.app.ListEventsForDay)
}

func (s *eventService) ListEventsForWeek(
	ctx context.Context,
	req *pb.ListEventsRequest,
) (*pb.ListEventsResponse, error) {
	return s.listEvents(ctx, req.GetDate(), s.app.ListEventsForWeek)
}

func (s *eventService) ListEventsForMonth(
	ctx context.Context,
	req *pb.ListEventsRequest,
) (*pb.ListEventsResponse, error) {
	return s.listEvents(ctx, req.GetDate(), s.app.ListEventsForMonth)
}

func (s *eventService) listEvents(
	ctx context.Context,
	date *timestamppb.Timestamp,
	list func(context.Context, time.Time) ([]storage.Event, error),
) (*pb.ListEventsResponse, error) {
	events, err := list(ctx, timestampToTime(date))
	if err != nil {
		return nil, statusFromError(err)
	}
	return &pb.ListEventsResponse{Events: eventsToPB(events)}, nil
}

func statusFromError(err error) error {
	switch {
	case errors.Is(err, storage.ErrInvalidEvent):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, storage.ErrDateBusy):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, storage.ErrEventNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
