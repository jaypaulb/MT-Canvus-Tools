package canvas

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var getPermissionsCmd = &cobra.Command{
	Use:   "get-permissions <canvas-id>",
	Short: "Get canvas permissions",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetPermissions,
}

var setPermissionsCmd = &cobra.Command{
	Use:   "set-permissions <canvas-id>",
	Short: "Set canvas permissions",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetPermissions,
}

var (
	permissionsJSON string
	permissionsFile string
)

func init() {
	setPermissionsCmd.Flags().StringVar(&permissionsJSON, "json", "", "Permissions as JSON")
	setPermissionsCmd.Flags().StringVar(&permissionsFile, "file", "", "Path to JSON file")
	CanvasCmd.AddCommand(getPermissionsCmd)
	CanvasCmd.AddCommand(setPermissionsCmd)
}

func runGetPermissions(cmd *cobra.Command, args []string) error {
	canvasID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	perms, err := sess.GetCanvasPermissions(context.Background(), canvasID)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), perms, "")
}

func runSetPermissions(cmd *cobra.Command, args []string) error {
	canvasID := args[0]

	if permissionsJSON == "" && permissionsFile == "" {
		return fmt.Errorf("either --json or --file must be provided")
	}

	var perms canvus.CanvasPermissions
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

	updated, err := sess.SetCanvasPermissions(context.Background(), canvasID, perms)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), updated, "")
}
