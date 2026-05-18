package util

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger(true)
	if logger == nil {
		t.Error("NewLogger() returned nil")
	}
	if !logger.verbose {
		t.Error("NewLogger(true) should set verbose to true")
	}
}

func TestLogger_Debug(t *testing.T) {
	tests := []struct {
		name          string
		verbose       bool
		expectOutput  bool
	}{
		{
			name:         "debug logs when verbose is true",
			verbose:      true,
			expectOutput: true,
		},
		{
			name:         "debug does not log when verbose is false",
			verbose:      false,
			expectOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(tt.verbose)
			logger.SetOutput(&buf)

			logger.Debug("test message")

			output := buf.String()
			hasOutput := len(output) > 0

			if hasOutput != tt.expectOutput {
				t.Errorf("Debug() output = %v, want %v", hasOutput, tt.expectOutput)
			}

			if tt.expectOutput && !strings.Contains(output, "DEBUG") {
				t.Error("Debug() output should contain 'DEBUG'")
			}
		})
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.SetOutput(&buf)

	logger.Info("test info message")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Error("Info() output should contain 'INFO'")
	}
	if !strings.Contains(output, "test info message") {
		t.Error("Info() output should contain the message")
	}
}

func TestInitLogger(t *testing.T) {
	InitLogger(true)
	if globalLogger == nil {
		t.Error("InitLogger() should set globalLogger")
	}
	if !globalLogger.verbose {
		t.Error("InitLogger(true) should set verbose to true")
	}
}

func TestGlobalLogFunctions(t *testing.T) {
	// Initialize global logger
	InitLogger(true)

	// Test that global functions don't panic
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")
}
