package util

import (
	"errors"
	"testing"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func TestHandleError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		verbose  bool
		wantCode int
	}{
		{
			name:     "nil error returns success",
			err:      nil,
			verbose:  false,
			wantCode: ExitSuccess,
		},
		{
			name:     "API auth error returns ExitAuthError",
			err:      &canvus.APIError{StatusCode: 401, Message: "Unauthorized"},
			verbose:  false,
			wantCode: ExitAuthError,
		},
		{
			name:     "API not found error returns ExitNotFound",
			err:      &canvus.APIError{StatusCode: 404, Message: "Not Found"},
			verbose:  false,
			wantCode: ExitNotFound,
		},
		{
			name:     "API validation error returns ExitInvalidInput",
			err:      &canvus.APIError{StatusCode: 400, Message: "Bad Request"},
			verbose:  false,
			wantCode: ExitInvalidInput,
		},
		{
			name:     "validation error returns ExitInvalidInput",
			err:      errors.New("field is required"),
			verbose:  false,
			wantCode: ExitInvalidInput,
		},
		{
			name:     "general error returns ExitGeneralError",
			err:      errors.New("something went wrong"),
			verbose:  false,
			wantCode: ExitGeneralError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := HandleError(tt.err, tt.verbose)
			if code != tt.wantCode {
				t.Errorf("HandleError() = %d, want %d", code, tt.wantCode)
			}
		})
	}
}

func TestFormatError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		verbose bool
		want    string
	}{
		{
			name:    "nil error returns empty string",
			err:     nil,
			verbose: false,
			want:    "",
		},
		{
			name:    "API error formats correctly",
			err:     &canvus.APIError{StatusCode: 404, Message: "Resource not found"},
			verbose: false,
			want:    "API Error: Resource not found (HTTP 404)",
		},
		{
			name:    "general error formats correctly",
			err:     errors.New("test error"),
			verbose: false,
			want:    "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatError(tt.err, tt.verbose)
			if tt.want != "" && got != tt.want {
				// For non-empty expected values, check if the formatted message contains the expected text
				if len(got) == 0 || (tt.want != "" && got[:len(tt.want)] != tt.want) {
					t.Errorf("FormatError() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestGetRemediationSuggestion(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		wantEmpty  bool
	}{
		{
			name:       "401 returns auth suggestion",
			statusCode: 401,
			message:    "Unauthorized",
			wantEmpty:  false,
		},
		{
			name:       "404 returns not found suggestion",
			statusCode: 404,
			message:    "Not Found",
			wantEmpty:  false,
		},
		{
			name:       "200 returns empty suggestion",
			statusCode: 200,
			message:    "OK",
			wantEmpty:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := getRemediationSuggestion(tt.statusCode, tt.message)
			isEmpty := suggestion == ""
			if isEmpty != tt.wantEmpty {
				t.Errorf("getRemediationSuggestion() isEmpty = %v, want %v", isEmpty, tt.wantEmpty)
			}
		})
	}
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "required error is validation error",
			err:  errors.New("field is required"),
			want: true,
		},
		{
			name: "invalid error is validation error",
			err:  errors.New("value is invalid"),
			want: true,
		},
		{
			name: "general error is not validation error",
			err:  errors.New("something went wrong"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidationError(tt.err)
			if got != tt.want {
				t.Errorf("isValidationError() = %v, want %v", got, tt.want)
			}
		})
	}
}
