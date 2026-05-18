package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set the value of a specific configuration setting.

The value will be saved to ~/.canvus/config.yaml.
The configuration file will be created if it doesn't exist.

Examples:
  # Set server URL
  canvus config set url https://canvus.example.com/api/public/v1

  # Set output format
  canvus config set output json

  # Set API key
  canvus config set api_key your-api-key-here

  # Set timeout
  canvus config set timeout 60

  # Set insecure mode
  canvus config set insecure true`,
	Args: cobra.ExactArgs(2),
	RunE: runSet,
}

func runSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	valueStr := args[1]

	// Validate key is a recognized config field
	validKeys := map[string]string{
		"url":      "string",
		"api_key":  "string",
		"username": "string",
		"password": "string",
		"insecure": "bool",
		"output":   "string",
		"timeout":  "int",
	}

	valueType, valid := validKeys[key]
	if !valid {
		return fmt.Errorf("unknown configuration key '%s': valid keys are url, api_key, username, password, insecure, output, timeout", key)
	}

	// Parse and validate value based on type
	var value interface{}
	var err error

	switch valueType {
	case "string":
		value = valueStr
		// Additional validation for specific string keys
		if key == "output" {
			validFormats := map[string]bool{
				"json":  true,
				"yaml":  true,
				"table": true,
				"text":  true,
			}
			if !validFormats[valueStr] {
				return fmt.Errorf("invalid output format '%s': must be one of json, yaml, table, text", valueStr)
			}
		}
	case "bool":
		value, err = strconv.ParseBool(valueStr)
		if err != nil {
			return fmt.Errorf("invalid boolean value '%s': must be true or false", valueStr)
		}
	case "int":
		value, err = strconv.Atoi(valueStr)
		if err != nil {
			return fmt.Errorf("invalid integer value '%s'", valueStr)
		}
		// Validate timeout is positive
		if key == "timeout" && value.(int) <= 0 {
			return fmt.Errorf("timeout must be a positive integer")
		}
	}

	// Get config file path
	configPath, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Load existing config or create new
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Try to read existing config, ignore error if file doesn't exist
	_ = v.ReadInConfig()

	// Set the value
	v.Set(key, value)

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file
	if err := v.WriteConfig(); err != nil {
		// If file doesn't exist, SafeWriteConfig will fail, so we try WriteConfigAs
		if err := v.WriteConfigAs(configPath); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
	}

	// Display confirmation
	displayValue := value
	if isSensitiveKey(key) {
		displayValue = maskValue(fmt.Sprintf("%v", value))
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Configuration updated: %s = %v\n", key, displayValue)
	fmt.Fprintf(cmd.OutOrStdout(), "Saved to: %s\n", configPath)

	return nil
}
