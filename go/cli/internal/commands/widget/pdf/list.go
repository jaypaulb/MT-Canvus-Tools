package pdf

import (
	"context"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <canvas-id>",
	Short: "List pdf widgets",
	Args:  cobra.ExactArgs(1),
	RunE:  runList,
}

func init() {
	PDFCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	canvasID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	items, err := sess.ListPDFs(context.Background(), canvasID)
	if err != nil {
		return err
	}

	list := make([]interface{}, len(items))
	for i, item := range items {
		list[i] = item
	}
	return output.OutputList(cmd.Context(), list, "")
}
