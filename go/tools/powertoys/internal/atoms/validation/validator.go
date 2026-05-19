package validation

import (
	"fmt"
	"strings"
)

// Error represents a validation error.
type Error struct {
	Field   string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Result represents the result of validation.
type Result struct {
	Valid    bool
	Errors   []*Error
	Warnings []string
}

// NewResult creates a new validation result.
func NewResult() *Result {
	return &Result{
		Valid:    true,
		Errors:   make([]*Error, 0),
		Warnings: make([]string, 0),
	}
}

// AddError adds a validation error.
func (r *Result) AddError(field, message string) {
	r.Valid = false
	r.Errors = append(r.Errors, &Error{
		Field:   field,
		Message: message,
	})
}

// AddWarning adds a validation warning.
func (r *Result) AddWarning(message string) {
	r.Warnings = append(r.Warnings, message)
}

// HasErrors returns true if there are validation errors.
func (r *Result) HasErrors() bool {
	return len(r.Errors) > 0
}

// GetErrorMessages returns all error messages as a single string.
func (r *Result) GetErrorMessages() string {
	if !r.HasErrors() {
		return ""
	}

	messages := make([]string, len(r.Errors))
	for i, err := range r.Errors {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}

// Validator is the base interface for validators.
type Validator interface {
	Validate(value interface{}) *Result
}

// RequiredValidator validates that a value is not empty.
type RequiredValidator struct {
	FieldName string
}

// Validate checks if the value is not empty.
func (v *RequiredValidator) Validate(value interface{}) *Result {
	result := NewResult()

	if value == nil {
		result.AddError(v.FieldName, "is required")
		return result
	}

	switch val := value.(type) {
	case string:
		if strings.TrimSpace(val) == "" {
			result.AddError(v.FieldName, "is required")
		}
	case int:
		// Numbers are always present if not nil
	case bool:
		// Booleans are always present if not nil
	default:
		// For other types, check if zero value
		if val == nil {
			result.AddError(v.FieldName, "is required")
		}
	}

	return result
}

// RangeValidator validates that a numeric value is within a range.
type RangeValidator struct {
	FieldName string
	Min       int
	Max       int
}

// Validate checks if the value is within the range.
func (v *RangeValidator) Validate(value interface{}) *Result {
	result := NewResult()

	var num int
	switch val := value.(type) {
	case int:
		num = val
	case int32:
		num = int(val)
	case int64:
		num = int(val)
	default:
		result.AddError(v.FieldName, "must be a number")
		return result
	}

	if num < v.Min || num > v.Max {
		result.AddError(v.FieldName, fmt.Sprintf("must be between %d and %d", v.Min, v.Max))
	}

	return result
}

// PatternValidator validates that a string matches a pattern.
type PatternValidator struct {
	FieldName string
	Pattern   string
}

// Validate checks if the value matches the pattern.
func (v *PatternValidator) Validate(value interface{}) *Result {
	result := NewResult()

	str, ok := value.(string)
	if !ok {
		result.AddError(v.FieldName, "must be a string")
		return result
	}

	// Simple pattern matching - can be enhanced with regex
	if !strings.Contains(str, v.Pattern) {
		result.AddError(v.FieldName, fmt.Sprintf("must contain pattern: %s", v.Pattern))
	}

	return result
}
