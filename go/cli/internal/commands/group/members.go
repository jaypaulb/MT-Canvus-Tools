package group

import (
	"context"
	"strconv"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var listMembersCmd = &cobra.Command{
	Use:   "list-members <group-id>",
	Short: "List group members",
	Args:  cobra.ExactArgs(1),
	RunE:  runListMembers,
}

func init() {
	GroupCmd.AddCommand(listMembersCmd)
}

func runListMembers(cmd *cobra.Command, args []string) error {
	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	members, err := sess.ListGroupMembers(context.Background(), groupID)
	if err != nil {
		return err
	}

	items := make([]interface{}, len(members))
	for i, member := range members {
		items[i] = member
	}
	return output.OutputList(cmd.Context(), items, "")
}
