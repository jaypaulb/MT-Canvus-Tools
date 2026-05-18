package system

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var getConfigCmd = &cobra.Command{
	Use:   "get-config",
	Short: "Get server configuration",
	RunE:  runGetConfig,
}

var updateConfigCmd = &cobra.Command{
	Use:   "update-config",
	Short: "Update server configuration",
	RunE:  runUpdateConfig,
}

var sendTestEmailCmd = &cobra.Command{
	Use:   "send-test-email <recipient>",
	Short: "Send a test email to the given recipient",
	Args:  cobra.ExactArgs(1),
	RunE:  runSendTestEmail,
}

var (
	updateConfigJSON string
)

func init() {
	updateConfigCmd.Flags().StringVar(&updateConfigJSON, "json", "", "Config data as JSON")
	updateConfigCmd.MarkFlagRequired("json")
	SystemCmd.AddCommand(getConfigCmd)
	SystemCmd.AddCommand(updateConfigCmd)
	SystemCmd.AddCommand(sendTestEmailCmd)
}

func runGetConfig(cmd *cobra.Command, args []string) error {
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	config, err := sess.GetServerConfig(context.Background())
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), config, "")
}

func runUpdateConfig(cmd *cobra.Command, args []string) error {
	var req canvus.ServerConfig
	if err := json.Unmarshal([]byte(updateConfigJSON), &req); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	config, err := sess.UpdateServerConfig(context.Background(), req)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), config, "")
}

func runSendTestEmail(cmd *cobra.Command, args []string) error {
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	recipient := args[0]
	if err := sess.SendTestEmail(context.Background(), recipient); err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), fmt.Sprintf("Test email sent to %s", recipient))
	return nil
}
