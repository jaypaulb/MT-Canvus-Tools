// Package commands implements the Cobra subcommands for db-solver.
package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	canvussdk "github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/backup"
	canvusinternal "github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/database"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/filesystem"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/logging"
)

// LookupHashCommand processes assets that have no hash by querying the Postgres
// database, then searching the assets and backup folders.
type LookupHashCommand struct {
	config       *config.Config
	dryRun       bool
	iniPath      string
	skipArchived bool
	lowMemory    bool
}

// NewLookupHashCommand creates a new LookupHashCommand.
func NewLookupHashCommand(cfg *config.Config) *LookupHashCommand {
	return &LookupHashCommand{config: cfg}
}

// SetDryRun enables or disables dry-run mode.
func (cmd *LookupHashCommand) SetDryRun(v bool) { cmd.dryRun = v }

// SetINIPath overrides the auto-detected INI path.
func (cmd *LookupHashCommand) SetINIPath(v string) { cmd.iniPath = v }

// SetSkipArchived controls whether archived canvases are skipped.
func (cmd *LookupHashCommand) SetSkipArchived(v bool) { cmd.skipArchived = v }

// SetLowMemory enables low-memory mode (on-demand search instead of pre-built catalogs).
func (cmd *LookupHashCommand) SetLowMemory(v bool) { cmd.lowMemory = v }

// Execute runs the lookup-hash workflow.
func (cmd *LookupHashCommand) Execute(cobraCmd *cobra.Command, args []string) error {
	logger := logging.GetLogger()
	logger.Info("Starting Hash Lookup for Assets Without Hash")
	if cmd.dryRun {
		logger.Info("DRY-RUN MODE: no files will be restored")
	}

	// Step 1: Validate database connection before touching the API.
	logger.Info("Step 1: Validating database configuration...")
	dbClient, err := cmd.connectDatabase(logger)
	if err != nil {
		return err
	}
	defer dbClient.Close()

	// Step 2: Connect to Canvus Server.
	logger.Info("Step 2: Validating Canvus Server configuration...")
	session, err := cmd.connectCanvus(logger)
	if err != nil {
		return err
	}
	ctx := context.Background()
	defer session.Logout(ctx) //nolint:errcheck

	// Step 3: Validate filesystem paths.
	logger.Info("Step 3: Validating filesystem paths...")
	assetsFolder := cmd.config.Paths.AssetsFolder
	backupFolder := cmd.config.Paths.BackupRootFolder
	if err := requireDir(assetsFolder, "assets folder"); err != nil {
		return err
	}
	if err := requireDir(backupFolder, "backup root folder"); err != nil {
		return err
	}

	// Step 4: Discover assets (including those without a hash).
	logger.Info("Step 4: Discovering assets from Canvus Server...")
	opts := canvusinternal.DiscoveryOptions{SkipArchived: cmd.skipArchived}
	discovery, err := canvusinternal.DiscoverAllAssetsWithOptions(session, cmd.config.Performance.MaxConcurrentAPI, opts)
	if err != nil {
		return fmt.Errorf("asset discovery: %w", err)
	}
	logger.Info("Found %d assets without hash", len(discovery.AssetsWithoutHash))
	if len(discovery.AssetsWithoutHash) == 0 {
		logger.Info("No assets without hash found. Nothing to process.")
		return nil
	}

	// Step 5: Build hash catalogs (skip in low-memory mode).
	var assetsCatalog, backupCatalog *filesystem.HashCatalog
	if !cmd.lowMemory {
		logger.Info("Step 5: Building hash catalogs...")
		assetsCatalog, backupCatalog = cmd.buildCatalogs(assetsFolder, backupFolder, logger)
	} else {
		logger.Info("Step 5: LOW MEMORY MODE — skipping catalog pre-build")
	}

	// Step 6: Batch-load database records for all assets without hash.
	logger.Info("Step 6: Loading database records...")
	dbCatalog := buildDBCatalog(dbClient, discovery.AssetsWithoutHash, logger)
	logger.Info("Database catalog ready: %d records", len(dbCatalog))

	// Step 7: Deduplicate and process.
	logger.Info("Step 7: Deduplicating assets...")
	uniqueList, assetGroups := deduplicateAssets(discovery.AssetsWithoutHash, logger)

	logger.Info("Processing %d unique files with %d workers...", len(uniqueList), cmd.config.Performance.MaxConcurrentAPI)
	uniqueResults := cmd.processAssetsParallel(uniqueList, dbCatalog, assetsCatalog, backupCatalog)

	// Expand results back to all widget instances.
	results := expandResults(uniqueList, uniqueResults, assetGroups)

	// Step 8: Generate report.
	logger.Info("Step 8: Generating report...")
	if err := cmd.generateReport(results); err != nil {
		return fmt.Errorf("generate report: %w", err)
	}

	// Step 9: Summary.
	cmd.printSummary(results, uniqueList, logger)
	return nil
}

// connectDatabase resolves the INI path, loads config, and opens a DB client.
func (cmd *LookupHashCommand) connectDatabase(logger *logging.Logger) (*database.Client, error) {
	iniPath := cmd.iniPath
	if iniPath == "" {
		var err error
		iniPath, err = database.FindINIFile()
		if err != nil {
			logger.Warn("Could not auto-detect INI file: %v", err)
			iniPath = `C:\ProgramData\MultiTaction\Canvus\mt-canvus-server.ini`
		}
	}
	logger.Info("Loading database configuration from: %s", iniPath)
	dbConfig, err := database.LoadDBConfigFromINI(iniPath)
	if err != nil {
		return nil, fmt.Errorf("load database configuration: %w", err)
	}
	logger.Info("Testing database connection...")
	dbClient, err := database.NewClient(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("connect to database (%s:%s/%s): %w",
			dbConfig.Host, dbConfig.Port, dbConfig.Database, err)
	}
	logger.Info("Database connection successful")
	return dbClient, nil
}

// connectCanvus builds a Canvus session and authenticates.
func (cmd *LookupHashCommand) connectCanvus(logger *logging.Logger) (*canvussdk.Session, error) {
	if cmd.config.CanvusServer.Username == "" || cmd.config.CanvusServer.Password == "" {
		return nil, fmt.Errorf("canvus server credentials not configured")
	}
	session := newSession(cmd.config)
	ctx := context.Background()
	logger.Info("Authenticating with Canvus Server...")
	if err := session.Login(ctx, cmd.config.CanvusServer.Username, cmd.config.CanvusServer.Password); err != nil {
		return nil, fmt.Errorf("canvus authentication failed (URL=%s): %w", cmd.config.GetCanvusAPIURL(), err)
	}
	logger.Info("Canvus Server authentication successful")
	return session, nil
}

// buildCatalogs pre-builds hash catalogs for the assets and backup folders.
func (cmd *LookupHashCommand) buildCatalogs(assetsFolder, backupFolder string, logger *logging.Logger) (*filesystem.HashCatalog, *filesystem.HashCatalog) {
	ac := filesystem.NewHashCatalog()
	if err := ac.BuildCatalogFromFolder(assetsFolder); err != nil {
		logger.Warn("Failed to build assets catalog: %v", err)
		ac = nil
	}
	bc := filesystem.NewHashCatalog()
	if err := bc.BuildCatalogFromFolder(backupFolder); err != nil {
		logger.Warn("Failed to build backup catalog: %v", err)
		bc = nil
	}
	if ac != nil {
		logger.Info("Assets catalog: %d hashes", ac.Size())
	}
	if bc != nil {
		logger.Info("Backup catalog: %d hashes", bc.Size())
	}
	return ac, bc
}

// requireDir returns an error if path does not exist or is empty.
func requireDir(path, label string) error {
	if path == "" {
		return fmt.Errorf("%s not configured", label)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist: %s", label, path)
	}
	return nil
}

// printSummary logs per-category unique and instance counts.
func (cmd *LookupHashCommand) printSummary(results []HashLookupResult, uniqueList []canvusinternal.AssetInfo, logger *logging.Logger) {
	type assetKey struct{ CanvasName, OriginalFilename string }
	inDB := make(map[assetKey]bool)
	inAssets := make(map[assetKey]bool)
	inBackup := make(map[assetKey]bool)
	notFound := make(map[assetKey]bool)

	for _, r := range results {
		k := assetKey{r.Asset.CanvasName, r.Asset.OriginalFilename}
		if r.FoundInDatabase {
			inDB[k] = true
		}
		if r.FoundInAssetsFolder {
			inAssets[k] = true
		}
		if r.FoundInBackup {
			inBackup[k] = true
		}
		if !r.FoundInDatabase && !r.FoundInAssetsFolder && !r.FoundInBackup {
			notFound[k] = true
		}
	}

	logger.Info("Summary:")
	logger.Info("  Total: %d widget instances across %d unique files", len(results), len(uniqueList))
	logger.Info("  Found in database: %d unique files", len(inDB))
	logger.Info("  Found in assets folder: %d unique files", len(inAssets))
	logger.Info("  Found in backup: %d unique files", len(inBackup))
	logger.Info("  Not found: %d unique files", len(notFound))
}

// newSession builds a Canvus SDK Session from config. Uses an insecure HTTP client
// when InsecureTLS is true (common for self-signed Canvus server certs).
func newSession(cfg *config.Config) *canvussdk.Session {
	sessionCfg := canvussdk.DefaultSessionConfig()
	sessionCfg.BaseURL = cfg.GetCanvusAPIURL()
	if cfg.CanvusServer.InsecureTLS {
		return canvussdk.NewSession(sessionCfg, canvussdk.WithHTTPClient(insecureHTTPClient()))
	}
	return canvussdk.NewSession(sessionCfg)
}

// buildDBCatalog pre-loads database records for all assets-without-hash into a map
// keyed by original filename, avoiding repeated per-asset queries.
func buildDBCatalog(dbClient *database.Client, assets []canvusinternal.AssetInfo, logger *logging.Logger) map[string]database.AssetFileRecord {
	catalog := make(map[string]database.AssetFileRecord)
	for _, asset := range assets {
		if asset.OriginalFilename == "" {
			continue
		}
		if _, ok := catalog[asset.OriginalFilename]; ok {
			continue
		}
		records, err := dbClient.FindAssetByOriginalFilename(asset.OriginalFilename)
		if err != nil || len(records) == 0 {
			records, err = dbClient.FindAssetByOriginalFilenameCaseInsensitive(asset.OriginalFilename)
		}
		if err == nil && len(records) > 0 {
			catalog[asset.OriginalFilename] = records[0]
		}
	}
	return catalog
}

type assetKey struct {
	CanvasName       string
	OriginalFilename string
}

// deduplicateAssets groups assets by (CanvasName, OriginalFilename) and returns
// a slice of representative assets and the full group map.
func deduplicateAssets(assets []canvusinternal.AssetInfo, logger *logging.Logger) ([]canvusinternal.AssetInfo, map[assetKey][]canvusinternal.AssetInfo) {
	groups := make(map[assetKey][]canvusinternal.AssetInfo)
	skipped := 0
	for _, a := range assets {
		if a.OriginalFilename == "" {
			skipped++
			continue
		}
		k := assetKey{a.CanvasName, a.OriginalFilename}
		groups[k] = append(groups[k], a)
	}
	unique := make([]canvusinternal.AssetInfo, 0, len(groups))
	for _, group := range groups {
		unique = append(unique, group[0])
	}
	logger.Info("Deduplicated: %d unique files from %d total (skipped %d without filename)",
		len(unique), len(assets), skipped)
	return unique, groups
}

// expandResults replicates each unique result across all widget instances in the group.
func expandResults(uniqueList []canvusinternal.AssetInfo, uniqueResults []HashLookupResult, groups map[assetKey][]canvusinternal.AssetInfo) []HashLookupResult {
	var all []HashLookupResult
	for i, u := range uniqueList {
		k := assetKey{u.CanvasName, u.OriginalFilename}
		for _, instance := range groups[k] {
			r := uniqueResults[i]
			r.Asset = instance
			all = append(all, r)
		}
	}
	return all
}

// HashLookupResult holds the result of processing one asset-without-hash.
type HashLookupResult struct {
	Asset               canvusinternal.AssetInfo
	FoundInDatabase     bool
	PublicHash          string
	PrivateHash         string
	FoundInAssetsFolder bool
	AssetsFolderPath    string
	FoundInBackup       bool
	BackupPath          string
	Error               string
}

// getTargetPath calculates the restore target path using the two-char subfolder convention.
func (cmd *LookupHashCommand) getTargetPath(assetsFolder, hash string, backupFile backup.BackupFile) string {
	if len(hash) >= 2 {
		return fmt.Sprintf("%s/%s/%s%s", assetsFolder, hash[:2], backupFile.Hash, backupFile.Extension)
	}
	return fmt.Sprintf("%s/%s%s", assetsFolder, backupFile.Hash, backupFile.Extension)
}
