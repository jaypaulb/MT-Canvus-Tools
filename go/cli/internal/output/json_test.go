package output

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONFormatter_Format(t *testing.T) {
	formatter := &JSONFormatter{}

	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
		check   func(string) bool // custom validation function
	}{
		{
			name: "single object",
			data: map[string]interface{}{
				"id":   "123",
				"name": "Test Canvas",
				"type": "canvas",
			},
			wantErr: false,
			check: func(output string) bool {
				// Should be valid JSON
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					return false
				}
				// Should contain expected fields
				return result["id"] == "123" && result["name"] == "Test Canvas"
			},
		},
		{
			name: "array of objects",
			data: []map[string]interface{}{
				{"id": "1", "name": "First"},
				{"id": "2", "name": "Second"},
			},
			wantErr: false,
			check: func(output string) bool {
				// Should be valid JSON array
				var result []map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					return false
				}
				return len(result) == 2 && result[0]["id"] == "1"
			},
		},
		{
			name: "nested structures",
			data: map[string]interface{}{
				"canvas": map[string]interface{}{
					"id":   "123",
					"meta": map[string]interface{}{
						"created": "2024-01-01",
						"author":  "user1",
					},
				},
			},
			wantErr: false,
			check: func(output string) bool {
				// Should be valid JSON with proper nesting
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					return false
				}
				canvas, ok := result["canvas"].(map[string]interface{})
				if !ok {
					return false
				}
				meta, ok := canvas["meta"].(map[string]interface{})
				return ok && meta["author"] == "user1"
			},
		},
		{
			name:    "nil data returns empty object",
			data:    nil,
			wantErr: false,
			check: func(output string) bool {
				return output == "{}"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := formatter.Format(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONFormatter.Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check that output is properly indented with 2 spaces
				if !strings.Contains(result, "  ") && tt.data != nil {
					t.Errorf("JSONFormatter.Format() output is not indented with 2 spaces")
				}

				// Run custom validation
				if tt.check != nil && !tt.check(result) {
					t.Errorf("JSONFormatter.Format() output failed custom validation: %s", result)
				}
			}
		})
	}
}
