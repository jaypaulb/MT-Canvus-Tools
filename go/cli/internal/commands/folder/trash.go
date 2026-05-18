package folder

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var trashCmd = &cobra.Command{
	Use:   "trash <folder-id>",
	Short: "Move folder to trash",
	Long: `Move a folder to the trash.

Examples:
  # Move folder to trash
  canvus folder trash folder-123`,
	Args: cobra.ExactArgs(1),
	RunE: runTrash,
}

func init() {
	FolderCmd.AddCommand(trashCmd)
}

func runTrash(cmd *cobra.Command, args []string) error {
	folderID := args[0]

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	folder, err := sess.TrashFolder(context.Background(), folderID, "")
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), folder, "")
}
