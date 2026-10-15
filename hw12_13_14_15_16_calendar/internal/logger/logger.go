package logger

import (
	"fmt"
	"io"
	"os"
)

const (
	levelDebug = iota
	levelInfo
	levelWarn
	levelError
)

type Logger struct {
	level  int
	output io.Writer
}

func New(level string) *Logger {
	return &Logger{level: parseLevel(level), output: os.Stdout}
}

func (l *Logger) Debug(msg string) {
	l.write(levelDebug, "DEBUG", msg)
}

func (l *Logger) Info(msg string) {
	l.write(levelInfo, "INFO", msg)
}

func (l *Logger) Warn(msg string) {
	l.write(levelWarn, "WARN", msg)
}

func (l *Logger) Error(msg string) {
	l.write(levelError, "ERROR", msg)
}

func (l *Logger) write(level int, levelName, msg string) {
	if level < l.level {
		return
	}
	fmt.Fprintln(l.output, levelName, msg)
}

func parseLevel(level string) int {
	switch level {
	case "debug":
		return levelDebug
	case "warn":
		return levelWarn
	case "error":
		return levelError
	default:
		return levelInfo
	}
}
