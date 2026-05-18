// Package logging provides a simple levelled logger used throughout the db-solver tool.
//
// It wraps log/slog with an application-level "Verbose" concept (debug messages
// that are gated on an explicit verbose flag rather than the slog level).
// The singleton GetLogger() pattern is intentional — every internal package
// calls it so the application entry point can configure logging once.
package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// LogLevel represents the logging level.
type LogLevel int

const (
	// DEBUG logs wire-level details.
	DEBUG LogLevel = iota
	// INFO logs significant lifecycle events.
	INFO
	// WARN logs recoverable problems.
	WARN
	// ERROR logs operations that failed and were surfaced to the caller.
	ERROR
)

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses a log level string.
func ParseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn", "warning":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

// Logger wraps slog with a verbose mode and optional file output.
type Logger struct {
	slogger *slog.Logger
	verbose bool
	level   LogLevel
	// fileLogger is a secondary slog logger writing to a file when configured.
	fileLogger *slog.Logger
}

// NewLogger creates a new Logger instance.
func NewLogger(level LogLevel, verbose bool, logFile string) (*Logger, error) {
	var slogLevel slog.Level
	switch level {
	case DEBUG:
		slogLevel = slog.LevelDebug
	case INFO:
		slogLevel = slog.LevelInfo
	case WARN:
		slogLevel = slog.LevelWarn
	case ERROR:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slogLevel})
	l := &Logger{
		slogger: slog.New(handler),
		verbose: verbose,
		level:   level,
	}

	if logFile != "" {
		dir := filepath.Dir(logFile)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create log directory: %w", err)
		}
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, fmt.Errorf("open log file: %w", err)
		}
		fh := slog.NewTextHandler(f, &slog.HandlerOptions{Level: slogLevel})
		l.fileLogger = slog.New(fh)
	}

	return l, nil
}

func (l *Logger) emit(lvl slog.Level, msg string) {
	l.slogger.Log(nil, lvl, msg) //nolint:staticcheck // context intentionally nil for process-lifetime logger
	if l.fileLogger != nil {
		l.fileLogger.Log(nil, lvl, msg) //nolint:staticcheck
	}
}

// Debug logs a debug message.
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= DEBUG {
		l.emit(slog.LevelDebug, fmt.Sprintf(format, args...))
	}
}

// Info logs an info message.
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= INFO {
		l.emit(slog.LevelInfo, fmt.Sprintf(format, args...))
	}
}

// Warn logs a warning message.
func (l *Logger) Warn(format string, args ...interface{}) {
	l.emit(slog.LevelWarn, fmt.Sprintf(format, args...))
}

// Error logs an error message.
func (l *Logger) Error(format string, args ...interface{}) {
	l.emit(slog.LevelError, fmt.Sprintf(format, args...))
}

// Verbose logs a verbose (debug) message only when verbose mode is enabled.
func (l *Logger) Verbose(format string, args ...interface{}) {
	if l.verbose {
		l.emit(slog.LevelDebug, "VERBOSE: "+fmt.Sprintf(format, args...))
	}
}

// SetVerbose sets the verbose mode.
func (l *Logger) SetVerbose(verbose bool) { l.verbose = verbose }

// SetLevel sets the logging level.
func (l *Logger) SetLevel(level LogLevel) { l.level = level }

// globalLogger is the process-lifetime singleton.
var globalLogger *Logger

// InitLogger initialises the global logger.
func InitLogger(level LogLevel, verbose bool, logFile string) error {
	var err error
	globalLogger, err = NewLogger(level, verbose, logFile)
	return err
}

// GetLogger returns the global logger, creating a default one if needed.
func GetLogger() *Logger {
	if globalLogger == nil {
		globalLogger, _ = NewLogger(INFO, false, "")
	}
	return globalLogger
}
