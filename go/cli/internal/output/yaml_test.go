package output

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLFormatter_Format(t *testing.T) {
	formatter := &YAMLFormatter{}

	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
		check   func(string) bool
	}{
		{
			name:    "nil data returns empty object",
			data:    nil,
			wantErr: false,
			check: func(output string) bool {
				return output == "{}"
			},
		},
		{
			name: "single object",
			data: map[string]interface{}{
				"id":   "123",
				"name": "Test Canvas",
			},
			wantErr: false,
			check: func(output string) bool {
				var result map[string]interface{}
				if err := yaml.Unmarshal([]byte(output), &result); err != nil {
					return false
				}
				return result["id"] == "123" && result["name"] == "Test Canvas"
			},
		},
		{
			name: "array of objects",
			data: []map[string]string{
				{"id": "1", "name": "First"},
				{"id": "2", "name": "Second"},
			},
			wantErr: false,
			check: func(output string) bool {
				var result []map[string]string
				if err := yaml.Unmarshal([]byte(output), &result); err != nil {
					return false
				}
				return len(result) == 2 && result[0]["id"] == "1"
			},
		},
		{
			name: "nested structures",
			data: map[string]interface{}{
				"canvas": map[string]interface{}{
					"id": "123",
					"meta": map[string]string{
						"created": "2024-01-01",
						"author":  "user1",
					},
				},
			},
			wantErr: false,
			check: func(output string) bool {
				var result map[string]interface{}
				if err := yaml.Unmarshal([]byte(output), &result); err != nil {
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
			name: "empty slice",
			data: []string{},
			wantErr: false,
			check: func(output string) bool {
				return strings.TrimSpace(output) == "[]"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := formatter.Format(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("YAMLFormatter.Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("YAMLFormatter.Format() output failed validation: %s", result)
			}
		})
	}
}
