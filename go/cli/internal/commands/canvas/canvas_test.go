package canvas

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCanvasCmd_Registration(t *testing.T) {
	// Test that CanvasCmd is properly initialized
	if CanvasCmd == nil {
		t.Fatal("CanvasCmd should not be nil")
	}

	if CanvasCmd.Use != "canvas" {
		t.Errorf("expected Use to be 'canvas', got '%s'", CanvasCmd.Use)
	}

	if CanvasCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if CanvasCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestCanvasCmd_HelpText(t *testing.T) {
	// Create a root command and add CanvasCmd to it
	rootCmd := &cobra.Command{Use: "canvus"}
	rootCmd.AddCommand(CanvasCmd)

	// Test that help text is available
	help := CanvasCmd.Long
	if len(help) < 50 {
		t.Error("Help text should be descriptive (at least 50 characters)")
	}

	// Verify help text mentions key operations
	keywords := []string{"canvas", "create", "list", "manage"}
	for _, keyword := range keywords {
		if !containsIgnoreCase(help, keyword) {
			t.Errorf("Help text should mention '%s'", keyword)
		}
	}
}

func TestValidateCanvasID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid ID",
			id:      "canvas-123",
			wantErr: false,
		},
		{
			name:    "empty ID",
			id:      "",
			wantErr: true,
		},
		{
			name:    "whitespace ID",
			id:      "   ",
			wantErr: false, // trimming is not done in validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCanvasID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCanvasID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCanvasName(t *testing.T) {
	tests := []struct {
		name       string
		canvasName string
		wantErr    bool
	}{
		{
			name:       "valid name",
			canvasName: "My Canvas",
			wantErr:    false,
		},
		{
			name:       "empty name",
			canvasName: "",
			wantErr:    true,
		},
		{
			name:       "long name",
			canvasName: "This is a very long canvas name that should still be valid",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCanvasName(tt.canvasName)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCanvasName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}
