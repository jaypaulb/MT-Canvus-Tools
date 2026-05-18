package note

import (
	"context"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <canvas-id> <note-id>",
	Short: "Get note details",
	Args:  cobra.ExactArgs(2),
	RunE:  runGet,
}

func init() {
	NoteCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	canvasID, noteID := args[0], args[1]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	note, err := sess.GetNote(context.Background(), canvasID, noteID)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), note, "")
}
