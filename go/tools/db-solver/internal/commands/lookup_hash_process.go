package commands

import (
	"path/filepath"
	"sync"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/backup"
	canvusinternal "github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/database"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/filesystem"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/logging"
)

// processAssetsParallel runs processAssetWithCatalogs for each unique asset using
// a worker pool sized to MaxConcurrentAPI.
func (cmd *LookupHashCommand) processAssetsParallel(
	assets []canvusinternal.AssetInfo,
	dbCatalog map[string]database.AssetFileRecord,
	assetsCatalog *filesystem.HashCatalog,
	backupCatalog *filesystem.HashCatalog,
) []HashLookupResult {
	logger := logging.GetLogger()
	workers := cmd.config.Performance.MaxConcurrentAPI
	if workers <= 0 {
		workers = 4
	}
	results := make([]HashLookupResult, len(assets))
	workChan := make(chan int, len(assets))

	var wg sync.WaitGroup
	var progressMu sync.Mutex
	processed := 0

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range workChan {
				progressMu.Lock()
				processed++
				cur := processed
				progressMu.Unlock()
				if cur%100 == 0 || cur == 1 {
					logger.Info("Progress: %d/%d unique files processed", cur, len(assets))
				}
				results[i] = cmd.processAssetWithCatalogs(assets[i], dbCatalog, assetsCatalog, backupCatalog)
			}
		}()
	}
	for i := range assets {
		workChan <- i
	}
	close(workChan)
	wg.Wait()
	logger.Info("Completed processing %d assets", len(assets))
	return results
}

// processAssetWithCatalogs resolves an asset-without-hash using pre-built catalogs
// (O(1) hash lookups) or on-demand filesystem search in low-memory mode.
func (cmd *LookupHashCommand) processAssetWithCatalogs(
	asset canvusinternal.AssetInfo,
	dbCatalog map[string]database.AssetFileRecord,
	assetsCatalog *filesystem.HashCatalog,
	backupCatalog *filesystem.HashCatalog,
) HashLookupResult {
	logger := logging.GetLogger()
	result := HashLookupResult{Asset: asset}

	// 1. Database lookup.
	dbRecord, found := dbCatalog[asset.OriginalFilename]
	if !found {
		result.Error = "not found in database"
		logger.Info("DB: %q on canvas %q → Not Found", asset.OriginalFilename, asset.CanvasName)
		return result
	}
	result.FoundInDatabase = true
	result.PublicHash = dbRecord.PublicHash
	result.PrivateHash = dbRecord.PrivateHash

	hashSuffix := hashSuffix4(dbRecord.PublicHash)
	logger.Info("DB: %q on canvas %q → [...%s]", asset.OriginalFilename, asset.CanvasName, hashSuffix)

	if dbRecord.PrivateHash == "" {
		return result
	}

	// 2. Assets folder lookup.
	foundInAssets, assetsPath := cmd.lookupInFolder(
		dbRecord.PrivateHash, cmd.config.Paths.AssetsFolder, assetsCatalog)
	if foundInAssets {
		result.FoundInAssetsFolder = true
		result.AssetsFolderPath = assetsPath
		logger.Info("Assets: %q [...%s] → Found at %s", asset.OriginalFilename, hashSuffix, assetsPath)
		return result
	}
	logger.Info("Assets: %q [...%s] → Not Found", asset.OriginalFilename, hashSuffix)

	// 3. Backup folder lookup.
	foundInBackup, backupPath, backupExt := cmd.lookupInBackup(
		dbRecord.PrivateHash, cmd.config.Paths.BackupRootFolder, backupCatalog)
	if foundInBackup {
		result.FoundInBackup = true
		result.BackupPath = backupPath
		targetPath := cmd.getTargetPathFromHash(dbRecord.PrivateHash, backupExt)
		if cmd.dryRun {
			logger.Info("Backup: %q [...%s] → Found at %s | Restore: DRY-RUN → Target: %s",
				asset.OriginalFilename, hashSuffix, backupPath, targetPath)
		} else {
			logger.Info("Backup: %q [...%s] → Found at %s | Restore: Active → Target: %s",
				asset.OriginalFilename, hashSuffix, backupPath, targetPath)
			// Restoration is logged but deferred — full restore is implemented in the restorer.
		}
	} else {
		logger.Info("Backup: %q [...%s] → Not Found", asset.OriginalFilename, hashSuffix)
	}
	return result
}

// lookupInFolder checks for a hash using the pre-built catalog if available,
// falling back to an on-demand filesystem search.
func (cmd *LookupHashCommand) lookupInFolder(
	hash, folder string,
	catalog *filesystem.HashCatalog,
) (bool, string) {
	if catalog != nil {
		if entry, ok := catalog.Lookup(hash); ok {
			return true, entry.FilePath
		}
		return false, ""
	}
	r, err := filesystem.SearchHashInAssetsFolder(folder, hash)
	if err == nil && r.Found {
		return true, r.FilePath
	}
	return false, ""
}

// lookupInBackup checks the backup root for a hash using the catalog or
// on-demand search.
func (cmd *LookupHashCommand) lookupInBackup(
	hash, backupRoot string,
	catalog *filesystem.HashCatalog,
) (bool, string, string) {
	if catalog != nil {
		if entry, ok := catalog.Lookup(hash); ok {
			return true, entry.FilePath, entry.Extension
		}
		return false, "", ""
	}
	searcher := backup.NewSearcher(backupRoot)
	sr, err := searcher.SearchForAssets([]string{hash})
	if err == nil {
		if files, ok := sr.FoundFiles[hash]; ok && len(files) > 0 {
			return true, files[0].Path, files[0].Extension
		}
	}
	return false, "", ""
}

// getTargetPathFromHash builds the restore target path using the two-char subfolder
// convention used by the Canvus asset store.
func (cmd *LookupHashCommand) getTargetPathFromHash(hash, ext string) string {
	if len(hash) >= 2 {
		return filepath.Join(cmd.config.Paths.AssetsFolder, hash[:2], hash+ext)
	}
	return filepath.Join(cmd.config.Paths.AssetsFolder, hash+ext)
}

// hashSuffix4 returns the last four characters of a hash, or "N/A".
func hashSuffix4(hash string) string {
	if len(hash) >= 4 {
		return hash[len(hash)-4:]
	}
	return "N/A"
}
