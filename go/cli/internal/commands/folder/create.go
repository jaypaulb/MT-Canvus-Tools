package folder

import (
	"context"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new folder",
	Long: `Create a new folder in the Canvus system.

Examples:
  # Create a root-level folder
  canvus folder create --name "My Folder"

  # Create a folder inside another folder
  canvus folder create --name "Subfolder" --parent-id folder-123`,
	RunE: runCreate,
}

var (
	createName     string
	createParentID string
)

func init() {
	createCmd.Flags().StringVar(&createName, "name", "", "Folder name")
	createCmd.Flags().StringVar(&createParentID, "parent-id", "", "Parent folder ID (optional)")
	createCmd.MarkFlagRequired("name")
	FolderCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	sess, err := getSession(cmd)
	if err != nil {
		return err
	}

	req := canvus.CreateFolderRequest{
		Name:     createName,
		ParentID: createParentID,
	}

	folder, err := sess.CreateFolder(context.Background(), req)
	if err != nil {
		return err
	}

	return output.OutputSingle(cmd.Context(), folder, "")
}
