package workspace

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open <client-id>",
	Short: "Open canvas on workspace",
	Long:  "Open a canvas on one explicit workspace. Supply exactly one of --index, --name or --user; use --index 0 to select workspace zero.",
	Args:  cobra.ExactArgs(1),
	RunE:  runOpen,
}

var (
	openIndex    int
	openName     string
	openUser     string
	openCanvasID string
)

func init() {
	openCmd.Flags().IntVarP(&openIndex, "index", "i", 0, "Workspace index")
	openCmd.Flags().StringVarP(&openName, "name", "n", "", "Workspace name")
	openCmd.Flags().StringVarP(&openUser, "user", "u", "", "Workspace user")
	openCmd.Flags().StringVar(&openCanvasID, "canvas-id", "", "Canvas ID to open")
	openCmd.MarkFlagRequired("canvas-id")
	openCmd.MarkFlagsOneRequired("index", "name", "user")
	openCmd.MarkFlagsMutuallyExclusive("index", "name", "user")
	WorkspaceCmd.AddCommand(openCmd)
}

func runOpen(cmd *cobra.Command, args []string) error {
	clientID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	selector := canvus.WorkspaceSelector{}
	if cmd.Flags().Changed("index") {
		selector.Index = &openIndex
	}
	if cmd.Flags().Changed("name") {
		selector.Name = &openName
	}
	if cmd.Flags().Changed("user") {
		selector.User = &openUser
	}

	opts := canvus.OpenCanvasOptions{
		CanvasID: openCanvasID,
	}

	err = sess.OpenCanvasOnWorkspace(context.Background(), clientID, selector, opts)
	if err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Canvas opened successfully")
	return nil
}
