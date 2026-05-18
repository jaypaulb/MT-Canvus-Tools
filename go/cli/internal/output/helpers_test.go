package output

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/config"
)

func TestPrintOutputTo(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		format  string
		cfg     *config.Config
		wantErr bool
		check   func(string) bool
	}{
		{
			name:    "json format explicitly specified",
			data:    map[string]string{"id": "123", "name": "Test"},
			format:  "json",
			cfg:     nil,
			wantErr: false,
			check: func(output string) bool {
				return strings.Contains(output, `"id"`) && strings.Contains(output, `"123"`)
			},
		},
		{
			name:    "yaml format explicitly specified",
			data:    map[string]string{"id": "123", "name": "Test"},
			format:  "yaml",
			cfg:     nil,
			wantErr: false,
			check: func(output string) bool {
				return strings.Contains(output, "id:") && strings.Contains(output, "123")
			},
		},
		{
			name:   "format from config",
			data:   map[string]string{"id": "123"},
			format: "", // empty format should use config
			cfg: &config.Config{
				Output: "json",
			},
			wantErr: false,
			check: func(output string) bool {
				return strings.Contains(output, `"id"`)
			},
		},
		{
			name:    "default to table format when no format and no config",
			data:    []map[string]string{{"id": "1"}},
			format:  "",
			cfg:     nil,
			wantErr: false,
			check: func(output string) bool {
				// Table format should have header
				return strings.Contains(output, "id")
			},
		},
		{
			name:    "invalid format returns error",
			data:    map[string]string{"id": "123"},
			format:  "invalid",
			cfg:     nil,
			wantErr: true,
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.cfg != nil {
				ctx = config.WithConfig(ctx, tt.cfg)
			}

			buf := &bytes.Buffer{}
			err := PrintOutputTo(ctx, buf, tt.data, tt.format)

			if (err != nil) != tt.wantErr {
				t.Errorf("PrintOutputTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				output := buf.String()
				if !tt.check(output) {
					t.Errorf("PrintOutputTo() output failed validation:\n%s", output)
				}
			}
		})
	}
}

func TestPrintErrorTo(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		cfg   *config.Config
		check func(string) bool
	}{
		{
			name: "error in text format",
			err:  errors.New("test error"),
			cfg:  &config.Config{Output: "text"},
			check: func(output string) bool {
				return strings.Contains(output, "Error:") && strings.Contains(output, "test error")
			},
		},
		{
			name: "error in json format",
			err:  errors.New("test error"),
			cfg:  &config.Config{Output: "json"},
			check: func(output string) bool {
				return strings.Contains(output, `"error"`) &&
					strings.Contains(output, "test error") &&
					strings.Contains(output, `"success": false`)
			},
		},
		{
			name: "error in yaml format",
			err:  errors.New("test error"),
			cfg:  &config.Config{Output: "yaml"},
			check: func(output string) bool {
				return strings.Contains(output, "error:") &&
					strings.Contains(output, "test error") &&
					strings.Contains(output, "success: false")
			},
		},
		{
			name: "nil error outputs nothing",
			err:  nil,
			cfg:  nil,
			check: func(output string) bool {
				return output == ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.cfg != nil {
				ctx = config.WithConfig(ctx, tt.cfg)
			}

			buf := &bytes.Buffer{}
			PrintErrorTo(ctx, buf, tt.err)

			if tt.check != nil {
				output := buf.String()
				if !tt.check(output) {
					t.Errorf("PrintErrorTo() output failed validation:\n%s", output)
				}
			}
		})
	}
}

func TestOutputSuccessTo(t *testing.T) {
	tests := []struct {
		name    string
		message string
		cfg     *config.Config
		check   func(string) bool
	}{
		{
			name:    "success in text format",
			message: "Operation completed",
			cfg:     &config.Config{Output: "text"},
			check: func(output string) bool {
				return strings.Contains(output, "Operation completed")
			},
		},
		{
			name:    "success in json format",
			message: "Operation completed",
			cfg:     &config.Config{Output: "json"},
			check: func(output string) bool {
				return strings.Contains(output, `"message"`) &&
					strings.Contains(output, "Operation completed") &&
					strings.Contains(output, `"success": true`)
			},
		},
		{
			name:    "success in yaml format",
			message: "Operation completed",
			cfg:     &config.Config{Output: "yaml"},
			check: func(output string) bool {
				return strings.Contains(output, "message:") &&
					strings.Contains(output, "Operation completed") &&
					strings.Contains(output, "success: true")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.cfg != nil {
				ctx = config.WithConfig(ctx, tt.cfg)
			}

			buf := &bytes.Buffer{}
			OutputSuccessTo(ctx, buf, tt.message)

			if tt.check != nil {
				output := buf.String()
				if !tt.check(output) {
					t.Errorf("OutputSuccessTo() output failed validation:\n%s", output)
				}
			}
		})
	}
}
