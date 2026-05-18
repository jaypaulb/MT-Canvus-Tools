package folder

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all folders",
	Long: `List all folders in the Canvus system.

Examples:
  # List all folders
  canvus folder list

  # List folders as JSON
  canvus folder list --output json`,
	RunE: runList,
}

func init() {
	FolderCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	folders, err := sess.ListFolders(context.Background())
	if err != nil {
		return err
	}

	items := make([]interface{}, len(folders))
	for i, folder := range folders {
		items[i] = folder
	}

	return output.OutputList(cmd.Context(), items, "")
}
