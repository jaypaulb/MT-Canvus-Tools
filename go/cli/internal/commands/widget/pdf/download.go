package pdf

import (
	"context"
	"fmt"
	"os"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download <canvas-id> <pdf-id>",
	Short: "Download PDF file",
	Args:  cobra.ExactArgs(2),
	RunE:  runDownload,
}

var downloadOutput string

func init() {
	downloadCmd.Flags().StringVar(&downloadOutput, "output-file", "", "Output file path (required)")
	downloadCmd.MarkFlagRequired("output-file")
	PDFCmd.AddCommand(downloadCmd)
}

func runDownload(cmd *cobra.Command, args []string) error {
	canvasID, pdfID := args[0], args[1]
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	data, err := sess.DownloadPDF(context.Background(), canvasID, pdfID)
	if err != nil {
		return err
	}

	if err := os.WriteFile(downloadOutput, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	output.OutputSuccess(cmd.Context(), fmt.Sprintf("PDF saved to %s", downloadOutput))
	return nil
}
