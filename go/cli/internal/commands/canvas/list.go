package canvas

import (
	"context"
	"fmt"
	"strings"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List canvases",
	Long: `List all canvases in your Canvus workspace.

You can optionally filter canvases by folder or by name pattern.

Examples:
  # List all canvases
  canvus canvas list

  # List canvases in a specific folder
  canvus canvas list --folder folder-123

  # Filter canvases by name pattern
  canvus canvas list --filter "Project"

  # Combine folder and filter
  canvus canvas list --folder folder-123 --filter "2024"

  # Output as JSON
  canvus canvas list --output json`,
	Args: cobra.NoArgs,
	RunE: runList,
}

var (
	listFolderID string
	listFilter   string
)

func init() {
	listCmd.Flags().StringVar(&listFolderID, "folder", "", "Filter by folder ID")
	listCmd.Flags().StringVar(&listFilter, "filter", "", "Filter canvases by name pattern")
	CanvasCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	// Get SDK session from context
	sess, err := session.GetSession(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Build filter for SDK if folder is specified
	var filter *canvus.Filter
	if listFolderID != "" {
		filter = &canvus.Filter{
			Criteria: map[string]interface{}{
				"folder_id": listFolderID,
			},
		}
	}

	// List canvases using SDK
	ctx := context.Background()
	canvases, err := sess.ListCanvases(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to list canvases: %w", err)
	}

	// Apply client-side name filtering if filter is provided
	var filteredCanvases []canvus.Canvas
	if listFilter != "" {
		filterLower := strings.ToLower(listFilter)
		for _, canvas := range canvases {
			if strings.Contains(strings.ToLower(canvas.Name), filterLower) {
				filteredCanvases = append(filteredCanvases, canvas)
			}
		}
	} else {
		filteredCanvases = canvases
	}

	// Convert to []interface{} for output formatting
	items := make([]interface{}, len(filteredCanvases))
	for i, canvas := range filteredCanvases {
		items[i] = canvas
	}

	// Format and output the results
	return output.OutputList(cmd.Context(), items, "")
}
