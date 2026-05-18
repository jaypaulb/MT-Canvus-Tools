// Package output provides formatting capabilities for CLI output.
// It supports multiple output formats (JSON, table, text, YAML) with
// a unified interface for consistent formatting across all commands.
package output

import (
	"fmt"
)

// Formatter is the interface that all output formatters must implement.
// It provides a single Format method that takes arbitrary data and
// returns a formatted string representation.
type Formatter interface {
	// Format takes arbitrary data and returns a formatted string.
	// Returns an error if the data cannot be formatted.
	Format(data interface{}) (string, error)
}

// NewFormatter creates and returns a Formatter based on the specified format type.
// Supported formats: json, yaml, table, text
// Returns an error if the format is not supported.
func NewFormatter(format string) (Formatter, error) {
	switch format {
	case "json":
		return &JSONFormatter{}, nil
	case "yaml":
		return &YAMLFormatter{}, nil
	case "table":
		return &TableFormatter{}, nil
	case "text":
		return &TextFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported format '%s': must be one of json, yaml, table, text", format)
	}
}
