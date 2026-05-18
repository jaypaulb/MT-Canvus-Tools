package output

import (
	"strings"
	"testing"
)

func TestTextFormatter_Format(t *testing.T) {
	tests := []struct {
		name      string
		formatter *TextFormatter
		data      interface{}
		wantErr   bool
		check     func(string) bool
	}{
		{
			name:      "nil data returns empty string",
			formatter: &TextFormatter{},
			data:      nil,
			wantErr:   false,
			check: func(output string) bool {
				return output == ""
			},
		},
		{
			name:      "empty slice returns empty string",
			formatter: &TextFormatter{},
			data:      []string{},
			wantErr:   false,
			check: func(output string) bool {
				return output == ""
			},
		},
		{
			name:      "single object as key-value pairs",
			formatter: &TextFormatter{},
			data: map[string]string{
				"id":   "123",
				"name": "Test Canvas",
			},
			wantErr: false,
			check: func(output string) bool {
				// Should contain key-value format with colons
				return strings.Contains(output, "id: 123") &&
					strings.Contains(output, "name: Test Canvas")
			},
		},
		{
			name:      "list with id field",
			formatter: &TextFormatter{},
			data: []map[string]string{
				{"id": "1", "name": "First"},
				{"id": "2", "name": "Second"},
				{"id": "3", "name": "Third"},
			},
			wantErr: false,
			check: func(output string) bool {
				// Should output newline-separated IDs (default field)
				lines := strings.Split(output, "\n")
				return len(lines) == 3 &&
					strings.Contains(output, "1") &&
					strings.Contains(output, "2") &&
					strings.Contains(output, "3")
			},
		},
		{
			name: "list with custom fields",
			formatter: &TextFormatter{
				Fields: []string{"name", "id"},
			},
			data: []map[string]string{
				{"id": "1", "name": "First"},
				{"id": "2", "name": "Second"},
			},
			wantErr: false,
			check: func(output string) bool {
				// Should output specified fields space-separated
				lines := strings.Split(output, "\n")
				return len(lines) == 2 &&
					strings.Contains(lines[0], "First") &&
					strings.Contains(lines[0], "1") &&
					strings.Contains(lines[1], "Second") &&
					strings.Contains(lines[1], "2")
			},
		},
		{
			name:      "simple string value",
			formatter: &TextFormatter{},
			data:      "simple value",
			wantErr:   false,
			check: func(output string) bool {
				return output == "simple value"
			},
		},
		{
			name:      "list without id falls back to first field",
			formatter: &TextFormatter{},
			data: []map[string]string{
				{"title": "First Item"},
				{"title": "Second Item"},
			},
			wantErr: false,
			check: func(output string) bool {
				// Should output newline-separated values
				lines := strings.Split(output, "\n")
				return len(lines) == 2
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.formatter.Format(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("TextFormatter.Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("TextFormatter.Format() output failed validation:\n%s", result)
			}
		})
	}
}
