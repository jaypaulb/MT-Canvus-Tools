package workspace

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <client-id>",
	Short: "Update workspace",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
}

var (
	updateIndex     int
	updateName      string
	updateUser      string
	updateInfoPanel bool
	updatePinned    bool
	updateViewRect  string
)

func init() {
	updateCmd.Flags().IntVarP(&updateIndex, "index", "i", 0, "Workspace index")
	updateCmd.Flags().StringVarP(&updateName, "name", "n", "", "Workspace name")
	updateCmd.Flags().StringVarP(&updateUser, "user", "u", "", "Workspace user")
	updateCmd.Flags().BoolVar(&updateInfoPanel, "info-panel", false, "Show info panel")
	updateCmd.Flags().BoolVar(&updatePinned, "pinned", false, "Pin workspace")
	updateCmd.Flags().StringVar(&updateViewRect, "view-rect", "", "View rectangle as JSON")
	WorkspaceCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	clientID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	selector := canvus.WorkspaceSelector{}
	if cmd.Flags().Changed("index") {
		selector.Index = &updateIndex
	}
	if cmd.Flags().Changed("name") {
		selector.Name = &updateName
	}
	if cmd.Flags().Changed("user") {
		selector.User = &updateUser
	}

	req := canvus.UpdateWorkspaceRequest{}
	if cmd.Flags().Changed("info-panel") {
		req.InfoPanelVisible = &updateInfoPanel
	}
	if cmd.Flags().Changed("pinned") {
		req.Pinned = &updatePinned
	}
	if updateViewRect != "" {
		var rect canvus.Rectangle
		if err := json.Unmarshal([]byte(updateViewRect), &rect); err != nil {
			return fmt.Errorf("failed to parse view-rect JSON: %w", err)
		}
		req.ViewRectangle = &rect
	}

	ws, err := sess.UpdateWorkspace(context.Background(), clientID, selector, req)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), ws, "")
}
