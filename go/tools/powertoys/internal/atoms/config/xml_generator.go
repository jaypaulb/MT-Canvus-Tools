package config

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

// XMLGenerator handles XML file generation.
type XMLGenerator struct{}

// NewXMLGenerator creates a new XML generator.
func NewXMLGenerator() *XMLGenerator {
	return &XMLGenerator{}
}

// Write writes a value to an XML file with proper formatting.
func (g *XMLGenerator) Write(filePath string, v interface{}) error {
	data, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal XML: %w", err)
	}

	// Add XML header
	xmlData := []byte(xml.Header)
	xmlData = append(xmlData, data...)

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, xmlData, 0644); err != nil {
		return fmt.Errorf("failed to write XML file: %w", err)
	}

	return nil
}

// Read reads an XML file and unmarshals it into the provided value.
func (g *XMLGenerator) Read(filePath string, v interface{}) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read XML file: %w", err)
	}

	if err := xml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	return nil
}
