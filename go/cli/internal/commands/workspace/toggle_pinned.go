package workspace

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
)

var togglePinnedCmd = &cobra.Command{
	Use:   "toggle-pinned <client-id>",
	Short: "Toggle workspace pinned state",
	Args:  cobra.ExactArgs(1),
	RunE:  runTogglePinned,
}

var (
	togglePinnedIndex int
	togglePinnedName  string
	togglePinnedUser  string
)

func init() {
	togglePinnedCmd.Flags().IntVarP(&togglePinnedIndex, "index", "i", 0, "Workspace index")
	togglePinnedCmd.Flags().StringVarP(&togglePinnedName, "name", "n", "", "Workspace name")
	togglePinnedCmd.Flags().StringVarP(&togglePinnedUser, "user", "u", "", "Workspace user")
	WorkspaceCmd.AddCommand(togglePinnedCmd)
}

func runTogglePinned(cmd *cobra.Command, args []string) error {
	clientID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	selector := canvus.WorkspaceSelector{}
	if cmd.Flags().Changed("index") {
		selector.Index = &togglePinnedIndex
	}
	if cmd.Flags().Changed("name") {
		selector.Name = &togglePinnedName
	}
	if cmd.Flags().Changed("user") {
		selector.User = &togglePinnedUser
	}

	err = sess.ToggleWorkspacePinned(context.Background(), clientID, selector)
	if err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Workspace pinned state toggled successfully")
	return nil
}
