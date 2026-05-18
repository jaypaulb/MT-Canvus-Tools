// Package config provides configuration management for the Canvus CLI.
// It supports loading configuration from command-line flags, environment
// variables, and a configuration file with proper precedence handling.
package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// contextKey is a type for context keys to avoid collisions
type contextKey int

const (
	configKey contextKey = iota
)

// Config holds all configuration settings for the CLI.
type Config struct {
	// URL is the Canvus server URL
	URL string

	// APIKey is the API key for authentication
	APIKey string

	// Username for username/password authentication
	Username string

	// Password for username/password authentication
	Password string

	// Insecure skips TLS certificate verification when true
	Insecure bool

	// Output format (json, yaml, table, text)
	Output string

	// Timeout in seconds for API requests
	Timeout int

	// Verbose enables detailed logging
	Verbose bool
}

// Load reads configuration from all sources (flags, env, config file)
// and returns a validated Config struct.
// Configuration precedence: flags > environment variables > config file > defaults
func Load() (*Config, error) {
	// Set default values
	viper.SetDefault("output", "table")
	viper.SetDefault("timeout", 30)
	viper.SetDefault("insecure", false)
	viper.SetDefault("verbose", false)

	// Set up config file path
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to determine config path: %w", err)
	}

	// Check if config file exists before trying to read it
	if _, err := os.Stat(configPath); err == nil {
		// File exists, try to read it
		viper.SetConfigFile(configPath)
		viper.SetConfigType("yaml")

		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}
	// If file doesn't exist, that's fine - we'll just use defaults and env vars/flags

	// Build Config struct from Viper
	cfg := &Config{
		URL:      viper.GetString("url"),
		APIKey:   viper.GetString("api_key"),
		Username: viper.GetString("username"),
		Password: viper.GetString("password"),
		Insecure: viper.GetBool("insecure"),
		Output:   viper.GetString("output"),
		Timeout:  viper.GetInt("timeout"),
		Verbose:  viper.GetBool("verbose"),
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that required configuration is present and valid.
func (c *Config) Validate() error {
	// URL is required
	if c.URL == "" {
		return fmt.Errorf("server URL is required (set via --url flag, CANVUS_URL env var, or url in config file)")
	}

	// Either API key or username/password is required
	hasAPIKey := c.APIKey != ""
	hasCredentials := c.Username != "" && c.Password != ""

	if !hasAPIKey && !hasCredentials {
		return fmt.Errorf("authentication is required: either --api-key or both --username and --password must be provided (or use CANVUS_API_KEY/CANVUS_USERNAME/CANVUS_PASSWORD env vars, or set in config file)")
	}

	// Validate output format
	validFormats := map[string]bool{
		"json":  true,
		"yaml":  true,
		"table": true,
		"text":  true,
	}
	if !validFormats[c.Output] {
		return fmt.Errorf("invalid output format '%s': must be one of json, yaml, table, text", c.Output)
	}

	return nil
}

// Save writes the configuration to the config file.
func Save(cfg *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("failed to determine config path: %w", err)
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getConfigPath returns the path to the config file.
func getConfigPath() (string, error) {
	// Get user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".canvus", "config.yaml"), nil
}

// GetConfig retrieves the Config from the context.
func GetConfig(ctx context.Context) (*Config, error) {
	cfg, ok := ctx.Value(configKey).(*Config)
	if !ok || cfg == nil {
		return nil, fmt.Errorf("configuration not found in context")
	}
	return cfg, nil
}

// WithConfig returns a new context with the Config attached.
func WithConfig(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, configKey, cfg)
}
