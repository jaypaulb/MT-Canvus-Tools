package note

import (
	"context"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <canvas-id> <note-id>",
	Short: "Delete note",
	Args:  cobra.ExactArgs(2),
	RunE:  runDelete,
}

func init() {
	NoteCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	canvasID, noteID := args[0], args[1]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	err = sess.DeleteNote(context.Background(), canvasID, noteID)
	if err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Note deleted successfully")
	return nil
}
