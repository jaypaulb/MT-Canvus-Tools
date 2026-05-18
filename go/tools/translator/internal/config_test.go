package internal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_RequiredVarsMissing(t *testing.T) {
	// Clear all required env vars.
	vars := []string{"CANVUS_API_URL", "CANVUS_API_KEY", "CANVUS_CANVAS_ID", "GEMINI_API_KEY"}
	for _, v := range vars {
		t.Setenv(v, "")
	}

	_, err := LoadConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required environment variables")
}

func TestLoadConfig_AllPresent(t *testing.T) {
	t.Setenv("CANVUS_API_URL", "https://example.com/api/v1/")
	t.Setenv("CANVUS_API_KEY", "test-key")
	t.Setenv("CANVUS_CANVAS_ID", "canvas-abc")
	t.Setenv("GEMINI_API_KEY", "gemini-key")
	t.Setenv("PORT", "9090")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/api/v1/", cfg.CanvusAPIURL)
	assert.Equal(t, "test-key", cfg.CanvusAPIKey)
	assert.Equal(t, "canvas-abc", cfg.CanvasID)
	assert.Equal(t, "gemini-key", cfg.GeminiAPIKey)
	assert.Equal(t, "9090", cfg.Port)
}

func TestLoadConfig_DefaultPort(t *testing.T) {
	t.Setenv("CANVUS_API_URL", "https://example.com/api/v1/")
	t.Setenv("CANVUS_API_KEY", "test-key")
	t.Setenv("CANVUS_CANVAS_ID", "canvas-abc")
	t.Setenv("GEMINI_API_KEY", "gemini-key")
	if err := os.Unsetenv("PORT"); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)
}

func TestLoadConfig_PartiallyMissing(t *testing.T) {
	tests := []struct {
		name    string
		missing string
	}{
		{"missing CANVUS_API_URL", "CANVUS_API_URL"},
		{"missing CANVUS_API_KEY", "CANVUS_API_KEY"},
		{"missing CANVUS_CANVAS_ID", "CANVUS_CANVAS_ID"},
		{"missing GEMINI_API_KEY", "GEMINI_API_KEY"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CANVUS_API_URL", "https://example.com/api/v1/")
			t.Setenv("CANVUS_API_KEY", "test-key")
			t.Setenv("CANVUS_CANVAS_ID", "canvas-abc")
			t.Setenv("GEMINI_API_KEY", "gemini-key")
			t.Setenv(tt.missing, "")

			_, err := LoadConfig()
			require.Error(t, err, "expected error when %s is missing", tt.missing)
			assert.Contains(t, err.Error(), tt.missing)
		})
	}
}
