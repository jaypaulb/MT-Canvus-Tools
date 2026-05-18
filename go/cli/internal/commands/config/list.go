package config

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration settings",
	Long: `List all configuration settings from ~/.canvus/config.yaml.

Sensitive values (api_key, password) are masked for security.

Example:
  canvus config list`,
	Args: cobra.NoArgs,
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	// Get config file path
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

	// Get all settings
	settings := v.AllSettings()

	if len(settings) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No configuration settings found.")
		return nil
	}

	// Display settings in table format
	fmt.Fprintf(cmd.OutOrStdout(), "Configuration file: %s\n\n", configPath)

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "KEY\tVALUE")
	fmt.Fprintln(w, "---\t-----")

	// Known keys in preferred order
	orderedKeys := []string{
		"url",
		"api_key",
		"username",
		"password",
		"insecure",
		"output",
		"timeout",
	}

	for _, key := range orderedKeys {
		if value, exists := settings[key]; exists {
			displayValue := fmt.Sprintf("%v", value)
			if isSensitiveKey(key) {
				displayValue = maskValue(displayValue)
			}
			fmt.Fprintf(w, "%s\t%s\n", key, displayValue)
		}
	}

	// Show any other keys that might exist
	for key, value := range settings {
		found := false
		for _, k := range orderedKeys {
			if k == key {
				found = true
				break
			}
		}
		if !found {
			displayValue := fmt.Sprintf("%v", value)
			fmt.Fprintf(w, "%s\t%s\n", key, displayValue)
		}
	}

	w.Flush()

	return nil
}
