package output

import (
	"testing"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		wantType string
		wantErr  bool
	}{
		{
			name:     "json formatter",
			format:   "json",
			wantType: "*output.JSONFormatter",
			wantErr:  false,
		},
		{
			name:     "yaml formatter",
			format:   "yaml",
			wantType: "*output.YAMLFormatter",
			wantErr:  false,
		},
		{
			name:     "table formatter",
			format:   "table",
			wantType: "*output.TableFormatter",
			wantErr:  false,
		},
		{
			name:     "text formatter",
			format:   "text",
			wantType: "*output.TextFormatter",
			wantErr:  false,
		},
		{
			name:     "unsupported format",
			format:   "xml",
			wantType: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := NewFormatter(tt.format)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewFormatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if formatter == nil {
					t.Errorf("NewFormatter() returned nil formatter")
					return
				}

				// Verify formatter implements the Formatter interface
				var _ Formatter = formatter

				// Test that formatter can actually format data
				_, err := formatter.Format(map[string]string{"test": "value"})
				if err != nil {
					t.Errorf("Formatter.Format() failed: %v", err)
				}
			}
		})
	}
}
