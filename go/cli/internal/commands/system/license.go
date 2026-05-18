package system

import (
	"context"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Display license information",
	Long: `Display license information from the Canvus server.

Shows license details including type, expiration, and capabilities.
Requires appropriate permissions to view license information.

Examples:
  # Get license information
  canvus system license

  # Get license information in JSON format
  canvus system license --output json`,
	RunE: runLicense,
}

func init() {
	SystemCmd.AddCommand(licenseCmd)
}

func runLicense(cmd *cobra.Command, args []string) error {
	// Get SDK session
	sess, err := getSession(cmd)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Get license info via SDK
	ctx := context.Background()
	license, err := sess.GetLicenseInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get license info: %w", err)
	}

	// Output result
	return output.PrintOutput(cmd.Context(), license, "")
}
