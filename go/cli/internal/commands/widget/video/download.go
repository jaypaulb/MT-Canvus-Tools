package video

import (
	"context"
	"fmt"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
	"os"
)

var downloadCmd = &cobra.Command{
	Use:   "download <canvas-id> <video-id>",
	Short: "Download video file",
	Args:  cobra.ExactArgs(2),
	RunE:  runDownload,
}

var downloadOutput string

func init() {
	downloadCmd.Flags().StringVar(&downloadOutput, "output-file", "", "Output file path (required)")
	downloadCmd.MarkFlagRequired("output-file")
	VideoCmd.AddCommand(downloadCmd)
}

func runDownload(cmd *cobra.Command, args []string) error {
	canvasID, videoID := args[0], args[1]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	data, err := sess.DownloadVideo(context.Background(), canvasID, videoID)
	if err != nil {
		return err
	}

	if err := os.WriteFile(downloadOutput, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	output.OutputSuccess(cmd.Context(), fmt.Sprintf("Video saved to %s", downloadOutput))
	return nil
}
