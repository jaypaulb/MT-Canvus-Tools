// Package canvus provides error types and utilities for the Canvus SDK.
package canvus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ErrorCode is the historic machine-readable error-code alias kept for
// backwards-compat with code that branched on string values. New code should
// prefer the sentinel `errors.Is(err, ErrXxx)` form declared below.
type ErrorCode = string

// Sentinel errors. Callers branch with errors.Is.
//
// Drift remediation #1 from go.md: replace string constants with errors.New
// sentinels. The string-typed constants below still exist as aliases so the
// migration is non-breaking, but new code MUST prefer the sentinels.
var (
	// ErrInvalidRequest is returned for HTTP 400-class request validation failures.
	ErrInvalidRequest = errors.New("invalid request")
	// ErrUnauthorized is returned when authentication is missing or rejected (401).
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden is returned when the principal lacks permission for the operation (403).
	ErrForbidden = errors.New("forbidden")
	// ErrNotFound is returned when the target resource does not exist (404).
	ErrNotFound = errors.New("not found")
	// ErrConflict is returned for state conflicts (409).
	ErrConflict = errors.New("conflict")
	// ErrTooManyRequests is returned when the caller is being throttled (429).
	ErrTooManyRequests = errors.New("too many requests")
	// ErrInternalServer is returned for 5xx server faults.
	ErrInternalServer = errors.New("internal server error")
	// ErrNotImplemented is returned for endpoints the server has refused to implement (501).
	ErrNotImplemented = errors.New("not implemented")
	// ErrServiceUnavailable is returned when the server is overloaded or in maintenance (503).
	ErrServiceUnavailable = errors.New("service unavailable")
	// ErrValidation is a generic client-side validation error.
	ErrValidation = errors.New("validation error")
	// ErrRateLimited is the SDK-internal counterpart to ErrTooManyRequests.
	ErrRateLimited = errors.New("rate limited")
	// ErrTimeout is returned when a request exceeds its deadline.
	ErrTimeout = errors.New("timeout")
	// ErrNetwork is returned when the network transport fails.
	ErrNetwork = errors.New("network error")
	// ErrUnexpected covers everything the SDK could not classify.
	ErrUnexpected = errors.New("unexpected error")
)

// Legacy string-typed error codes retained for compatibility. Prefer the
// sentinels above. These will eventually be removed.
const (
	CodeInvalidRequest     ErrorCode = "invalid_request"
	CodeUnauthorized       ErrorCode = "unauthorized"
	CodeForbidden          ErrorCode = "forbidden"
	CodeNotFound           ErrorCode = "not_found"
	CodeConflict           ErrorCode = "conflict"
	CodeTooManyRequests    ErrorCode = "too_many_requests"
	CodeInternalServer     ErrorCode = "internal_server_error"
	CodeNotImplemented     ErrorCode = "not_implemented"
	CodeServiceUnavailable ErrorCode = "service_unavailable"
	CodeValidation         ErrorCode = "validation_error"
	CodeRateLimited        ErrorCode = "rate_limited"
	CodeTimeout            ErrorCode = "timeout"
	CodeNetwork            ErrorCode = "network_error"
	CodeUnexpected         ErrorCode = "unexpected_error"
)

// APIError represents an error returned by the Canvus API.
type APIError struct {
	// StatusCode is the HTTP status code from the API response.
	StatusCode int `json:"status_code"`
	// Code is a machine-readable error code (typically the legacy ErrorCode string).
	Code ErrorCode `json:"code"`
	// Message is a human-readable error message.
	Message string `json:"message"`
	// RequestID is a unique identifier for the request, if available.
	RequestID string `json:"request_id,omitempty"`
	// Details contains additional error details, if any.
	Details map[string]any `json:"details,omitempty"`
	// Wrapped is the underlying error that triggered this one, if any.
	Wrapped error `json:"-"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "API error %d", e.StatusCode)
	if e.Code != "" {
		fmt.Fprintf(&b, " (%s)", e.Code)
	}
	if e.Message != "" {
		if e.Code != "" {
			b.WriteString(": ")
		} else {
			b.WriteString(" ")
		}
		b.WriteString(e.Message)
	}
	if e.RequestID != "" {
		fmt.Fprintf(&b, " [request_id=%s]", e.RequestID)
	}
	if e.Wrapped != nil {
		fmt.Fprintf(&b, ": %v", e.Wrapped)
	}
	return b.String()
}

// Unwrap returns the underlying error if any, and additionally surfaces the
// matching sentinel for errors.Is checks based on the HTTP status code.
func (e *APIError) Unwrap() error {
	if e.Wrapped != nil {
		return e.Wrapped
	}
	switch e.StatusCode {
	case http.StatusBadRequest:
		return ErrInvalidRequest
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusConflict:
		return ErrConflict
	case http.StatusTooManyRequests:
		return ErrTooManyRequests
	case http.StatusInternalServerError:
		return ErrInternalServer
	case http.StatusNotImplemented:
		return ErrNotImplemented
	case http.StatusServiceUnavailable:
		return ErrServiceUnavailable
	}
	return nil
}

// Is reports whether this error matches the target error.
// Supports matching against either another *APIError (status + code) or any of
// the package-level sentinels.
func (e *APIError) Is(target error) bool {
	if t, ok := target.(*APIError); ok {
		if t.StatusCode != 0 && e.StatusCode != t.StatusCode {
			return false
		}
		if t.Code != "" && e.Code != t.Code {
			return false
		}
		return true
	}
	// Sentinel matching via Unwrap.
	if u := e.Unwrap(); u != nil {
		return errors.Is(u, target)
	}
	return false
}

// WithDetails adds additional details to the error.
func (e *APIError) WithDetails(details map[string]any) *APIError {
	e.Details = details
	return e
}

// WithRequestID sets the request ID on the error.
func (e *APIError) WithRequestID(requestID string) *APIError {
	e.RequestID = requestID
	return e
}

// Wrap returns a new error that wraps the current error with additional context.
func (e *APIError) Wrap(err error) *APIError {
	return &APIError{
		StatusCode: e.StatusCode,
		Code:       e.Code,
		Message:    e.Message,
		RequestID:  e.RequestID,
		Details:    e.Details,
		Wrapped:    err,
	}
}

// NewAPIError creates a new APIError with the given status code and message.
func NewAPIError(statusCode int, code ErrorCode, message string) *APIError {
	return &APIError{StatusCode: statusCode, Code: code, Message: message}
}

// ErrorFromStatus creates an appropriate error based on the HTTP status code.
func ErrorFromStatus(statusCode int, message string) error {
	switch statusCode {
	case http.StatusBadRequest:
		return NewAPIError(statusCode, CodeInvalidRequest, message)
	case http.StatusUnauthorized:
		return NewAPIError(statusCode, CodeUnauthorized, message)
	case http.StatusForbidden:
		return NewAPIError(statusCode, CodeForbidden, message)
	case http.StatusNotFound:
		return NewAPIError(statusCode, CodeNotFound, message)
	case http.StatusConflict:
		return NewAPIError(statusCode, CodeConflict, message)
	case http.StatusTooManyRequests:
		return NewAPIError(statusCode, CodeTooManyRequests, message)
	case http.StatusInternalServerError:
		return NewAPIError(statusCode, CodeInternalServer, message)
	case http.StatusNotImplemented:
		return NewAPIError(statusCode, CodeNotImplemented, message)
	case http.StatusServiceUnavailable:
		return NewAPIError(statusCode, CodeServiceUnavailable, message)
	default:
		if statusCode >= 400 && statusCode < 500 {
			return NewAPIError(statusCode, CodeInvalidRequest, message)
		}
		if statusCode >= 500 {
			return NewAPIError(statusCode, CodeInternalServer, message)
		}
		return errors.New(message)
	}
}

// IsContextError checks if the error is a context-related error.
func IsContextError(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == CodeTimeout
	}
	return false
}

// IsRetryableError checks if the error is retryable.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	if IsContextError(err) {
		return false
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return true
	}
	switch apiErr.Code {
	case CodeTooManyRequests, CodeServiceUnavailable, CodeInternalServer:
		return true
	}
	if apiErr.StatusCode >= 500 {
		return true
	}
	return false
}

// ErrorResponse represents a standard error response from the API.
type ErrorResponse struct {
	Error            string         `json:"error,omitempty"`
	ErrorDescription string         `json:"error_description,omitempty"`
	ErrorURI         string         `json:"error_uri,omitempty"`
	RequestID        string         `json:"request_id,omitempty"`
	Details          map[string]any `json:"details,omitempty"`
}

// ParseErrorResponse parses an error response from the API.
func ParseErrorResponse(statusCode int, body []byte) *APIError {
	var resp ErrorResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return NewAPIError(statusCode, "", string(body))
	}
	err := NewAPIError(statusCode, "", resp.ErrorDescription)
	if resp.RequestID != "" {
		err.RequestID = resp.RequestID
	}
	if len(resp.Details) > 0 {
		err.Details = resp.Details
	}
	return err
}

// ValidationError represents a single field validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []*ValidationError

// Error implements the error interface.
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d validation errors:", len(e))
	for _, err := range e {
		fmt.Fprintf(&b, "\n- %s: %s", err.Field, err.Message)
	}
	return b.String()
}

// Add adds a new validation error.
func (e *ValidationErrors) Add(field, message string) {
	*e = append(*e, &ValidationError{Field: field, Message: message})
}

// HasErrors returns true if there are any validation errors.
func (e ValidationErrors) HasErrors() bool { return len(e) > 0 }

// WrapError wraps an error with additional context.
// If err is nil, returns nil. If msg is empty, returns err as-is.
func WrapError(err error, msg string) error {
	if err == nil {
		return nil
	}
	if msg == "" {
		return err
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// WrapErrorf wraps an error with additional formatted context.
// If err is nil, returns nil. If format is empty, returns err as-is.
func WrapErrorf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	if format == "" {
		return err
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}
