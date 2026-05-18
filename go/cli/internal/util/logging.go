// Package util provides common utilities for the Canvus CLI.
//
// The logging helpers are a thin convenience layer over log/slog (per the
// monorepo's Go conventions doc §5). InitLogger is called once from main and
// installs a text/JSON slog handler on the default logger; the package-level
// Debug/Info/Warn/Error helpers route through slog.Default with printf-style
// formatting so existing call sites keep working.
package util

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// Logger is a small struct kept for backwards-compatible test points.
// It captures a verbosity flag and a writable destination, and exposes the
// same Debug/Info/Warn/Error methods the legacy logger did. Internally it
// builds and uses an slog.Logger.
type Logger struct {
	verbose bool
	out     io.Writer
	logger  *slog.Logger
}

// NewLogger creates a new logger instance. Verbose enables Debug-level
// output; otherwise the threshold is Info.
func NewLogger(verbose bool) *Logger {
	l := &Logger{verbose: verbose, out: os.Stderr}
	l.rebuild()
	return l
}

func (l *Logger) rebuild() {
	level := slog.LevelInfo
	if l.verbose {
		level = slog.LevelDebug
	}
	handler := newHandler(l.out, level)
	l.logger = slog.New(handler)
}

// newHandler picks JSON when LOG_FORMAT=json (per conventions), text otherwise.
func newHandler(out io.Writer, level slog.Level) slog.Handler {
	opts := &slog.HandlerOptions{Level: level}
	if os.Getenv("LOG_FORMAT") == "json" {
		return slog.NewJSONHandler(out, opts)
	}
	return slog.NewTextHandler(out, opts)
}

// SetOutput sets the destination writer. Useful for tests.
func (l *Logger) SetOutput(w io.Writer) {
	l.out = w
	l.rebuild()
}

// Debug logs a debug message (only when verbose is enabled).
func (l *Logger) Debug(format string, args ...interface{}) {
	if !l.verbose {
		return
	}
	l.logger.LogAttrs(context.Background(), slog.LevelDebug, fmt.Sprintf(format, args...))
}

// Info logs an informational message.
func (l *Logger) Info(format string, args ...interface{}) {
	l.logger.LogAttrs(context.Background(), slog.LevelInfo, fmt.Sprintf(format, args...))
}

// Warn logs a warning message.
func (l *Logger) Warn(format string, args ...interface{}) {
	l.logger.LogAttrs(context.Background(), slog.LevelWarn, fmt.Sprintf(format, args...))
}

// Error logs an error message.
func (l *Logger) Error(format string, args ...interface{}) {
	l.logger.LogAttrs(context.Background(), slog.LevelError, fmt.Sprintf(format, args...))
}

// globalLogger is the package-level logger used by InitLogger and the bare
// Debug/Info/Warn/Error helpers. It exists for ergonomic CLI logging — the
// CLI is the application, the SDK is the library; per §5 of go.md a top-level
// slog handler is the right place for verbosity selection.
var globalLogger *Logger

// InitLogger initializes the global logger with verbose setting and installs
// it as the default slog logger so any code that grabs slog.Default observes
// the chosen level/format.
func InitLogger(verbose bool) {
	globalLogger = NewLogger(verbose)
	slog.SetDefault(globalLogger.logger)
}

// Debug logs a debug message using the global logger.
func Debug(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(format, args...)
	}
}

// Info logs an informational message using the global logger.
func Info(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Info(format, args...)
	}
}

// Warn logs a warning message using the global logger.
func Warn(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(format, args...)
	}
}

// Error logs an error message using the global logger.
func Error(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Error(format, args...)
	}
}
