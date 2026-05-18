package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/logging"
)

// generateReport writes a plain-text hash lookup report to the current directory.
func (cmd *LookupHashCommand) generateReport(results []HashLookupResult) error {
	logger := logging.GetLogger()
	reportPath := "hash_lookup_report.txt"

	mode := "LIVE"
	if cmd.dryRun {
		mode = "DRY-RUN"
	}

	var sb strings.Builder
	sb.WriteString("DB Solver — Hash Lookup Report (Assets Without Hash)\n")
	sb.WriteString(strings.Repeat("=", 60) + "\n")
	sb.WriteString(fmt.Sprintf("Generated : %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Mode      : %s\n", mode))
	sb.WriteString(fmt.Sprintf("Processed : %d\n\n", len(results)))

	for i, r := range results {
		sb.WriteString(fmt.Sprintf("Asset %d:\n", i+1))
		sb.WriteString(fmt.Sprintf("  Canvas  : %s (ID: %s)\n", r.Asset.CanvasName, r.Asset.CanvasID))
		sb.WriteString(fmt.Sprintf("  Widget  : %s (ID: %s, Type: %s)\n", r.Asset.WidgetName, r.Asset.WidgetID, r.Asset.WidgetType))
		sb.WriteString(fmt.Sprintf("  File    : %s\n", r.Asset.OriginalFilename))

		if r.FoundInDatabase {
			sb.WriteString("  [OK] Found in Database\n")
			sb.WriteString(fmt.Sprintf("       Public Hash  : %s\n", r.PublicHash))
			sb.WriteString(fmt.Sprintf("       Private Hash : %s\n", r.PrivateHash))

			if r.FoundInAssetsFolder {
				sb.WriteString(fmt.Sprintf("  [OK] Found in Assets Folder: %s\n", r.AssetsFolderPath))
			} else {
				sb.WriteString("  [--] Not found in Assets Folder\n")
				if r.FoundInBackup {
					sb.WriteString(fmt.Sprintf("  [OK] Found in Backup: %s\n", r.BackupPath))
				} else {
					sb.WriteString("  [--] Not found in Backup\n")
				}
			}
		} else {
			sb.WriteString("  [--] Not found in Database\n")
			if r.Error != "" {
				sb.WriteString(fmt.Sprintf("       Error: %s\n", r.Error))
			}
		}
		sb.WriteString("\n")
	}

	if err := os.WriteFile(reportPath, []byte(sb.String()), 0o644); err != nil {
		return fmt.Errorf("write report to %s: %w", reportPath, err)
	}
	logger.Info("Report saved to: %s", reportPath)
	return nil
}
