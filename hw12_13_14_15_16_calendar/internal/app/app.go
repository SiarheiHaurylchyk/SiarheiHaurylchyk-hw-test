package app

import (
	"context"
	"strings"
	"time"

	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage"
)

type App struct {
	logger  Logger
	storage storage.EventStorage
}

type Logger interface {
	Error(msg string)
}

func New(logger Logger, storage storage.EventStorage) *App {
	return &App{logger: logger, storage: storage}
}

func (a *App) CreateEvent(ctx context.Context, event storage.Event) error {
	if err := validateEvent(event); err != nil {
		return err
	}
	if err := a.storage.CreateEvent(ctx, event); err != nil {
		a.logger.Error("create event: " + err.Error())
		return err
	}
	return nil
}

func (a *App) UpdateEvent(ctx context.Context, event storage.Event) error {
	if err := validateEvent(event); err != nil {
		return err
	}
	if err := a.storage.UpdateEvent(ctx, event); err != nil {
		a.logger.Error("update event: " + err.Error())
		return err
	}
	return nil
}

func (a *App) DeleteEvent(ctx context.Context, eventID string) error {
	if strings.TrimSpace(eventID) == "" {
		return storage.ErrInvalidEvent
	}
	if err := a.storage.DeleteEvent(ctx, eventID); err != nil {
		a.logger.Error("delete event: " + err.Error())
		return err
	}
	return nil
}

func (a *App) ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	return a.listForPeriod(ctx, date, 0, 1)
}

func (a *App) ListEventsForWeek(ctx context.Context, date time.Time) ([]storage.Event, error) {
	return a.listForPeriod(ctx, date, 0, 7)
}

func (a *App) ListEventsForMonth(ctx context.Context, date time.Time) ([]storage.Event, error) {
	return a.listForPeriod(ctx, date, 1, 0)
}

func (a *App) listForPeriod(ctx context.Context, date time.Time, months, days int) ([]storage.Event, error) {
	if date.IsZero() {
		return nil, storage.ErrInvalidEvent
	}

	from := beginningOfDay(date)
	events, err := a.storage.ListEvents(ctx, from, from.AddDate(0, months, days))
	if err != nil {
		a.logger.Error("list events: " + err.Error())
		return nil, err
	}
	return events, nil
}

func validateEvent(event storage.Event) error {
	emptyID := strings.TrimSpace(event.ID) == ""
	emptyTitle := strings.TrimSpace(event.Title) == ""
	emptyUser := strings.TrimSpace(event.UserID) == ""
	if emptyID || emptyTitle || emptyUser {
		return storage.ErrInvalidEvent
	}
	if !event.EndTime.After(event.StartTime) {
		return storage.ErrInvalidEvent
	}
	return nil
}

func beginningOfDay(date time.Time) time.Time {
	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, date.Location())
}
