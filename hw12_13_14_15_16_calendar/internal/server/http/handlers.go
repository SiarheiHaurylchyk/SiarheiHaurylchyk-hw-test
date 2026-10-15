package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage"
)

const contentTypeJSON = "application/json"

type eventHandler struct {
	app Application
}

type eventDTO struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	Description  string    `json:"description"`
	UserID       string    `json:"userId"`
	NotifyBefore string    `json:"notifyBefore"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *eventHandler) create(w http.ResponseWriter, r *http.Request) {
	event, err := decodeEvent(r)
	if err != nil {
		writeError(w, storage.ErrInvalidEvent)
		return
	}
	if err = h.app.CreateEvent(r.Context(), event); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, eventToDTO(event))
}

func (h *eventHandler) update(w http.ResponseWriter, r *http.Request) {
	event, err := decodeEvent(r)
	if err != nil {
		writeError(w, storage.ErrInvalidEvent)
		return
	}
	event.ID = r.PathValue("id")
	if err = h.app.UpdateEvent(r.Context(), event); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, eventToDTO(event))
}

func (h *eventHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.app.DeleteEvent(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *eventHandler) listDay(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, h.app.ListEventsForDay)
}

func (h *eventHandler) listWeek(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, h.app.ListEventsForWeek)
}

func (h *eventHandler) listMonth(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, h.app.ListEventsForMonth)
}

func (h *eventHandler) list(
	w http.ResponseWriter,
	r *http.Request,
	list func(context.Context, time.Time) ([]storage.Event, error),
) {
	date, err := parseDate(r.URL.Query().Get("date"))
	if err != nil {
		writeError(w, err)
		return
	}

	events, err := list(r.Context(), date)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, eventsToDTO(events))
}

func decodeEvent(r *http.Request) (storage.Event, error) {
	var dto eventDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return storage.Event{}, err
	}

	var notifyBefore time.Duration
	if dto.NotifyBefore != "" {
		var err error
		notifyBefore, err = time.ParseDuration(dto.NotifyBefore)
		if err != nil {
			return storage.Event{}, err
		}
	}

	return storage.Event{
		ID:           dto.ID,
		Title:        dto.Title,
		StartTime:    dto.StartTime,
		EndTime:      dto.EndTime,
		Description:  dto.Description,
		UserID:       dto.UserID,
		NotifyBefore: notifyBefore,
	}, nil
}

func eventToDTO(event storage.Event) eventDTO {
	return eventDTO{
		ID:           event.ID,
		Title:        event.Title,
		StartTime:    event.StartTime,
		EndTime:      event.EndTime,
		Description:  event.Description,
		UserID:       event.UserID,
		NotifyBefore: event.NotifyBefore.String(),
	}
}

func eventsToDTO(events []storage.Event) []eventDTO {
	result := make([]eventDTO, 0, len(events))
	for _, event := range events {
		result = append(result, eventToDTO(event))
	}
	return result
}

func parseDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, storage.ErrInvalidEvent
	}
	if date, err := time.Parse(time.DateOnly, value); err == nil {
		return date, nil
	}
	if date, err := time.Parse(time.RFC3339, value); err == nil {
		return date, nil
	}
	return time.Time{}, storage.ErrInvalidEvent
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return
	}
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, storage.ErrInvalidEvent):
		status = http.StatusBadRequest
	case errors.Is(err, storage.ErrDateBusy):
		status = http.StatusConflict
	case errors.Is(err, storage.ErrEventNotFound):
		status = http.StatusNotFound
	}
	writeJSON(w, status, errorResponse{Error: err.Error()})
}
