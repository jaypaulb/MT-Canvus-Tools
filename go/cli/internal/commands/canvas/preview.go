package canvas

import (
	"context"
	"fmt"
	"os"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var previewCmd = &cobra.Command{
	Use:   "preview <canvas-id>",
	Short: "Get canvas preview image",
	Args:  cobra.ExactArgs(1),
	RunE:  runPreview,
}

var previewOutput string

func init() {
	previewCmd.Flags().StringVar(&previewOutput, "output-file", "", "Output file path (default: stdout)")
	CanvasCmd.AddCommand(previewCmd)
}

func runPreview(cmd *cobra.Command, args []string) error {
	canvasID := args[0]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	data, err := sess.GetCanvasPreview(context.Background(), canvasID)
	if err != nil {
		return err
	}

	if previewOutput != "" {
		if err := os.WriteFile(previewOutput, data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		output.OutputSuccess(cmd.Context(), fmt.Sprintf("Preview saved to %s", previewOutput))
	} else {
		os.Stdout.Write(data)
	}
	return nil
}
