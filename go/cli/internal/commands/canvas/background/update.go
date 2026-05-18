package background

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <canvas-id>",
	Short: "Update canvas background (color/haze)",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
}

var updateJSON string

func init() {
	updateCmd.Flags().StringVar(&updateJSON, "json", "", "Background config as JSON")
	updateCmd.MarkFlagRequired("json")
	BackgroundCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	canvasID := args[0]

	var req interface{}
	if err := json.Unmarshal([]byte(updateJSON), &req); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	err = sess.PatchCanvasBackground(context.Background(), canvasID, req)
	if err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Background updated successfully")
	return nil
}
