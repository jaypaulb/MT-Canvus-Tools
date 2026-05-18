package output

import (
	"encoding/json"
	"fmt"
)

// JSONFormatter formats data as pretty-printed JSON with 2-space indentation.
type JSONFormatter struct{}

// Format converts the input data to pretty-printed JSON.
// Supports both single objects and slices of objects.
func (f *JSONFormatter) Format(data interface{}) (string, error) {
	if data == nil {
		return "{}", nil
	}

	// Marshal with indentation (2 spaces)
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	return string(jsonBytes), nil
}
