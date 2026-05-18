package folder

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename <folder-id> <new-name>",
	Short: "Rename a folder",
	Long: `Rename an existing folder by ID.

Examples:
  # Rename a folder
  canvus folder rename folder-123 "New Folder Name"`,
	Args: cobra.ExactArgs(2),
	RunE: runRename,
}

func init() {
	FolderCmd.AddCommand(renameCmd)
}

func runRename(cmd *cobra.Command, args []string) error {
	folderID := args[0]
	newName := args[1]

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	folder, err := sess.RenameFolder(context.Background(), folderID, newName)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), folder, "")
}
