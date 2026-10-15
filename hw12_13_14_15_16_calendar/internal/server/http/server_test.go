package internalhttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeLogger struct {
	messages []string
}

func (l *fakeLogger) Info(msg string) {
	l.messages = append(l.messages, msg)
}

func TestHelloAndAccessLog(t *testing.T) {
	logg := &fakeLogger{}
	server := NewServer(logg, nil, "127.0.0.1", 8080)

	request := httptest.NewRequest(http.MethodGet, "/hello?q=1", nil)
	request.RemoteAddr = "66.249.65.3:12345"
	request.Header.Set("User-Agent", "Mozilla/5.0")
	response := httptest.NewRecorder()

	server.httpServer.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status code: got %d, want 200", response.Code)
	}
	if len(logg.messages) != 1 {
		t.Fatalf("log messages: got %d, want 1", len(logg.messages))
	}

	logLine := logg.messages[0]
	for _, part := range []string{"66.249.65.3", "GET /hello?q=1 HTTP/1.1 200", `"Mozilla/5.0"`} {
		if !strings.Contains(logLine, part) {
			t.Fatalf("log line %q must contain %q", logLine, part)
		}
	}
}
