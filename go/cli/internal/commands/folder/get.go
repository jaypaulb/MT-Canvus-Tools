package folder

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <folder-id>",
	Short: "Get folder details",
	Long: `Get details of a specific folder by ID.

Examples:
  # Get folder details
  canvus folder get folder-123

  # Get folder as JSON
  canvus folder get folder-123 --output json`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

func init() {
	FolderCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	folderID := args[0]

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	folder, err := sess.GetFolder(context.Background(), folderID)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), folder, "")
}
