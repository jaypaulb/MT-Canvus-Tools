// Package config manages runtime configuration for note-mapper.
//
// Credentials (CANVUS_API_URL + CANVUS_API_KEY) are supplied via environment
// variables and loaded once at startup. The Canvus canvas ID and Google Gemini
// key are also env-driven, following the Phase 4a CANVUS_* naming convention.
//
// The handler layer additionally accepts in-session overrides via
// POST /api/set-credentials (server URL and API key only), which is useful
// when operating the web UI against multiple Canvus servers without restarting
// the process.
package config

import (
	"os"
	"sync"
)

// Config holds application-level settings.
type Config struct {
	// CanvusURL is the base API URL, e.g. https://canvus.example.com/api/v1.
	CanvusURL string
	// APIKey is the Canvus API key (Private-Token).
	APIKey string
}

var (
	mu      sync.RWMutex
	current = &Config{}
)

// Load populates config from environment variables.
// CANVUS_API_URL maps to CanvusURL; CANVUS_API_KEY maps to APIKey.
// Returns the loaded config (may have empty fields if env vars are unset).
func Load() *Config {
	mu.Lock()
	defer mu.Unlock()
	current = &Config{
		CanvusURL: os.Getenv("CANVUS_API_URL"),
		APIKey:    os.Getenv("CANVUS_API_KEY"),
	}
	return &Config{
		CanvusURL: current.CanvusURL,
		APIKey:    current.APIKey,
	}
}

// Get returns a snapshot of the current config.
func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()
	return &Config{
		CanvusURL: current.CanvusURL,
		APIKey:    current.APIKey,
	}
}

// Set replaces the in-memory config. Used by the /api/set-credentials handler.
func Set(cfg *Config) {
	mu.Lock()
	defer mu.Unlock()
	current = &Config{
		CanvusURL: cfg.CanvusURL,
		APIKey:    cfg.APIKey,
	}
}
