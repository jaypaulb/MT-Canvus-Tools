package videooutput

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
	"strconv"
)

var setSourceCmd = &cobra.Command{
	Use:   "set-source <client-id> <index>",
	Short: "Set video output source",
	Args:  cobra.ExactArgs(2),
	RunE:  runSetSource,
}

var setSourceJSON string

func init() {
	setSourceCmd.Flags().StringVar(&setSourceJSON, "json", "", "Source data as JSON")
	setSourceCmd.MarkFlagRequired("json")
	VideoOutputCmd.AddCommand(setSourceCmd)
}

func runSetSource(cmd *cobra.Command, args []string) error {
	clientID := args[0]
	index, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid index: %w", err)
	}

	var req interface{}
	if err := json.Unmarshal([]byte(setSourceJSON), &req); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	err = sess.SetVideoOutputSource(context.Background(), clientID, index, req)
	if err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Video output source set successfully")
	return nil
}
