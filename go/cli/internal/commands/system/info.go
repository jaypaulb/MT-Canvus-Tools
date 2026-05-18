package system

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display system information",
	Long: `Display system information from the Canvus server.

Shows server connectivity status and license information. This command
can be used to verify that the CLI can connect to the Canvus server.

Examples:
  # Get system information
  canvus system info

  # Get system information in JSON format
  canvus system info --output json`,
	RunE: runInfo,
}

func init() {
	SystemCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	// Get SDK session
	sess, err := getSession(cmd)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Get license info as a proxy for system info
	// The SDK doesn't have a dedicated GetSystemInfo method,
	// so we use GetLicenseInfo to verify connectivity and get system details
	ctx := context.Background()
	license, err := sess.GetLicenseInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get system info: %w", err)
	}

	// Build system info response
	systemInfo := map[string]interface{}{
		"status":  "connected",
		"license": license,
	}

	// Output result
	return output.PrintOutput(cmd.Context(), systemInfo, "")
}
