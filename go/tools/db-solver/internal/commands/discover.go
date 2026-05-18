package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/backup"
	canvusinternal "github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/filesystem"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/logging"
)

// DiscoverCommand handles the discover subcommand.
type DiscoverCommand struct {
	config *config.Config
}

// NewDiscoverCommand creates a new DiscoverCommand.
func NewDiscoverCommand(cfg *config.Config) *DiscoverCommand {
	return &DiscoverCommand{config: cfg}
}

// Execute runs the discover workflow: connect, discover, scan filesystem, report.
func (cmd *DiscoverCommand) Execute() error {
	logger := logging.GetLogger()
	logger.Info("Starting asset discovery...")
	logger.Info("Canvus Server: %s", cmd.config.CanvusServer.URL)
	logger.Info("Assets folder: %s", cmd.config.Paths.AssetsFolder)

	ctx := context.Background()
	session := newSession(cmd.config)

	logger.Info("Authenticating with Canvus Server...")
	if err := session.Login(ctx, cmd.config.CanvusServer.Username, cmd.config.CanvusServer.Password); err != nil {
		return fmt.Errorf("authentication: %w", err)
	}
	defer session.Logout(ctx) //nolint:errcheck

	logger.Info("Discovering assets from Canvus API...")
	discoveryResult, err := canvusinternal.DiscoverAllAssets(session, cmd.config.Performance.MaxConcurrentAPI)
	if err != nil {
		return fmt.Errorf("asset discovery: %w", err)
	}
	logger.Info("Found %d canvases with %d media assets",
		len(discoveryResult.Canvases), len(discoveryResult.Assets))

	uniqueAssets := discoveryResult.GetUniqueAssets()
	logger.Info("Unique assets (deduplicated): %d", len(uniqueAssets))

	assetHashes := make([]string, len(uniqueAssets))
	for i, a := range uniqueAssets {
		assetHashes[i] = a.Hash
	}

	logger.Info("Scanning assets folder...")
	scanResult, err := filesystem.ScanAssetsFolder(cmd.config.Paths.AssetsFolder)
	if err != nil {
		return fmt.Errorf("filesystem scan: %w", err)
	}
	logger.Info("Found %d files in assets folder (%.2f MB)",
		len(scanResult.Files), float64(scanResult.TotalSize)/(1024*1024))

	missingAssets := filesystem.FindMissingAssets(assetHashes, scanResult)
	logger.Info("Missing assets: %d", len(missingAssets))

	var backupSearchResult *backup.SearchResult
	if len(missingAssets) > 0 {
		logger.Info("Searching for missing assets in backup folder...")
		searcher := backup.NewSearcher(cmd.config.Paths.BackupRootFolder)
		backupSearchResult, err = searcher.SearchForAssets(missingAssets)
		if err != nil {
			return fmt.Errorf("backup search: %w", err)
		}
		searcher.SortBackupFiles(backupSearchResult)
		if len(backupSearchResult.FoundFiles) > 0 {
			logger.Info("Found %d missing assets in backup folder", len(backupSearchResult.FoundFiles))
		}
	}

	if len(missingAssets) > 0 {
		logger.Info("Generating reports...")
		if err := cmd.generateReports(discoveryResult, missingAssets, uniqueAssets, backupSearchResult); err != nil {
			return fmt.Errorf("report generation: %w", err)
		}
	} else {
		logger.Info("No missing assets found!")
	}

	cmd.printSummary(discoveryResult, scanResult, missingAssets)
	return nil
}

// generateReports produces both a detailed text report and a CSV.
func (cmd *DiscoverCommand) generateReports(
	discoveryResult *canvusinternal.DiscoveryResult,
	missingAssets []string,
	uniqueAssets []canvusinternal.AssetInfo,
	backupSearchResult *backup.SearchResult,
) error {
	missingMap := make(map[string]bool, len(missingAssets))
	for _, h := range missingAssets {
		missingMap[h] = true
	}
	var missingInfos []canvusinternal.AssetInfo
	for _, a := range uniqueAssets {
		if missingMap[a.Hash] {
			missingInfos = append(missingInfos, a)
		}
	}
	if err := cmd.generateDetailedReport(missingInfos, backupSearchResult); err != nil {
		return fmt.Errorf("detailed report: %w", err)
	}
	if err := cmd.generateCSVReport(missingInfos, backupSearchResult); err != nil {
		return fmt.Errorf("CSV report: %w", err)
	}
	return nil
}

func (cmd *DiscoverCommand) generateDetailedReport(missingAssets []canvusinternal.AssetInfo, backupResult *backup.SearchResult) error {
	reportPath := filepath.Join(cmd.config.Paths.OutputFolder, "missing_assets_report.txt")

	canvasMap := make(map[string][]canvusinternal.AssetInfo)
	for _, a := range missingAssets {
		canvasMap[a.CanvasName] = append(canvasMap[a.CanvasName], a)
	}

	var sb strings.Builder
	sb.WriteString("DB Solver — Missing Assets Report\n")
	sb.WriteString(fmt.Sprintf("Generated     : %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Missing total : %d\n", len(missingAssets)))
	if backupResult != nil {
		sb.WriteString(fmt.Sprintf("Found in backup: %d\n", len(backupResult.FoundFiles)))
		sb.WriteString(fmt.Sprintf("Still missing  : %d\n", len(backupResult.MissingHashes)))
	}
	sb.WriteString("\n")

	for canvasName, assets := range canvasMap {
		sb.WriteString(fmt.Sprintf("Canvas: %s (ID: %s)\n", canvasName, assets[0].CanvasID))
		for _, a := range assets {
			sb.WriteString(fmt.Sprintf("  Widget: %s (ID: %s, Type: %s)\n", a.WidgetName, a.WidgetID, a.WidgetType))
			sb.WriteString(fmt.Sprintf("    Hash: %s\n", a.Hash))
			if a.OriginalFilename != "" {
				sb.WriteString(fmt.Sprintf("    File: %s\n", a.OriginalFilename))
			}
			if backupResult != nil {
				if files, ok := backupResult.FoundFiles[a.Hash]; ok && len(files) > 0 {
					best := files[0]
					sb.WriteString("    Backup: Found\n")
					sb.WriteString(fmt.Sprintf("      Path    : %s\n", best.Path))
					sb.WriteString(fmt.Sprintf("      Size    : %d bytes (%.2f MB)\n", best.Size, float64(best.Size)/(1024*1024)))
					sb.WriteString(fmt.Sprintf("      Modified: %s\n", best.ModifiedTime.Format("2006-01-02 15:04:05")))
					sb.WriteString(fmt.Sprintf("      Copies  : %d\n", len(files)))
				} else {
					sb.WriteString("    Backup: Not found\n")
				}
			}
			sb.WriteString("\n")
		}
	}

	if err := writeFile(reportPath, sb.String()); err != nil {
		return err
	}
	fmt.Printf("Detailed report saved to: %s\n", reportPath)
	return nil
}

func (cmd *DiscoverCommand) generateCSVReport(missingAssets []canvusinternal.AssetInfo, backupResult *backup.SearchResult) error {
	reportPath := filepath.Join(cmd.config.Paths.OutputFolder, "missing_assets.csv")

	var sb strings.Builder
	sb.WriteString("Hash,WidgetType,OriginalFilename,CanvasID,CanvasName,WidgetID,WidgetName,BackupStatus,BackupPath,BackupSize,BackupModified,BackupCount,AllBackupPaths\n")

	for _, a := range missingAssets {
		status, backupPath, backupSize, backupMod, backupCount, allPaths := "Not Found", "", "", "", "0", ""
		if backupResult != nil {
			if files, ok := backupResult.FoundFiles[a.Hash]; ok && len(files) > 0 {
				best := files[0]
				status = "Found"
				backupPath = best.Path
				backupSize = fmt.Sprintf("%d", best.Size)
				backupMod = best.ModifiedTime.Format("2006-01-02 15:04:05")
				backupCount = fmt.Sprintf("%d", len(files))
				paths := make([]string, len(files))
				for i, f := range files {
					paths[i] = f.Path
				}
				allPaths = strings.Join(paths, ";")
			}
		}
		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			a.Hash, a.WidgetType, a.OriginalFilename, a.CanvasID, a.CanvasName,
			a.WidgetID, a.WidgetName,
			status, backupPath, backupSize, backupMod, backupCount, allPaths))
	}

	if err := writeFile(reportPath, sb.String()); err != nil {
		return err
	}
	fmt.Printf("CSV report saved to: %s\n", reportPath)
	return nil
}

func (cmd *DiscoverCommand) printSummary(
	discoveryResult *canvusinternal.DiscoveryResult,
	scanResult *filesystem.ScanResult,
	missingAssets []string,
) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("DISCOVERY SUMMARY")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Discovery Duration : %v\n", discoveryResult.Duration)
	fmt.Printf("Total Canvases     : %d\n", len(discoveryResult.Canvases))
	fmt.Printf("Total Media Assets : %d\n", len(discoveryResult.Assets))
	fmt.Printf("Unique Assets      : %d\n", len(discoveryResult.GetUniqueAssets()))
	fmt.Printf("Files in Folder    : %d\n", len(scanResult.Files))
	fmt.Printf("Assets Size        : %.2f MB\n", float64(scanResult.TotalSize)/(1024*1024))
	fmt.Printf("Missing Assets     : %d\n", len(missingAssets))
	if discoveryResult.ServerValidation != nil {
		fmt.Printf("Unique asset count : %d\n", discoveryResult.ServerValidation.TotalAssets)
	}
	if len(discoveryResult.Errors) > 0 {
		fmt.Printf("Errors encountered : %d\n", len(discoveryResult.Errors))
		for _, e := range discoveryResult.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}
	fmt.Println(strings.Repeat("=", 60))
}

// writeFile writes content to filename, creating directories as needed.
func writeFile(filename, content string) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return fmt.Errorf("create report directory: %w", err)
	}
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file %s: %w", filename, err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("write to file %s: %w", filename, err)
	}
	return nil
}
