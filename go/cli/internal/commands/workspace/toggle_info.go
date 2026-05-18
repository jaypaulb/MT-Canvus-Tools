package workspace

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var toggleInfoCmd = &cobra.Command{
	Use:   "toggle-info <client-id>",
	Short: "Toggle workspace info panel",
	Args:  cobra.ExactArgs(1),
	RunE:  runToggleInfo,
}

var (
	toggleInfoIndex int
	toggleInfoName  string
	toggleInfoUser  string
)

func init() {
	toggleInfoCmd.Flags().IntVarP(&toggleInfoIndex, "index", "i", 0, "Workspace index")
	toggleInfoCmd.Flags().StringVarP(&toggleInfoName, "name", "n", "", "Workspace name")
	toggleInfoCmd.Flags().StringVarP(&toggleInfoUser, "user", "u", "", "Workspace user")
	WorkspaceCmd.AddCommand(toggleInfoCmd)
}

func runToggleInfo(cmd *cobra.Command, args []string) error {
	clientID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	selector := canvus.WorkspaceSelector{}
	if cmd.Flags().Changed("index") {
		selector.Index = &toggleInfoIndex
	}
	if cmd.Flags().Changed("name") {
		selector.Name = &toggleInfoName
	}
	if cmd.Flags().Changed("user") {
		selector.User = &toggleInfoUser
	}

	err = sess.ToggleWorkspaceInfoPanel(context.Background(), clientID, selector)
	if err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Info panel toggled successfully")
	return nil
}
