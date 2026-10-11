package memorystorage

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage"
)

type Storage struct {
	mu     sync.RWMutex
	events map[string]storage.Event
}

func New() *Storage {
	return &Storage{events: make(map[string]storage.Event)}
}

func (s *Storage) CreateEvent(_ context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isTimeBusy(event) {
		return storage.ErrDateBusy
	}

	s.events[event.ID] = event
	return nil
}

func (s *Storage) UpdateEvent(_ context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, found := s.events[event.ID]; !found {
		return storage.ErrEventNotFound
	}
	if s.isTimeBusy(event) {
		return storage.ErrDateBusy
	}

	s.events[event.ID] = event
	return nil
}

func (s *Storage) DeleteEvent(_ context.Context, eventID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, found := s.events[eventID]; !found {
		return storage.ErrEventNotFound
	}

	delete(s.events, eventID)
	return nil
}

func (s *Storage) ListEvents(_ context.Context, from, to time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]storage.Event, 0)
	for _, event := range s.events {
		if !event.StartTime.Before(from) && event.StartTime.Before(to) {
			result = append(result, event)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime.Before(result[j].StartTime)
	})
	return result, nil
}

func (s *Storage) isTimeBusy(newEvent storage.Event) bool {
	for _, event := range s.events {
		if event.UserID != newEvent.UserID || event.ID == newEvent.ID {
			continue
		}
		if event.StartTime.Before(newEvent.EndTime) && newEvent.StartTime.Before(event.EndTime) {
			return true
		}
	}
	return false
}
