package folder

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <folder-id>",
	Short: "Delete a folder",
	Long: `Permanently delete a folder by ID.

Examples:
  # Delete a folder
  canvus folder delete folder-123`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

func init() {
	FolderCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	folderID := args[0]

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	err = sess.DeleteFolder(context.Background(), folderID)
	if err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Folder deleted successfully")
	return nil
}
