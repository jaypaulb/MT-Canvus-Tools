package token

import (
	"context"
	"strconv"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <token-id>",
	Short: "Delete an access token",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

var deleteUserID string

func init() {
	deleteCmd.Flags().StringVar(&deleteUserID, "user-id", "", "User ID")
	deleteCmd.MarkFlagRequired("user-id")
	TokenCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	tokenID := args[0]
	userID, err := strconv.ParseInt(deleteUserID, 10, 64)
	if err != nil {
		return err
	}

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	err = sess.DeleteAccessToken(context.Background(), userID, tokenID)
	if err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Access token deleted successfully")
	return nil
}
