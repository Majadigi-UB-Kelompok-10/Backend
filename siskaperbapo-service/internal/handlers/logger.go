package handlers

import (
	"fmt"
	"log"
	"time"
)

type LogLevel string

const (
	LogDebug LogLevel = "DEBUG"
	LogInfo  LogLevel = "INFO"
	LogWarn  LogLevel = "WARN"
	LogError LogLevel = "ERROR"
)

type LogEntry struct {
	Level     LogLevel
	Message   string
	Timestamp time.Time
	Context   map[string]interface{}
	Error     error
}

type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) logEntry(entry LogEntry) {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	logMsg := fmt.Sprintf("[%s] %s - %s", timestamp, entry.Level, entry.Message)

	if entry.Error != nil {
		logMsg += fmt.Sprintf(" | error: %v", entry.Error)
	}

	if len(entry.Context) > 0 {
		logMsg += " | "
		for key, val := range entry.Context {
			logMsg += fmt.Sprintf("%s=%v ", key, val)
		}
	}

	switch entry.Level {
	case LogError:
		log.Printf("%s", logMsg)
	case LogWarn:
		log.Printf("%s", logMsg)
	case LogInfo:
		log.Printf("%s", logMsg)
	case LogDebug:
		log.Printf("%s", logMsg)
	default:
		log.Println(logMsg)
	}
}

func (l *Logger) Debug(msg string, ctx map[string]interface{}) {
	l.logEntry(LogEntry{
		Level:     LogDebug,
		Message:   msg,
		Timestamp: time.Now(),
		Context:   ctx,
	})
}

func (l *Logger) Info(msg string, ctx map[string]interface{}) {
	l.logEntry(LogEntry{
		Level:     LogInfo,
		Message:   msg,
		Timestamp: time.Now(),
		Context:   ctx,
	})
}

func (l *Logger) Warn(msg string, ctx map[string]interface{}) {
	l.logEntry(LogEntry{
		Level:     LogWarn,
		Message:   msg,
		Timestamp: time.Now(),
		Context:   ctx,
	})
}

func (l *Logger) Error(msg string, err error, ctx map[string]interface{}) {
	l.logEntry(LogEntry{
		Level:     LogError,
		Message:   msg,
		Timestamp: time.Now(),
		Error:     err,
		Context:   ctx,
	})
}

func WithContext(pairs ...interface{}) map[string]interface{} {
	ctx := make(map[string]interface{})
	for i := 0; i < len(pairs); i += 2 {
		if i+1 < len(pairs) {
			ctx[pairs[i].(string)] = pairs[i+1]
		}
	}
	return ctx
}
