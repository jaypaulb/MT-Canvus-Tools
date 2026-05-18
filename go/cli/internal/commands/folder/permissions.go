package folder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
)

var getPermissionsCmd = &cobra.Command{
	Use:   "get-permissions <folder-id>",
	Short: "Get folder permissions",
	Long: `Get permission overrides for a folder.

Examples:
  # Get folder permissions
  canvus folder get-permissions folder-123

  # Get permissions as JSON
  canvus folder get-permissions folder-123 --output json`,
	Args: cobra.ExactArgs(1),
	RunE: runGetPermissions,
}

var setPermissionsCmd = &cobra.Command{
	Use:   "set-permissions <folder-id>",
	Short: "Set folder permissions",
	Long: `Set permission overrides for a folder.

The permissions should be provided as a JSON string or file.

Examples:
  # Set permissions from JSON string
  canvus folder set-permissions folder-123 --json '{"editors_can_share":true,"users":[],"groups":[]}'

  # Set permissions from file
  canvus folder set-permissions folder-123 --file permissions.json`,
	Args: cobra.ExactArgs(1),
	RunE: runSetPermissions,
}

var (
	permissionsJSON string
	permissionsFile string
)

func init() {
	setPermissionsCmd.Flags().StringVar(&permissionsJSON, "json", "", "Permissions as JSON string")
	setPermissionsCmd.Flags().StringVar(&permissionsFile, "file", "", "Path to JSON file containing permissions")
	FolderCmd.AddCommand(getPermissionsCmd)
	FolderCmd.AddCommand(setPermissionsCmd)
}

func runGetPermissions(cmd *cobra.Command, args []string) error {
	folderID := args[0]

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	perms, err := sess.GetFolderPermissions(context.Background(), folderID)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), perms, "")
}

func runSetPermissions(cmd *cobra.Command, args []string) error {
	folderID := args[0]

	if permissionsJSON == "" && permissionsFile == "" {
		return fmt.Errorf("either --json or --file must be provided")
	}

	var perms canvus.FolderPermissions
	if permissionsFile != "" {
		data, err := os.ReadFile(permissionsFile)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		if err := json.Unmarshal(data, &perms); err != nil {
			return fmt.Errorf("failed to parse JSON from file: %w", err)
		}
	} else {
		if err := json.Unmarshal([]byte(permissionsJSON), &perms); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	updated, err := sess.SetFolderPermissions(context.Background(), folderID, perms)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), updated, "")
}
