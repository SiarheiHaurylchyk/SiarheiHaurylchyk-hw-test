package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage/memory"
)

type stubLogger struct{}

func (stubLogger) Error(string) {}

func TestListPeriodBounds(t *testing.T) {
	calendar := New(stubLogger{}, memorystorage.New())
	ctx := context.Background()
	may10 := time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC)

	events := []storage.Event{
		newStoredEvent("may10", may10.Add(10*time.Hour)),
		newStoredEvent("may16", time.Date(2026, time.May, 16, 23, 0, 0, 0, time.UTC)),
		newStoredEvent("may17", time.Date(2026, time.May, 17, 0, 0, 0, 0, time.UTC)),
		newStoredEvent("jun10", time.Date(2026, time.June, 10, 0, 0, 0, 0, time.UTC)),
	}
	for _, event := range events {
		if err := calendar.CreateEvent(ctx, event); err != nil {
			t.Fatal(err)
		}
	}

	day, err := calendar.ListEventsForDay(ctx, may10.Add(15*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	assertIDs(t, day, "may10")

	week, err := calendar.ListEventsForWeek(ctx, may10)
	if err != nil {
		t.Fatal(err)
	}
	assertIDs(t, week, "may10", "may16")

	month, err := calendar.ListEventsForMonth(ctx, time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	assertIDs(t, month, "may10", "may16", "may17")
}

func TestRejectsInvalidAndBusyEvents(t *testing.T) {
	calendar := New(stubLogger{}, memorystorage.New())
	ctx := context.Background()

	if err := calendar.CreateEvent(ctx, storage.Event{}); !errors.Is(err, storage.ErrInvalidEvent) {
		t.Fatalf("empty event: %v", err)
	}
	if _, err := calendar.ListEventsForDay(ctx, time.Time{}); !errors.Is(err, storage.ErrInvalidEvent) {
		t.Fatalf("zero date: %v", err)
	}

	event := newStoredEvent("1", time.Date(2026, time.May, 10, 10, 0, 0, 0, time.UTC))
	if err := calendar.CreateEvent(ctx, event); err != nil {
		t.Fatal(err)
	}
	event.ID = "2"
	if err := calendar.CreateEvent(ctx, event); !errors.Is(err, storage.ErrDateBusy) {
		t.Fatalf("busy time: %v", err)
	}
	if err := calendar.DeleteEvent(ctx, " "); !errors.Is(err, storage.ErrInvalidEvent) {
		t.Fatalf("blank id: %v", err)
	}
	if err := calendar.DeleteEvent(ctx, "missing"); !errors.Is(err, storage.ErrEventNotFound) {
		t.Fatalf("missing event: %v", err)
	}
}

func newStoredEvent(id string, start time.Time) storage.Event {
	return storage.Event{
		ID:        id,
		Title:     "meeting",
		UserID:    "user",
		StartTime: start,
		EndTime:   start.Add(time.Hour),
	}
}

func assertIDs(t *testing.T, events []storage.Event, ids ...string) {
	t.Helper()
	if len(events) != len(ids) {
		t.Fatalf("ids: got %d events, want %d (%v)", len(events), len(ids), events)
	}
	for i, id := range ids {
		if events[i].ID != id {
			t.Fatalf("event %d: got %s, want %s", i, events[i].ID, id)
		}
	}
}
