package image

import (
	"context"
	"fmt"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
	"os"
)

var downloadCmd = &cobra.Command{
	Use:   "download <canvas-id> <image-id>",
	Short: "Download image file",
	Args:  cobra.ExactArgs(2),
	RunE:  runDownload,
}

var downloadOutput string

func init() {
	downloadCmd.Flags().StringVar(&downloadOutput, "output-file", "", "Output file path (required)")
	downloadCmd.MarkFlagRequired("output-file")
	ImageCmd.AddCommand(downloadCmd)
}

func runDownload(cmd *cobra.Command, args []string) error {
	canvasID, imageID := args[0], args[1]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	data, err := sess.DownloadImage(context.Background(), canvasID, imageID)
	if err != nil {
		return err
	}

	if err := os.WriteFile(downloadOutput, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	output.OutputSuccess(cmd.Context(), fmt.Sprintf("Image saved to %s", downloadOutput))
	return nil
}
