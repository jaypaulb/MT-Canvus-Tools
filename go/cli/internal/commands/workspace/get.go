package workspace

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <client-id>",
	Short: "Get workspace details",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

var (
	getIndex int
	getName  string
	getUser  string
)

func init() {
	getCmd.Flags().IntVarP(&getIndex, "index", "i", 0, "Workspace index")
	getCmd.Flags().StringVarP(&getName, "name", "n", "", "Workspace name")
	getCmd.Flags().StringVarP(&getUser, "user", "u", "", "Workspace user")
	WorkspaceCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	clientID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	selector := canvus.WorkspaceSelector{}
	if cmd.Flags().Changed("index") {
		selector.Index = &getIndex
	}
	if cmd.Flags().Changed("name") {
		selector.Name = &getName
	}
	if cmd.Flags().Changed("user") {
		selector.User = &getUser
	}

	ws, err := sess.GetWorkspace(context.Background(), clientID, selector)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), ws, "")
}
