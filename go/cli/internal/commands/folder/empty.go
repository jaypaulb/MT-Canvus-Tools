package folder

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var emptyCmd = &cobra.Command{
	Use:   "empty <folder-id>",
	Short: "Delete all contents of a folder",
	Long: `Delete all children (subfolders and canvases) of a folder.

Examples:
  # Empty a folder
  canvus folder empty folder-123`,
	Args: cobra.ExactArgs(1),
	RunE: runEmpty,
}

func init() {
	FolderCmd.AddCommand(emptyCmd)
}

func runEmpty(cmd *cobra.Command, args []string) error {
	folderID := args[0]

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	err = sess.DeleteFolderContents(context.Background(), folderID)
	if err != nil {
		return err
	}

	output.OutputSuccess(cmd.Context(), "Folder contents deleted successfully")
	return nil
}
