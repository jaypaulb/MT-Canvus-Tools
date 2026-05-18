package mipmap

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get mipmap info",
	RunE:  runInfo,
}

var (
	infoCanvasID string
	infoHash     string
	infoPage     int
)

func init() {
	infoCmd.Flags().StringVar(&infoCanvasID, "canvas-id", "", "Canvas ID")
	infoCmd.Flags().StringVar(&infoHash, "hash", "", "Asset hash")
	infoCmd.Flags().IntVar(&infoPage, "page", 0, "Page number (optional)")
	infoCmd.MarkFlagRequired("canvas-id")
	infoCmd.MarkFlagRequired("hash")
	MipmapCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	var page *int
	if cmd.Flags().Changed("page") {
		page = &infoPage
	}

	info, err := sess.GetMipmapInfo(context.Background(), infoCanvasID, infoHash, page)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), info, "")
}
