package mipmap

import (
	"context"
	"fmt"
	"os"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var levelCmd = &cobra.Command{
	Use:   "level",
	Short: "Get mipmap level data",
	RunE:  runLevel,
}

var (
	levelCanvasID string
	levelHash     string
	levelLevel    int
	levelPage     int
	levelOutput   string
)

func init() {
	levelCmd.Flags().StringVar(&levelCanvasID, "canvas-id", "", "Canvas ID")
	levelCmd.Flags().StringVar(&levelHash, "hash", "", "Asset hash")
	levelCmd.Flags().IntVar(&levelLevel, "level", 0, "Mipmap level")
	levelCmd.Flags().IntVar(&levelPage, "page", 0, "Page number (optional)")
	levelCmd.Flags().StringVar(&levelOutput, "output-file", "", "Output file path (default: stdout)")
	levelCmd.MarkFlagRequired("canvas-id")
	levelCmd.MarkFlagRequired("hash")
	MipmapCmd.AddCommand(levelCmd)
}

func runLevel(cmd *cobra.Command, args []string) error {
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	var page *int
	if cmd.Flags().Changed("page") {
		page = &levelPage
	}

	data, err := sess.GetMipmapLevel(context.Background(), levelCanvasID, levelHash, levelLevel, page)
	if err != nil {
		return err
	}

	if levelOutput != "" {
		if err := os.WriteFile(levelOutput, data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		output.OutputSuccess(cmd.Context(), fmt.Sprintf("Mipmap level saved to %s", levelOutput))
	} else {
		os.Stdout.Write(data)
	}
	return nil
}
