package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config with API key",
			config: &Config{
				URL:    "https://canvus.example.com",
				APIKey: "test-api-key",
				Output: "json",
			},
			wantErr: false,
		},
		{
			name: "valid config with username/password",
			config: &Config{
				URL:      "https://canvus.example.com",
				Username: "testuser",
				Password: "testpass",
				Output:   "table",
			},
			wantErr: false,
		},
		{
			name: "missing URL",
			config: &Config{
				APIKey: "test-api-key",
				Output: "json",
			},
			wantErr: true,
		},
		{
			name: "missing authentication",
			config: &Config{
				URL:    "https://canvus.example.com",
				Output: "json",
			},
			wantErr: true,
		},
		{
			name: "invalid output format",
			config: &Config{
				URL:    "https://canvus.example.com",
				APIKey: "test-api-key",
				Output: "invalid",
			},
			wantErr: true,
		},
		{
			name: "missing password",
			config: &Config{
				URL:      "https://canvus.example.com",
				Username: "testuser",
				Output:   "json",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	// Save original state
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Reset viper state before each test
	resetViper := func() {
		viper.Reset()
		viper.SetEnvPrefix("CANVUS")
		viper.AutomaticEnv()
		viper.SetDefault("output", "table")
		viper.SetDefault("timeout", 30)
		viper.SetDefault("insecure", false)
		viper.SetDefault("verbose", false)
	}

	t.Run("load with env vars", func(t *testing.T) {
		// Create temporary directory for this test
		tempDir := t.TempDir()
		os.Setenv("HOME", tempDir)

		resetViper()
		os.Setenv("CANVUS_API_URL", "https://env.example.com")
		os.Setenv("CANVUS_API_KEY", "env-api-key")
		defer os.Unsetenv("CANVUS_API_URL")
		defer os.Unsetenv("CANVUS_API_KEY")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.URL != "https://env.example.com" {
			t.Errorf("expected URL from env, got %s", cfg.URL)
		}
		if cfg.APIKey != "env-api-key" {
			t.Errorf("expected API key from env, got %s", cfg.APIKey)
		}
	})

	t.Run("load with config file", func(t *testing.T) {
		// Create temporary directory for this test
		tempDir := t.TempDir()
		os.Setenv("HOME", tempDir)

		resetViper()
		// Clear env vars
		os.Unsetenv("CANVUS_API_URL")
		os.Unsetenv("CANVUS_API_KEY")

		// Create config directory and file
		configDir := filepath.Join(tempDir, ".canvus")
		os.MkdirAll(configDir, 0755)
		configFile := filepath.Join(configDir, "config.yaml")

		configContent := `url: https://config.example.com
api_key: config-api-key
output: json
timeout: 60
`
		if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.URL != "https://config.example.com" {
			t.Errorf("expected URL from config file, got %s", cfg.URL)
		}
		if cfg.APIKey != "config-api-key" {
			t.Errorf("expected API key from config file, got %s", cfg.APIKey)
		}
		if cfg.Output != "json" {
			t.Errorf("expected output json from config file, got %s", cfg.Output)
		}
		if cfg.Timeout != 60 {
			t.Errorf("expected timeout 60 from config file, got %d", cfg.Timeout)
		}
	})

	t.Run("defaults are applied", func(t *testing.T) {
		// Create temporary directory for this test
		tempDir := t.TempDir()
		os.Setenv("HOME", tempDir)

		resetViper()
		os.Setenv("CANVUS_API_URL", "https://default.example.com")
		os.Setenv("CANVUS_API_KEY", "default-api-key")
		defer os.Unsetenv("CANVUS_API_URL")
		defer os.Unsetenv("CANVUS_API_KEY")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Check defaults
		if cfg.Output != "table" {
			t.Errorf("expected default output table, got %s", cfg.Output)
		}
		if cfg.Timeout != 30 {
			t.Errorf("expected default timeout 30, got %d", cfg.Timeout)
		}
		if cfg.Insecure {
			t.Errorf("expected default insecure false, got true")
		}
	})

	t.Run("validation error on missing config", func(t *testing.T) {
		// Create temporary directory for this test
		tempDir := t.TempDir()
		os.Setenv("HOME", tempDir)

		resetViper()
		os.Unsetenv("CANVUS_API_URL")
		os.Unsetenv("CANVUS_API_KEY")
		os.Unsetenv("CANVUS_USERNAME")
		os.Unsetenv("CANVUS_PASSWORD")

		_, err := Load()
		if err == nil {
			t.Error("expected error for missing config, got nil")
		}
	})

	t.Run("precedence: env over config file", func(t *testing.T) {
		// Create temporary directory for this test
		tempDir := t.TempDir()
		os.Setenv("HOME", tempDir)

		resetViper()

		// Create config file
		configDir := filepath.Join(tempDir, ".canvus")
		os.MkdirAll(configDir, 0755)
		configFile := filepath.Join(configDir, "config.yaml")
		configContent := `url: https://config.example.com
api_key: config-api-key
`
		if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		// Set env vars that should override config file
		os.Setenv("CANVUS_API_URL", "https://env.example.com")
		os.Setenv("CANVUS_API_KEY", "env-api-key")
		defer os.Unsetenv("CANVUS_API_URL")
		defer os.Unsetenv("CANVUS_API_KEY")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Env vars should win
		if cfg.URL != "https://env.example.com" {
			t.Errorf("expected URL from env to override config file, got %s", cfg.URL)
		}
		if cfg.APIKey != "env-api-key" {
			t.Errorf("expected API key from env to override config file, got %s", cfg.APIKey)
		}
	})
}

func TestSave(t *testing.T) {
	// Save original state
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Create temporary directory for test config
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)

	cfg := &Config{
		URL:      "https://save.example.com",
		APIKey:   "save-api-key",
		Output:   "yaml",
		Timeout:  45,
		Insecure: true,
	}

	err := Save(cfg)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tempDir, ".canvus", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("config file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	// Check if content contains expected values
	contentStr := string(content)
	expectedStrings := []string{
		"https://save.example.com",
		"save-api-key",
		"yaml",
	}

	for _, expected := range expectedStrings {
		if !contains(contentStr, expected) {
			t.Errorf("config file missing expected value: %s", expected)
		}
	}
}

func TestWithConfigAndGetConfig(t *testing.T) {
	cfg := &Config{
		URL:    "https://context.example.com",
		APIKey: "context-api-key",
		Output: "json",
	}

	ctx := context.Background()
	ctx = WithConfig(ctx, cfg)

	retrievedCfg, err := GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if retrievedCfg.URL != cfg.URL {
		t.Errorf("expected URL %s, got %s", cfg.URL, retrievedCfg.URL)
	}
	if retrievedCfg.APIKey != cfg.APIKey {
		t.Errorf("expected APIKey %s, got %s", cfg.APIKey, retrievedCfg.APIKey)
	}
}

func TestGetConfigMissing(t *testing.T) {
	ctx := context.Background()
	_, err := GetConfig(ctx)
	if err == nil {
		t.Error("expected error for missing config in context, got nil")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
