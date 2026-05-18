package mipmap

import (
	"context"
	"fmt"
	"os"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var assetCmd = &cobra.Command{
	Use:   "asset",
	Short: "Get asset by hash",
	RunE:  runAsset,
}

var (
	assetCanvasID string
	assetHash     string
	assetOutput   string
)

func init() {
	assetCmd.Flags().StringVar(&assetCanvasID, "canvas-id", "", "Canvas ID")
	assetCmd.Flags().StringVar(&assetHash, "hash", "", "Asset hash")
	assetCmd.Flags().StringVar(&assetOutput, "output-file", "", "Output file path (default: stdout)")
	assetCmd.MarkFlagRequired("canvas-id")
	assetCmd.MarkFlagRequired("hash")
	MipmapCmd.AddCommand(assetCmd)
}

func runAsset(cmd *cobra.Command, args []string) error {
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	data, err := sess.GetAssetByHash(context.Background(), assetCanvasID, assetHash)
	if err != nil {
		return err
	}

	if assetOutput != "" {
		if err := os.WriteFile(assetOutput, data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		output.OutputSuccess(cmd.Context(), fmt.Sprintf("Asset saved to %s", assetOutput))
	} else {
		os.Stdout.Write(data)
	}
	return nil
}
