package memorystorage

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage"
)

var today = time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC)

func newEvent(id, userID string, startHour int) storage.Event {
	startTime := today.Add(time.Duration(startHour) * time.Hour)
	return storage.Event{
		ID:        id,
		Title:     "meeting",
		UserID:    userID,
		StartTime: startTime,
		EndTime:   startTime.Add(time.Hour),
	}
}

func TestCreateAndListEvents(t *testing.T) {
	eventStorage := New()
	ctx := context.Background()

	if err := eventStorage.CreateEvent(ctx, newEvent("2", "user", 15)); err != nil {
		t.Fatal(err)
	}
	if err := eventStorage.CreateEvent(ctx, newEvent("1", "user", 10)); err != nil {
		t.Fatal(err)
	}
	if err := eventStorage.CreateEvent(ctx, newEvent("3", "user", 25)); err != nil {
		t.Fatal(err)
	}

	events, err := eventStorage.ListEvents(ctx, today, today.AddDate(0, 0, 1))
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 2 || events[0].ID != "1" || events[1].ID != "2" {
		t.Fatalf("expected events 1 and 2 sorted by time, got %+v", events)
	}
}

func TestCreateEventDateBusy(t *testing.T) {
	eventStorage := New()
	ctx := context.Background()

	if err := eventStorage.CreateEvent(ctx, newEvent("1", "user", 10)); err != nil {
		t.Fatal(err)
	}

	err := eventStorage.CreateEvent(ctx, newEvent("2", "user", 10))
	if !errors.Is(err, storage.ErrDateBusy) {
		t.Fatalf("expected ErrDateBusy, got %v", err)
	}

	if err = eventStorage.CreateEvent(ctx, newEvent("3", "other-user", 10)); err != nil {
		t.Fatal(err)
	}
	if err = eventStorage.CreateEvent(ctx, newEvent("4", "user", 11)); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateEvent(t *testing.T) {
	eventStorage := New()
	ctx := context.Background()

	if err := eventStorage.CreateEvent(ctx, newEvent("1", "user", 10)); err != nil {
		t.Fatal(err)
	}
	if err := eventStorage.CreateEvent(ctx, newEvent("2", "user", 12)); err != nil {
		t.Fatal(err)
	}

	changedEvent := newEvent("1", "user", 8)
	changedEvent.Title = "new title"
	if err := eventStorage.UpdateEvent(ctx, changedEvent); err != nil {
		t.Fatal(err)
	}
	events, _ := eventStorage.ListEvents(ctx, today, today.AddDate(0, 0, 1))
	if events[0].Title != "new title" {
		t.Fatalf("event was not updated: %+v", events[0])
	}

	err := eventStorage.UpdateEvent(ctx, newEvent("unknown", "user", 20))
	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}

	err = eventStorage.UpdateEvent(ctx, newEvent("1", "user", 12))
	if !errors.Is(err, storage.ErrDateBusy) {
		t.Fatalf("expected ErrDateBusy, got %v", err)
	}
}

func TestDeleteEvent(t *testing.T) {
	eventStorage := New()
	ctx := context.Background()

	if err := eventStorage.CreateEvent(ctx, newEvent("1", "user", 10)); err != nil {
		t.Fatal(err)
	}
	if err := eventStorage.DeleteEvent(ctx, "1"); err != nil {
		t.Fatal(err)
	}

	events, _ := eventStorage.ListEvents(ctx, today, today.AddDate(0, 0, 1))
	if len(events) != 0 {
		t.Fatalf("expected no events, got %+v", events)
	}

	err := eventStorage.DeleteEvent(ctx, "1")
	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestConcurrentCreateEvent(t *testing.T) {
	eventStorage := New()
	ctx := context.Background()
	const eventsCount = 50

	var wg sync.WaitGroup
	for i := 0; i < eventsCount; i++ {
		wg.Add(1)
		go func(number int) {
			defer wg.Done()
			id := strconv.Itoa(number)
			if err := eventStorage.CreateEvent(ctx, newEvent(id, "user-"+id, 10)); err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()

	events, err := eventStorage.ListEvents(ctx, today, today.AddDate(0, 0, 1))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != eventsCount {
		t.Fatalf("expected %d events, got %d", eventsCount, len(events))
	}
}
