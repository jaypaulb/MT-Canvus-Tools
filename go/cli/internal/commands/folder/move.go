package folder

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move <folder-id> <parent-folder-id>",
	Short: "Move a folder",
	Long: `Move a folder to a different parent folder.

Examples:
  # Move a folder
  canvus folder move folder-123 parent-456

  # Move with conflict resolution
  canvus folder move folder-123 parent-456 --conflicts skip`,
	Args: cobra.ExactArgs(2),
	RunE: runMove,
}

var moveConflicts string

func init() {
	moveCmd.Flags().StringVar(&moveConflicts, "conflicts", "", "Conflict resolution strategy (skip, replace)")
	FolderCmd.AddCommand(moveCmd)
}

func runMove(cmd *cobra.Command, args []string) error {
	folderID := args[0]
	parentID := args[1]

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	folder, err := sess.MoveFolder(context.Background(), folderID, parentID, moveConflicts)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), folder, "")
}
