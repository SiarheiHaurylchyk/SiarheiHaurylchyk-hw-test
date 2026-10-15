package internalgrpc

import (
	"context"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/app"
	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/server/grpc/pb"
	memorystorage "github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage/memory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type noopLogger struct{}

func (noopLogger) Error(string) {}

type grpcLogger struct {
	mu       sync.Mutex
	messages []string
}

func (l *grpcLogger) Info(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = append(l.messages, msg)
}

func (l *grpcLogger) contains(part string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, message := range l.messages {
		if strings.Contains(message, part) {
			return true
		}
	}
	return false
}

func TestGRPCEventFlow(t *testing.T) {
	client, logg := newTestClient(t, app.New(noopLogger{}, memorystorage.New()))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Date(2026, time.May, 10, 10, 0, 0, 0, time.UTC)
	created, err := client.CreateEvent(ctx, &pb.CreateEventRequest{
		Event: samplePBEvent("1", "user", start),
	}, grpc.WaitForReady(true))
	if err != nil {
		t.Fatal(err)
	}
	if created.GetEvent().GetTitle() != "meeting" || created.GetEvent().GetNotifyBefore().AsDuration() != 30*time.Minute {
		t.Fatalf("created event: %+v", created.GetEvent())
	}
	if !logg.contains("EventService/CreateEvent") || !logg.contains("127.0.0.1") {
		t.Fatalf("grpc access log does not describe the request: %v", logg.messages)
	}

	day, err := client.ListEventsForDay(ctx, &pb.ListEventsRequest{Date: timestamppb.New(start)})
	if err != nil {
		t.Fatal(err)
	}
	if len(day.GetEvents()) != 1 || day.GetEvents()[0].GetId() != "1" {
		t.Fatalf("day events: %+v", day.GetEvents())
	}

	week, err := client.ListEventsForWeek(ctx, &pb.ListEventsRequest{Date: timestamppb.New(start)})
	if err != nil || len(week.GetEvents()) != 1 {
		t.Fatalf("week events: %v %+v", err, week.GetEvents())
	}
	month, err := client.ListEventsForMonth(ctx, &pb.ListEventsRequest{
		Date: timestamppb.New(time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)),
	})
	if err != nil || len(month.GetEvents()) != 1 {
		t.Fatalf("month events: %v %+v", err, month.GetEvents())
	}

	updatedEvent := samplePBEvent("1", "user", start)
	updatedEvent.Title = "renamed"
	updated, err := client.UpdateEvent(ctx, &pb.UpdateEventRequest{Id: "1", Event: updatedEvent})
	if err != nil || updated.GetEvent().GetTitle() != "renamed" {
		t.Fatalf("update: %v %+v", err, updated.GetEvent())
	}
	if _, err = client.DeleteEvent(ctx, &pb.DeleteEventRequest{Id: "1"}); err != nil {
		t.Fatal(err)
	}
}

func TestGRPCEventErrors(t *testing.T) {
	client, _ := newTestClient(t, app.New(noopLogger{}, memorystorage.New()))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.CreateEvent(ctx, &pb.CreateEventRequest{}, grpc.WaitForReady(true))
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("empty event code: %v", err)
	}

	start := time.Date(2026, time.May, 10, 10, 0, 0, 0, time.UTC)
	_, err = client.CreateEvent(ctx, &pb.CreateEventRequest{Event: samplePBEvent("1", "user", start)})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.CreateEvent(ctx, &pb.CreateEventRequest{Event: samplePBEvent("2", "user", start)})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("busy time code: %v", err)
	}
	_, err = client.CreateEvent(ctx, &pb.CreateEventRequest{Event: samplePBEvent("3", "other", start)})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.DeleteEvent(ctx, &pb.DeleteEventRequest{Id: "missing"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("missing event code: %v", err)
	}
	_, err = client.UpdateEvent(ctx, &pb.UpdateEventRequest{Id: "missing", Event: samplePBEvent("missing", "user", start)})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("missing update code: %v", err)
	}
	_, err = client.ListEventsForDay(ctx, &pb.ListEventsRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("empty date code: %v", err)
	}
}

func newTestClient(t *testing.T, calendar *app.App) (pb.EventServiceClient, *grpcLogger) {
	t.Helper()

	logg := &grpcLogger{}
	server := NewServer(logg, calendar, "127.0.0.1", 0)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.serve(listener)
	}()

	conn, err := grpc.NewClient(
		listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Errorf("close grpc connection: %v", err)
		}
		server.srv.GracefulStop()
		if err := <-serveErr; err != nil {
			t.Errorf("grpc server: %v", err)
		}
	})
	return pb.NewEventServiceClient(conn), logg
}

func samplePBEvent(id, userID string, start time.Time) *pb.Event {
	return &pb.Event{
		Id:           id,
		Title:        "meeting",
		StartTime:    timestamppb.New(start),
		EndTime:      timestamppb.New(start.Add(time.Hour)),
		UserId:       userID,
		NotifyBefore: durationpb.New(30 * time.Minute),
	}
}
