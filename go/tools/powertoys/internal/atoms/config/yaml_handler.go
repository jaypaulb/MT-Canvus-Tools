package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// YAMLHandler handles YAML file parsing and writing.
type YAMLHandler struct{}

// NewYAMLHandler creates a new YAML handler.
func NewYAMLHandler() *YAMLHandler {
	return &YAMLHandler{}
}

// Read reads a YAML file and unmarshals it into the provided value.
func (h *YAMLHandler) Read(filePath string, v interface{}) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	if err := yaml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return nil
}

// Write writes a value to a YAML file.
func (h *YAMLHandler) Write(filePath string, v interface{}) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}

	return nil
}
