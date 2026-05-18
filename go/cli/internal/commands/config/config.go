// Package config provides configuration management commands for the Canvus CLI.
package config

import (
	"github.com/spf13/cobra"
)

// ConfigCmd is the parent command for all config operations
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long: `Manage Canvus CLI configuration settings.

Configuration is stored in ~/.canvus/config.yaml and can be set via:
  - Command-line flags (highest priority)
  - Environment variables (CANVUS_*)
  - Configuration file (lowest priority)

Available settings:
  - url:      Canvus server URL
  - api_key:  API key for authentication
  - username: Username for login authentication
  - password: Password for login authentication
  - insecure: Skip TLS certificate verification (true/false)
  - output:   Default output format (json, yaml, table, text)
  - timeout:  Request timeout in seconds

Sensitive values (api_key, password) are masked when displayed.`,
}

func init() {
	// Add subcommands
	ConfigCmd.AddCommand(getCmd)
	ConfigCmd.AddCommand(setCmd)
	ConfigCmd.AddCommand(listCmd)
}
