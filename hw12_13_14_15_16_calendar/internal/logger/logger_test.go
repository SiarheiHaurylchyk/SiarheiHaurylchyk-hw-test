package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoggerLevel(t *testing.T) {
	output := &bytes.Buffer{}
	logg := &Logger{level: parseLevel("warn"), output: output}

	logg.Info("info message")
	logg.Error("error message")

	text := output.String()
	if strings.Contains(text, "info message") {
		t.Fatalf("info must be hidden on warn level, got: %q", text)
	}
	if !strings.Contains(text, "ERROR error message") {
		t.Fatalf("error must be written, got: %q", text)
	}
}
