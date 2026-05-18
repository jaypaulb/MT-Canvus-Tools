package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	// Set version info for testing
	Version = "v1.0.0"
	Commit = "abc123"
	Date = "2025-11-20"

	tests := []struct {
		name     string
		wantText []string
	}{
		{
			name: "displays version information",
			wantText: []string{
				"Canvus CLI",
				"Version:    v1.0.0",
				"Commit:     abc123",
				"Build Date: 2025-11-20",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create command
			cmd := GetVersionCmd()

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)

			// Execute command - note that Run doesn't return error
			cmd.Run(cmd, []string{})

			// Check output
			output := buf.String()
			for _, want := range tt.wantText {
				if !strings.Contains(output, want) {
					t.Errorf("expected output to contain %q, got:\n%s", want, output)
				}
			}
		})
	}
}

func TestVersionCommandFormat(t *testing.T) {
	// Set version info
	Version = "dev"
	Commit = "none"
	Date = "unknown"

	cmd := GetVersionCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	cmd.Run(cmd, []string{})

	output := buf.String()

	// Verify all three fields are present
	requiredFields := []string{"Version:", "Commit:", "Build Date:"}
	for _, field := range requiredFields {
		if !strings.Contains(output, field) {
			t.Errorf("expected output to contain %q field", field)
		}
	}
}
