package output_test

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
)

// ExampleJSONFormatter demonstrates JSON formatting
func ExampleJSONFormatter() {
	formatter := &output.JSONFormatter{}

	data := map[string]interface{}{
		"id":     "canvas-123",
		"name":   "My Canvas",
		"status": "active",
	}

	result, _ := formatter.Format(data)
	fmt.Println(result)

	// Output:
	// {
	//   "id": "canvas-123",
	//   "name": "My Canvas",
	//   "status": "active"
	// }
}

// ExampleTableFormatter demonstrates table formatting
func ExampleTableFormatter() {
	formatter := &output.TableFormatter{}

	data := []map[string]string{
		{"id": "1", "name": "Canvas One", "status": "active"},
		{"id": "2", "name": "Canvas Two", "status": "inactive"},
	}

	result, _ := formatter.Format(data)
	fmt.Println(result)

	// Output is a formatted table with headers
}

// ExamplePrintOutput demonstrates the high-level helper
func ExamplePrintOutput() {
	ctx := context.Background()
	cfg := &config.Config{
		URL:    "https://example.com",
		APIKey: "test-key",
		Output: "json",
	}
	ctx = config.WithConfig(ctx, cfg)

	data := map[string]string{
		"message": "Operation successful",
	}

	// Format will be taken from config (json)
	output.PrintOutput(ctx, data, "")

	// Output will be JSON formatted
}
