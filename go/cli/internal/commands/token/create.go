package token

import (
	"context"
	"strconv"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new access token",
	RunE:  runCreate,
}

var (
	createUserID string
	createDesc   string
)

func init() {
	createCmd.Flags().StringVar(&createUserID, "user-id", "", "User ID")
	createCmd.Flags().StringVar(&createDesc, "description", "", "Token description")
	createCmd.MarkFlagRequired("user-id")
	createCmd.MarkFlagRequired("description")
	TokenCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	userID, err := strconv.ParseInt(createUserID, 10, 64)
	if err != nil {
		return err
	}

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	req := canvus.CreateAccessTokenRequest{
		Description: createDesc,
	}

	token, err := sess.CreateAccessToken(context.Background(), userID, req)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), token, "")
}
