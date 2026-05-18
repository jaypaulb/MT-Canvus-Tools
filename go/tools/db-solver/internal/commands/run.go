package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/backup"
	canvusinternal "github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/filesystem"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/logging"
)

// RunCommand implements the run subcommand (complete discover → search → report workflow).
type RunCommand struct {
	config *config.Config
}

// NewRunCommand creates a new RunCommand.
func NewRunCommand(cfg *config.Config) *RunCommand {
	return &RunCommand{config: cfg}
}

// Execute runs the complete workflow sequentially.
func (cmd *RunCommand) Execute(cobraCmd *cobra.Command, args []string) error {
	logger := logging.GetLogger()
	logger.Info("Starting complete workflow")

	ctx := context.Background()
	session := newSession(cmd.config)

	logger.Info("Authenticating with Canvus Server: %s", cmd.config.CanvusServer.URL)
	if err := session.Login(ctx, cmd.config.CanvusServer.Username, cmd.config.CanvusServer.Password); err != nil {
		return fmt.Errorf("authentication: %w", err)
	}
	defer session.Logout(ctx) //nolint:errcheck

	// Step 1: Discover assets.
	logger.Info("Step 1: Discovering assets...")
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

	// Step 2: Scan local assets folder.
	logger.Info("Step 2: Scanning local assets folder...")
	scanResult, err := filesystem.ScanAssetsFolder(cmd.config.Paths.AssetsFolder)
	if err != nil {
		return fmt.Errorf("filesystem scan: %w", err)
	}
	logger.Info("Found %d files in assets folder (%.2f MB)",
		len(scanResult.Files), float64(scanResult.TotalSize)/(1024*1024))

	missingAssets := filesystem.FindMissingAssets(assetHashes, scanResult)
	logger.Info("Missing assets: %d", len(missingAssets))

	if len(missingAssets) == 0 {
		logger.Info("No missing assets found — all assets present!")
		return nil
	}

	// Step 3: Search backup.
	logger.Info("Step 3: Searching backup folder for missing assets...")
	searcher := backup.NewSearcher(cmd.config.Paths.BackupRootFolder)
	backupSearchResult, err := searcher.SearchForAssets(missingAssets)
	if err != nil {
		return fmt.Errorf("backup search: %w", err)
	}
	searcher.SortBackupFiles(backupSearchResult)

	if len(backupSearchResult.FoundFiles) > 0 {
		logger.Info("Found %d missing assets in backup folder", len(backupSearchResult.FoundFiles))
	} else {
		logger.Info("No assets found in backup")
	}

	// Step 4: Generate reports.
	logger.Info("Step 4: Generating reports...")
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
	discoverCmd := NewDiscoverCommand(cmd.config)
	if err := discoverCmd.generateReports(discoveryResult, missingAssets, missingInfos, backupSearchResult); err != nil {
		return fmt.Errorf("report generation: %w", err)
	}

	// Step 5: Summary.
	logger.Info("Step 5: Workflow summary")
	logger.Info("  Canvases: %d", len(discoveryResult.Canvases))
	logger.Info("  Unique assets: %d", len(uniqueAssets))
	logger.Info("  Files in folder: %d", len(scanResult.Files))
	logger.Info("  Missing assets: %d", len(missingAssets))
	if backupSearchResult != nil {
		logger.Info("  Found in backup: %d", len(backupSearchResult.FoundFiles))
		logger.Info("  Still missing: %d", len(backupSearchResult.MissingHashes))
	}
	logger.Info("Complete workflow finished. Reports: missing_assets_report.txt, missing_assets.csv")
	return nil
}
