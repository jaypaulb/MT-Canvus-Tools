package config

import (
	"os"
	"testing"
)

func TestLoad_fromEnv(t *testing.T) {
	t.Setenv("CANVUS_API_URL", "https://test.example.com/api/v1")
	t.Setenv("CANVUS_API_KEY", "test-api-key")

	cfg := Load()
	if cfg.CanvusURL != "https://test.example.com/api/v1" {
		t.Errorf("CanvusURL: got %q, want %q", cfg.CanvusURL, "https://test.example.com/api/v1")
	}
	if cfg.APIKey != "test-api-key" {
		t.Errorf("APIKey: got %q, want %q", cfg.APIKey, "test-api-key")
	}
}

func TestLoad_missingEnv(t *testing.T) {
	os.Unsetenv("CANVUS_API_URL")
	os.Unsetenv("CANVUS_API_KEY")

	cfg := Load()
	if cfg.CanvusURL != "" {
		t.Errorf("expected empty CanvusURL, got %q", cfg.CanvusURL)
	}
	if cfg.APIKey != "" {
		t.Errorf("expected empty APIKey, got %q", cfg.APIKey)
	}
}

func TestSetAndGet(t *testing.T) {
	Set(&Config{
		CanvusURL: "https://set.example.com/api/v1",
		APIKey:    "set-key",
	})
	got := Get()
	if got.CanvusURL != "https://set.example.com/api/v1" {
		t.Errorf("CanvusURL: got %q", got.CanvusURL)
	}
	if got.APIKey != "set-key" {
		t.Errorf("APIKey: got %q", got.APIKey)
	}
}

func TestGet_returnsSnapshot(t *testing.T) {
	Set(&Config{CanvusURL: "original", APIKey: "original-key"})
	snap := Get()
	Set(&Config{CanvusURL: "changed", APIKey: "changed-key"})
	// Snapshot must not reflect subsequent Set.
	if snap.CanvusURL != "original" {
		t.Errorf("snapshot.CanvusURL mutated: got %q", snap.CanvusURL)
	}
}
