package app

import (
	"context"

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

func (a *App) CreateEvent(ctx context.Context, id, title string) error {
	err := a.storage.CreateEvent(ctx, storage.Event{ID: id, Title: title})
	if err != nil {
		a.logger.Error("create event: " + err.Error())
	}
	return err
}
