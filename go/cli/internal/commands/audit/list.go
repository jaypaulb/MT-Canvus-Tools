package audit

import (
	"context"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List audit events",
	RunE:  runList,
}

var listPerPage int

func init() {
	listCmd.Flags().IntVar(&listPerPage, "per-page", 50, "Results per page")
	AuditCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	opts := &canvus.AuditLogOptions{
		PerPage: listPerPage,
	}

	events, err := sess.ListAuditEvents(context.Background(), opts)
	if err != nil {
		return err
	}

	items := make([]interface{}, len(events))
	for i, event := range events {
		items[i] = event
	}
	return output.OutputList(cmd.Context(), items, "")
}
