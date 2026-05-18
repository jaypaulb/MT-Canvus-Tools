package note

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <canvas-id> <note-id>",
	Short: "Update note",
	Args:  cobra.ExactArgs(2),
	RunE:  runUpdate,
}

var updateJSON string

func init() {
	updateCmd.Flags().StringVar(&updateJSON, "json", "", "Update data as JSON")
	updateCmd.MarkFlagRequired("json")
	NoteCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	canvasID, noteID := args[0], args[1]

	var req interface{}
	if err := json.Unmarshal([]byte(updateJSON), &req); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	note, err := sess.UpdateNote(context.Background(), canvasID, noteID, req)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), note, "")
}
