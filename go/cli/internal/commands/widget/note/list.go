package note

import (
	"context"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <canvas-id>",
	Short: "List note widgets",
	Args:  cobra.ExactArgs(1),
	RunE:  runList,
}

func init() {
	NoteCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	canvasID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	notes, err := sess.ListNotes(context.Background(), canvasID)
	if err != nil {
		return err
	}

	items := make([]interface{}, len(notes))
	for i, note := range notes {
		items[i] = note
	}
	return output.OutputList(cmd.Context(), items, "")
}
