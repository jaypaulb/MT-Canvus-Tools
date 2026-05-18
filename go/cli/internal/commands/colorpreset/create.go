package colorpreset

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <canvas-id>",
	Short: "Create color preset",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

var createJSON string

func init() {
	createCmd.Flags().StringVar(&createJSON, "json", "", "Preset data as JSON")
	createCmd.MarkFlagRequired("json")
	ColorPresetCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	canvasID := args[0]

	var req interface{}
	if err := json.Unmarshal([]byte(createJSON), &req); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	preset, err := sess.CreateColorPreset(context.Background(), canvasID, req)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), preset, "")
}
