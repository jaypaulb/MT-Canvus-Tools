package token

import (
	"context"
	"strconv"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List access tokens for a user",
	RunE:  runList,
}

var listUserID string

func init() {
	listCmd.Flags().StringVar(&listUserID, "user-id", "", "User ID")
	listCmd.MarkFlagRequired("user-id")
	TokenCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	userID, err := strconv.ParseInt(listUserID, 10, 64)
	if err != nil {
		return err
	}

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	tokens, err := sess.ListAccessTokens(context.Background(), userID)
	if err != nil {
		return err
	}

	items := make([]interface{}, len(tokens))
	for i, token := range tokens {
		items[i] = token
	}
	return output.OutputList(cmd.Context(), items, "")
}
