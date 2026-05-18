package output

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// YAMLFormatter formats data as YAML.
type YAMLFormatter struct{}

// Format converts the input data to YAML format.
// Supports both single objects and arrays.
func (f *YAMLFormatter) Format(data interface{}) (string, error) {
	if data == nil {
		return "{}", nil
	}

	// Marshal to YAML
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data to YAML: %w", err)
	}

	return string(yamlBytes), nil
}
