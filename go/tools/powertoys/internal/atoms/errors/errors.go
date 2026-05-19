package errors

import (
	"fmt"
)

// AppError represents an application error with user-friendly message.
type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error.
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Error codes
const (
	ErrCodeFileNotFound  = "FILE_NOT_FOUND"
	ErrCodeFileRead      = "FILE_READ_ERROR"
	ErrCodeFileWrite     = "FILE_WRITE_ERROR"
	ErrCodeInvalidConfig = "INVALID_CONFIG"
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeBackupFailed  = "BACKUP_FAILED"
	ErrCodePathError     = "PATH_ERROR"
)

// User-friendly error messages
var (
	ErrFileNotFound = func(path string) *AppError {
		return NewAppError(
			ErrCodeFileNotFound,
			fmt.Sprintf("File not found: %s", path),
			nil,
		)
	}

	ErrFileRead = func(path string, err error) *AppError {
		return NewAppError(
			ErrCodeFileRead,
			fmt.Sprintf("Failed to read file: %s", path),
			err,
		)
	}

	ErrFileWrite = func(path string, err error) *AppError {
		return NewAppError(
			ErrCodeFileWrite,
			fmt.Sprintf("Failed to write file: %s", path),
			err,
		)
	}

	ErrInvalidConfig = func(details string) *AppError {
		return NewAppError(
			ErrCodeInvalidConfig,
			fmt.Sprintf("Invalid configuration: %s", details),
			nil,
		)
	}

	ErrValidation = func(details string) *AppError {
		return NewAppError(
			ErrCodeValidation,
			fmt.Sprintf("Validation failed: %s", details),
			nil,
		)
	}

	ErrBackupFailed = func(err error) *AppError {
		return NewAppError(
			ErrCodeBackupFailed,
			"Failed to create backup",
			err,
		)
	}

	ErrPathError = func(path string, err error) *AppError {
		return NewAppError(
			ErrCodePathError,
			fmt.Sprintf("Path error: %s", path),
			err,
		)
	}
)
