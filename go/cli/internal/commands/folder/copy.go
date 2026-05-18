package folder

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy <folder-id> <parent-folder-id>",
	Short: "Copy a folder",
	Long: `Copy a folder to a different parent folder.

Examples:
  # Copy a folder
  canvus folder copy folder-123 parent-456

  # Copy with conflict resolution
  canvus folder copy folder-123 parent-456 --conflicts skip`,
	Args: cobra.ExactArgs(2),
	RunE: runCopy,
}

var copyConflicts string

func init() {
	copyCmd.Flags().StringVar(&copyConflicts, "conflicts", "", "Conflict resolution strategy (skip, replace)")
	FolderCmd.AddCommand(copyCmd)
}

func runCopy(cmd *cobra.Command, args []string) error {
	folderID := args[0]
	parentID := args[1]

	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	folder, err := sess.CopyFolder(context.Background(), folderID, parentID, copyConflicts)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), folder, "")
}
