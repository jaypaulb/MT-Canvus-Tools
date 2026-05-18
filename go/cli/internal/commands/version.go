// Package commands provides the version command for the Canvus CLI.
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information - these should be set by build flags
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long: `Display version information for the Canvus CLI.

Shows:
  - CLI version number
  - Git commit hash
  - Build date

This information is set at build time using ldflags.`,
	Args: cobra.NoArgs,
	Run:  runVersion,
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Fprintf(cmd.OutOrStdout(), "Canvus CLI\n")
	fmt.Fprintf(cmd.OutOrStdout(), "Version:    %s\n", Version)
	fmt.Fprintf(cmd.OutOrStdout(), "Commit:     %s\n", Commit)
	fmt.Fprintf(cmd.OutOrStdout(), "Build Date: %s\n", Date)
}

// GetVersionCmd returns the version command for registration with root command
func GetVersionCmd() *cobra.Command {
	return versionCmd
}
