package logger

import (
	"fmt"
	"sync"
	"time"
)

type LogLevel int

const (
	LevelInfo LogLevel = iota
	LevelWarn
	LevelError
)

func (l LogLevel) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type LogEntry struct {
	Time    time.Time
	Level   LogLevel
	Message string
}

func (le LogEntry) String() string {
	return fmt.Sprintf("[%s] %s: %s",
		le.Time.Format("2006-01-02 15:04:05"),
		le.Level.String(),
		le.Message)
}

type Logger struct {
	entries []LogEntry
	mu      sync.RWMutex
	onLog   func(LogEntry)
}

func NewLogger() *Logger {
	return &Logger{
		entries: make([]LogEntry, 0),
	}
}

func (l *Logger) SetOnLog(callback func(LogEntry)) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.onLog = callback
}

func (l *Logger) log(level LogLevel, message string) {
	entry := LogEntry{
		Time:    time.Now(),
		Level:   level,
		Message: message,
	}

	l.mu.Lock()
	l.entries = append(l.entries, entry)
	if len(l.entries) > 1000 {
		l.entries = l.entries[len(l.entries)-1000:]
	}
	callback := l.onLog
	l.mu.Unlock()

	if callback != nil {
		callback(entry)
	}
}

func (l *Logger) Info(message string) {
	l.log(LevelInfo, message)
}

func (l *Logger) Warn(message string) {
	l.log(LevelWarn, message)
}

func (l *Logger) Error(message string) {
	l.log(LevelError, message)
}

func (l *Logger) GetAll() []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	entries := make([]LogEntry, len(l.entries))
	copy(entries, l.entries)
	return entries
}

func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = make([]LogEntry, 0)
}

