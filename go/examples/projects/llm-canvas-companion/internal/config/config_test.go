package config

import (
	"testing"
)

func TestLoad_MissingRequired(t *testing.T) {
	for _, k := range []string{"CANVUS_API_URL", "CANVUS_API_KEY", "CANVAS_ID"} {
		t.Setenv(k, "")
	}

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when required env vars are missing, got nil")
	}
}

func TestLoad_MinimalRequired(t *testing.T) {
	t.Setenv("CANVUS_API_URL", "https://example.com/api/v1/")
	t.Setenv("CANVUS_API_KEY", "test-key")
	t.Setenv("CANVAS_ID", "canvas-123")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CanvusAPIURL != "https://example.com/api/v1/" {
		t.Errorf("CanvusAPIURL: got %q", cfg.CanvusAPIURL)
	}
	if cfg.CanvusAPIKey != "test-key" {
		t.Errorf("CanvusAPIKey: got %q", cfg.CanvusAPIKey)
	}
	if cfg.CanvasID != "canvas-123" {
		t.Errorf("CanvasID: got %q", cfg.CanvasID)
	}
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("CANVUS_API_URL", "https://example.com/api/v1/")
	t.Setenv("CANVUS_API_KEY", "k")
	t.Setenv("CANVAS_ID", "c")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BaseLLMURL != "http://127.0.0.1:1234/v1" {
		t.Errorf("BaseLLMURL default wrong: %q", cfg.BaseLLMURL)
	}
	if cfg.OpenAINoteModel != "gpt-4" {
		t.Errorf("OpenAINoteModel default wrong: %q", cfg.OpenAINoteModel)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("MaxRetries default wrong: %d", cfg.MaxRetries)
	}
	if cfg.SnapshotDrainSeconds != 3 {
		t.Errorf("SnapshotDrainSeconds default wrong: %d", cfg.SnapshotDrainSeconds)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("CANVUS_API_URL", "https://example.com/api/v1/")
	t.Setenv("CANVUS_API_KEY", "k")
	t.Setenv("CANVAS_ID", "c")
	t.Setenv("MAX_RETRIES", "7")
	t.Setenv("SNAPSHOT_DRAIN_SECONDS", "10")
	t.Setenv("OPENAI_NOTE_MODEL", "gpt-3.5-turbo")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxRetries != 7 {
		t.Errorf("MaxRetries: got %d, want 7", cfg.MaxRetries)
	}
	if cfg.SnapshotDrainSeconds != 10 {
		t.Errorf("SnapshotDrainSeconds: got %d, want 10", cfg.SnapshotDrainSeconds)
	}
	if cfg.OpenAINoteModel != "gpt-3.5-turbo" {
		t.Errorf("OpenAINoteModel: got %q", cfg.OpenAINoteModel)
	}
}

func TestLoad_LegacyOpenAIKey(t *testing.T) {
	t.Setenv("CANVUS_API_URL", "https://example.com/api/v1/")
	t.Setenv("CANVUS_API_KEY", "k")
	t.Setenv("CANVAS_ID", "c")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENAI_KEY", "legacy-key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OpenAIAPIKey != "legacy-key" {
		t.Errorf("OpenAIAPIKey from legacy var: got %q", cfg.OpenAIAPIKey)
	}
}
