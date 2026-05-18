// Package internal contains the implementation packages for the
// canvus-translator tool. It is not importable outside this module.
package internal

import (
	"fmt"
	"os"
)

// Config holds all runtime settings resolved from environment variables.
// The old repo used CANVUS_SERVER; this port follows the Phase 4a
// env-var unification and uses CANVUS_API_URL instead.
type Config struct {
	// CANVUS_API_URL — base URL of the Canvus API server, e.g.
	// https://your-canvus-server.example.com/api/v1/
	CanvusAPIURL string

	// CANVUS_API_KEY — private API token (Private-Token header).
	CanvusAPIKey string

	// CANVUS_CANVAS_ID — ID of the canvas whose notes will be translated.
	CanvasID string

	// GEMINI_API_KEY — Google Gemini API key.
	GeminiAPIKey string

	// PORT — HTTP port the server listens on. Defaults to "8080".
	Port string
}

// LoadConfig resolves configuration from environment variables.
// It returns a non-nil error if any required variable is missing.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		CanvusAPIURL: os.Getenv("CANVUS_API_URL"),
		CanvusAPIKey: os.Getenv("CANVUS_API_KEY"),
		CanvasID:     os.Getenv("CANVUS_CANVAS_ID"),
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		Port:         os.Getenv("PORT"),
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	var missing []string
	if cfg.CanvusAPIURL == "" {
		missing = append(missing, "CANVUS_API_URL")
	}
	if cfg.CanvusAPIKey == "" {
		missing = append(missing, "CANVUS_API_KEY")
	}
	if cfg.CanvasID == "" {
		missing = append(missing, "CANVUS_CANVAS_ID")
	}
	if cfg.GeminiAPIKey == "" {
		missing = append(missing, "GEMINI_API_KEY")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}
	return cfg, nil
}
