package output

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/config"
)

// PrintOutput formats and prints data to stdout using the specified format.
// If format is empty, it uses the format from config or defaults to "table".
func PrintOutput(ctx context.Context, data interface{}, format string) error {
	return PrintOutputTo(ctx, os.Stdout, data, format)
}

// PrintOutputTo formats and prints data to the specified writer using the specified format.
// If format is empty, it uses the format from config or defaults to "table".
func PrintOutputTo(ctx context.Context, w io.Writer, data interface{}, format string) error {
	// If no format specified, try to get from config
	if format == "" {
		cfg, err := config.GetConfig(ctx)
		if err == nil && cfg.Output != "" {
			format = cfg.Output
		} else {
			// Default to table format
			format = "table"
		}
	}

	// Create formatter
	formatter, err := NewFormatter(format)
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}

	// Format data
	output, err := formatter.Format(data)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	// Print to writer
	fmt.Fprintln(w, output)
	return nil
}

// PrintError formats and prints an error message to stderr.
// The error format depends on the output format for consistency.
func PrintError(ctx context.Context, err error) {
	PrintErrorTo(ctx, os.Stderr, err)
}

// PrintErrorTo formats and prints an error message to the specified writer.
func PrintErrorTo(ctx context.Context, w io.Writer, err error) {
	if err == nil {
		return
	}

	// Get format from config or default to text for errors
	format := "text"
	cfg, configErr := config.GetConfig(ctx)
	if configErr == nil && cfg.Output != "" {
		format = cfg.Output
	}

	// Format error based on output format
	var errorOutput string
	switch format {
	case "json":
		// JSON format for structured error output
		errorData := map[string]interface{}{
			"error":   err.Error(),
			"success": false,
		}
		formatter := &JSONFormatter{}
		output, _ := formatter.Format(errorData)
		errorOutput = output
	case "yaml":
		// YAML format for structured error output
		errorData := map[string]interface{}{
			"error":   err.Error(),
			"success": false,
		}
		formatter := &YAMLFormatter{}
		output, _ := formatter.Format(errorData)
		errorOutput = output
	default:
		// Text format for human-readable errors
		errorOutput = fmt.Sprintf("Error: %v", err)
	}

	fmt.Fprintln(w, errorOutput)
}

// OutputSuccess prints a success message to stdout.
func OutputSuccess(ctx context.Context, message string) {
	OutputSuccessTo(ctx, os.Stdout, message)
}

// OutputSuccessTo prints a success message to the specified writer.
func OutputSuccessTo(ctx context.Context, w io.Writer, message string) {
	// Get format from config or default to text
	format := "text"
	cfg, err := config.GetConfig(ctx)
	if err == nil && cfg.Output != "" {
		format = cfg.Output
	}

	// Format success message based on output format
	var output string
	switch format {
	case "json":
		successData := map[string]interface{}{
			"message": message,
			"success": true,
		}
		formatter := &JSONFormatter{}
		output, _ = formatter.Format(successData)
	case "yaml":
		successData := map[string]interface{}{
			"message": message,
			"success": true,
		}
		formatter := &YAMLFormatter{}
		output, _ = formatter.Format(successData)
	default:
		output = message
	}

	fmt.Fprintln(w, output)
}

// OutputList formats and outputs a list of items using the specified format.
// This is a convenience wrapper around PrintOutput for list-type data.
func OutputList(ctx context.Context, items []interface{}, format string) error {
	return PrintOutput(ctx, items, format)
}

// OutputSingle formats and outputs a single item using the specified format.
// This is a convenience wrapper around PrintOutput for single-item data.
func OutputSingle(ctx context.Context, item interface{}, format string) error {
	return PrintOutput(ctx, item, format)
}
