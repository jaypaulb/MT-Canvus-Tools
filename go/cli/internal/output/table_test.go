package output

import (
	"strings"
	"testing"
)

func TestTableFormatter_Format(t *testing.T) {
	tests := []struct {
		name      string
		formatter *TableFormatter
		data      interface{}
		wantErr   bool
		check     func(string) bool
	}{
		{
			name:      "nil data returns empty string",
			formatter: &TableFormatter{},
			data:      nil,
			wantErr:   false,
			check: func(output string) bool {
				return output == ""
			},
		},
		{
			name:      "empty slice returns empty string",
			formatter: &TableFormatter{},
			data:      []map[string]string{},
			wantErr:   false,
			check: func(output string) bool {
				return output == ""
			},
		},
		{
			name:      "single object as key-value pairs",
			formatter: &TableFormatter{},
			data: map[string]string{
				"id":   "123",
				"name": "Test Canvas",
			},
			wantErr: false,
			check: func(output string) bool {
				// Should contain key-value format
				return strings.Contains(output, "id:") &&
					strings.Contains(output, "123") &&
					strings.Contains(output, "name:") &&
					strings.Contains(output, "Test Canvas")
			},
		},
		{
			name:      "multiple items as table",
			formatter: &TableFormatter{},
			data: []map[string]string{
				{"id": "1", "name": "First"},
				{"id": "2", "name": "Second"},
				{"id": "3", "name": "Third"},
			},
			wantErr: false,
			check: func(output string) bool {
				// Should contain header row
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) < 4 { // header + separator + 3 rows
					return false
				}
				// Should contain all data
				return strings.Contains(output, "First") &&
					strings.Contains(output, "Second") &&
					strings.Contains(output, "Third")
			},
		},
		{
			name: "custom columns",
			formatter: &TableFormatter{
				Columns: []string{"name", "id"},
			},
			data: []map[string]string{
				{"id": "1", "name": "First", "extra": "ignore"},
				{"id": "2", "name": "Second", "extra": "ignore"},
			},
			wantErr: false,
			check: func(output string) bool {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				// First line should be header with our custom order
				header := lines[0]
				nameIdx := strings.Index(header, "name")
				idIdx := strings.Index(header, "id")
				// name should come before id in the output
				return nameIdx >= 0 && idIdx > nameIdx &&
					!strings.Contains(output, "extra")
			},
		},
		{
			name:      "table has header separator",
			formatter: &TableFormatter{},
			data: []map[string]string{
				{"id": "1", "name": "Test"},
			},
			wantErr: false,
			check: func(output string) bool {
				// Should have a separator line with dashes
				return strings.Contains(output, "---")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.formatter.Format(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("TableFormatter.Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("TableFormatter.Format() output failed validation:\n%s", result)
			}
		})
	}
}
