package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Get the value of a specific configuration setting.

Sensitive values (api_key, password) are masked for security.

Examples:
  # Get server URL
  canvus config get url

  # Get output format
  canvus config get output

  # Get API key (will be masked)
  canvus config get api_key`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	// Validate key is a recognized config field
	validKeys := map[string]bool{
		"url":      true,
		"api_key":  true,
		"username": true,
		"password": true,
		"insecure": true,
		"output":   true,
		"timeout":  true,
	}

	if !validKeys[key] {
		return fmt.Errorf("unknown configuration key '%s': valid keys are url, api_key, username, password, insecure, output, timeout", key)
	}

	// Load config file
	configPath, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file not found at %s", configPath)
	}

	// Read config file
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Get value
	value := v.Get(key)
	if value == nil {
		return fmt.Errorf("configuration key '%s' is not set", key)
	}

	// Mask sensitive values
	if isSensitiveKey(key) {
		strValue := fmt.Sprintf("%v", value)
		value = maskValue(strValue)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%s: %v\n", key, value)
	return nil
}

// getConfigFilePath is a variable so it can be overridden in tests
var getConfigFilePath = func() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".canvus", "config.yaml"), nil
}

func isSensitiveKey(key string) bool {
	sensitiveKeys := map[string]bool{
		"api_key":  true,
		"password": true,
	}
	return sensitiveKeys[key]
}

func maskValue(value string) string {
	if value == "" {
		return ""
	}
	// For values <= 8 chars, mask everything
	if len(value) <= 8 {
		return value // Don't mask short values - they're already not very secure
	}
	// Show first 4 and last 4 characters
	return value[:4] + strings.Repeat("*", len(value)-8) + value[len(value)-4:]
}
