// Package util provides common utilities for the Canvus CLI.
package util

import (
	"errors"
	"fmt"
	"os"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// Exit codes for different error types
const (
	ExitSuccess          = 0
	ExitGeneralError     = 1
	ExitAuthError        = 2
	ExitNotFound         = 3
	ExitPermissionDenied = 4
	ExitInvalidInput     = 5
)

// HandleError processes an error and returns the appropriate exit code.
// It extracts SDK APIError details if available and formats user-friendly messages.
func HandleError(err error, verbose bool) int {
	if err == nil {
		return ExitSuccess
	}

	// Check if this is a canvus APIError
	var apiErr *canvus.APIError
	if errors.As(err, &apiErr) {
		return handleAPIError(apiErr, verbose)
	}

	// Handle other known error types
	if isValidationError(err) {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return ExitInvalidInput
	}

	// General error
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	if verbose {
		fmt.Fprintf(os.Stderr, "\nStack trace:\n%+v\n", err)
	}
	return ExitGeneralError
}

// handleAPIError handles Canvus API-specific errors.
func handleAPIError(err *canvus.APIError, verbose bool) int {
	statusCode := err.StatusCode
	message := err.Message
	code := err.Code

	// Determine exit code based on HTTP status
	var exitCode int
	switch {
	case statusCode == 401 || statusCode == 403:
		exitCode = ExitAuthError
	case statusCode == 404:
		exitCode = ExitNotFound
	case statusCode == 400 || statusCode == 422:
		exitCode = ExitInvalidInput
	default:
		exitCode = ExitGeneralError
	}

	// Format error message
	fmt.Fprintf(os.Stderr, "API Error: %s (HTTP %d)\n", message, statusCode)
	if code != "" && verbose {
		fmt.Fprintf(os.Stderr, "Error Code: %s\n", code)
	}

	// Provide remediation suggestions
	suggestion := getRemediationSuggestion(statusCode, message)
	if suggestion != "" {
		fmt.Fprintf(os.Stderr, "\nSuggestion: %s\n", suggestion)
	}

	// Show stack trace in verbose mode
	if verbose {
		fmt.Fprintf(os.Stderr, "\nDetailed error:\n%+v\n", err)
	}

	return exitCode
}

// getRemediationSuggestion provides user-friendly suggestions for common errors.
func getRemediationSuggestion(statusCode int, message string) string {
	switch statusCode {
	case 401:
		return "Check your API key or credentials. You may need to login again using 'canvus login'."
	case 403:
		return "You don't have permission to perform this operation. Contact your administrator."
	case 404:
		return "The requested resource was not found. Verify the ID and try again."
	case 400, 422:
		return "Check your input parameters. Use --help to see the correct command syntax."
	case 500, 502, 503:
		return "The server encountered an error. Try again later or contact support."
	default:
		return ""
	}
}

// isValidationError checks if an error is a validation error.
func isValidationError(err error) bool {
	msg := err.Error()
	// Check for common validation error patterns
	return contains(msg, "required") ||
		contains(msg, "invalid") ||
		contains(msg, "must be") ||
		contains(msg, "cannot be empty")
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (
		s[:len(substr)] == substr ||
		s[len(s)-len(substr):] == substr ||
		indexOf(s, substr) >= 0))
}

// indexOf returns the index of substr in s, or -1 if not found.
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// FormatError formats an error message for user display.
func FormatError(err error, verbose bool) string {
	if err == nil {
		return ""
	}

	var apiErr *canvus.APIError
	if errors.As(err, &apiErr) {
		msg := fmt.Sprintf("API Error: %s (HTTP %d)", apiErr.Message, apiErr.StatusCode)
		if apiErr.Code != "" && verbose {
			msg += fmt.Sprintf("\nError Code: %s", apiErr.Code)
		}
		if verbose {
			msg += fmt.Sprintf("\nDetails: %+v", apiErr)
		}
		return msg
	}

	if verbose {
		return fmt.Sprintf("%+v", err)
	}
	return err.Error()
}
