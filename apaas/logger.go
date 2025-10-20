package apaas

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// LoggerLevel represents the verbosity of logs emitted by the client.
type LoggerLevel int

// Logging levels aligned with the Node.js SDK.
const (
	LoggerLevelFatal LoggerLevel = iota
	LoggerLevelError
	LoggerLevelWarn
	LoggerLevelInfo
	LoggerLevelDebug
	LoggerLevelTrace
)

var levelNames = map[LoggerLevel]string{
	LoggerLevelFatal: "FATAL",
	LoggerLevelError: "ERROR",
	LoggerLevelWarn:  "WARN",
	LoggerLevelInfo:  "INFO",
	LoggerLevelDebug: "DEBUG",
	LoggerLevelTrace: "TRACE",
}

// Logger defines the behaviour required by the client for logging.
type Logger interface {
	Log(level LoggerLevel, format string, args ...any)
	SetLevel(level LoggerLevel)
	Level() LoggerLevel
}

type defaultLogger struct {
	mu     sync.RWMutex
	level  LoggerLevel
	logger *log.Logger
}

func newDefaultLogger() *defaultLogger {
	return &defaultLogger{
		level:  LoggerLevelInfo,
		logger: log.New(os.Stdout, "", 0),
	}
}

func (l *defaultLogger) Log(level LoggerLevel, format string, args ...any) {
	l.mu.RLock()
	currentLevel := l.level
	l.mu.RUnlock()

	if level > currentLevel {
		return
	}

	name, ok := levelNames[level]
	if !ok {
		name = fmt.Sprintf("LEVEL_%d", level)
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] [%s] %s", name, timestamp, message)
}

func (l *defaultLogger) SetLevel(level LoggerLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *defaultLogger) Level() LoggerLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// String returns the log level name.
func (l LoggerLevel) String() string {
	if name, ok := levelNames[l]; ok {
		return name
	}
	return fmt.Sprintf("LEVEL_%d", l)
}
