package colorpreset

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var patchCmd = &cobra.Command{
	Use:   "patch <canvas-id>",
	Short: "Patch color presets",
	Args:  cobra.ExactArgs(1),
	RunE:  runPatch,
}

var patchJSON string

func init() {
	patchCmd.Flags().StringVar(&patchJSON, "json", "", "Presets data as JSON")
	patchCmd.MarkFlagRequired("json")
	ColorPresetCmd.AddCommand(patchCmd)
}

func runPatch(cmd *cobra.Command, args []string) error {
	canvasID := args[0]

	var req canvus.ColorPresets
	if err := json.Unmarshal([]byte(patchJSON), &req); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	presets, err := sess.PatchColorPresets(context.Background(), canvasID, &req)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), presets, "")
}
