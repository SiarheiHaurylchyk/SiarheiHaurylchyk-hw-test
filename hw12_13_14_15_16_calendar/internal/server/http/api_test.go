package internalhttp

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/app"
	memorystorage "github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage/memory"
)

type noopLogger struct{}

func (noopLogger) Error(string) {}

func TestHTTPEventFlow(t *testing.T) {
	logg := &fakeLogger{}
	handler := newAPIHandler(logg)

	event := sampleEvent()
	createdResponse := doJSON(t, handler, http.MethodPost, "/events", event)
	requireStatus(t, createdResponse, http.StatusCreated)

	var created eventDTO
	if err := json.Unmarshal(createdResponse.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ID != "1" || created.Title != "meeting" || created.NotifyBefore != "30m0s" {
		t.Fatalf("created event: %+v", created)
	}
	if len(logg.messages) != 1 || !strings.Contains(logg.messages[0], "66.249.65.3") ||
		!strings.Contains(logg.messages[0], "POST /events HTTP/1.1 201") {
		t.Fatalf("access log: %v", logg.messages)
	}

	day := doJSON(t, handler, http.MethodGet, "/events/day?date=2026-05-10", nil)
	requireStatus(t, day, http.StatusOK)
	if events := decodeEvents(t, day); len(events) != 1 || events[0].ID != "1" {
		t.Fatalf("day events: %+v", events)
	}

	week := doJSON(t, handler, http.MethodGet, "/events/week?date=2026-05-10", nil)
	requireStatus(t, week, http.StatusOK)
	if len(decodeEvents(t, week)) != 1 {
		t.Fatal("event must be included in the week")
	}

	month := doJSON(t, handler, http.MethodGet, "/events/month?date=2026-05-01", nil)
	requireStatus(t, month, http.StatusOK)
	if len(decodeEvents(t, month)) != 1 {
		t.Fatal("event must be included in the month")
	}

	otherDay := doJSON(t, handler, http.MethodGet, "/events/day?date=2026-05-11", nil)
	requireStatus(t, otherDay, http.StatusOK)
	if body := strings.TrimSpace(otherDay.Body.String()); body != "[]" {
		t.Fatalf("other day: %s", body)
	}

	event.Title = "renamed"
	updated := doJSON(t, handler, http.MethodPut, "/events/1", event)
	requireStatus(t, updated, http.StatusOK)

	deleted := doJSON(t, handler, http.MethodDelete, "/events/1", nil)
	requireStatus(t, deleted, http.StatusNoContent)
}

func TestHTTPEventErrors(t *testing.T) {
	handler := newAPIHandler(&fakeLogger{})

	request := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader("{"))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	requireStatus(t, response, http.StatusBadRequest)

	invalid := doJSON(t, handler, http.MethodPost, "/events", eventDTO{Title: "meeting"})
	requireStatus(t, invalid, http.StatusBadRequest)

	event := sampleEvent()
	requireStatus(t, doJSON(t, handler, http.MethodPost, "/events", event), http.StatusCreated)

	conflict := sampleEvent()
	conflict.ID = "2"
	requireStatus(t, doJSON(t, handler, http.MethodPost, "/events", conflict), http.StatusConflict)

	otherUser := sampleEvent()
	otherUser.ID = "3"
	otherUser.UserID = "other"
	requireStatus(t, doJSON(t, handler, http.MethodPost, "/events", otherUser), http.StatusCreated)

	requireStatus(t, doJSON(t, handler, http.MethodPut, "/events/missing", event), http.StatusNotFound)
	requireStatus(t, doJSON(t, handler, http.MethodDelete, "/events/missing", nil), http.StatusNotFound)
	requireStatus(t, doJSON(t, handler, http.MethodGet, "/events/day?date=tomorrow", nil), http.StatusBadRequest)
	requireStatus(t, doJSON(t, handler, http.MethodGet, "/events/week", nil), http.StatusBadRequest)
}

func newAPIHandler(logg *fakeLogger) http.Handler {
	calendar := app.New(noopLogger{}, memorystorage.New())
	return NewServer(logg, calendar, "127.0.0.1", 8080).httpServer.Handler
}

func sampleEvent() eventDTO {
	start := time.Date(2026, time.May, 10, 10, 0, 0, 0, time.UTC)
	return eventDTO{
		ID:           "1",
		Title:        "meeting",
		StartTime:    start,
		EndTime:      start.Add(time.Hour),
		Description:  "daily",
		UserID:       "user",
		NotifyBefore: "30m",
	}
}

func doJSON(t *testing.T, handler http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var payload io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		payload = strings.NewReader(string(data))
	}

	request := httptest.NewRequest(method, path, payload)
	request.RemoteAddr = "66.249.65.3:12345"
	request.Header.Set("User-Agent", "calendar-test")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func decodeEvents(t *testing.T, response *httptest.ResponseRecorder) []eventDTO {
	t.Helper()
	var events []eventDTO
	if err := json.Unmarshal(response.Body.Bytes(), &events); err != nil {
		t.Fatal(err)
	}
	return events
}

func requireStatus(t *testing.T, response *httptest.ResponseRecorder, status int) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("status: got %d, want %d, body: %s", response.Code, status, response.Body.String())
	}
}
