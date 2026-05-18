package videooutput

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

// updateCmd PATCHes a video-output's source via the SDK's
// SetVideoOutputSourceByID helper. The legacy CLI called a now-removed
// UpdateVideoOutput method; the new SDK exposes the same endpoint through
// SetVideoOutputSourceByID (parity-matrix §1.1.10 row "Set video output
// source (by ID)").
//
// Behaviour note: the first positional argument was renamed from
// <canvas-id> to <client-id> because the endpoint is keyed on the client
// hosting the output, not on a canvas.
var updateCmd = &cobra.Command{
	Use:   "update <client-id> <output-id>",
	Short: "Update a video output's source",
	Args:  cobra.ExactArgs(2),
	RunE:  runUpdate,
}

var updateJSON string

func init() {
	updateCmd.Flags().StringVar(&updateJSON, "json", "", "Source data as JSON (matches the SetVideoOutputSourceByID body)")
	if err := updateCmd.MarkFlagRequired("json"); err != nil {
		panic(err)
	}
	VideoOutputCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	clientID, outputID := args[0], args[1]

	var req any
	if err := json.Unmarshal([]byte(updateJSON), &req); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	if err := sess.SetVideoOutputSourceByID(context.Background(), clientID, outputID, req); err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), fmt.Sprintf("Video output %s source updated", outputID))
	return nil
}
