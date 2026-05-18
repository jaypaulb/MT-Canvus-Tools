package pdf

import (
	"context"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <canvas-id> <pdf-id>",
	Short: "Get pdf details",
	Args:  cobra.ExactArgs(2),
	RunE:  runGet,
}

func init() {
	PDFCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	canvasID, itemID := args[0], args[1]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	item, err := sess.GetPDF(context.Background(), canvasID, itemID)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), item, "")
}
