package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/logging"
)

// Restorer copies backup files into the live assets folder.
type Restorer struct {
	assetsFolder string
	logger       *logging.Logger
}

// NewRestorer creates a Restorer that writes into assetsFolder.
func NewRestorer(assetsFolder string) *Restorer {
	return &Restorer{
		assetsFolder: assetsFolder,
		logger:       logging.GetLogger(),
	}
}

// RestoreResult summarises a restoration run.
type RestoreResult struct {
	RestoredFiles []string
	FailedFiles   []string
	TotalBytes    int64
	Errors        []string
}

// RestoreAssets copies the best (newest) backup file for each found hash into
// the assets folder, preserving the relative path structure from the backup.
func (r *Restorer) RestoreAssets(searchResult *SearchResult) (*RestoreResult, error) {
	result := &RestoreResult{
		RestoredFiles: make([]string, 0),
		FailedFiles:   make([]string, 0),
		Errors:        make([]string, 0),
	}
	if len(searchResult.FoundFiles) == 0 {
		r.logger.Info("No backup files to restore")
		return result, nil
	}
	r.logger.Info("Restoring %d assets to: %s", len(searchResult.FoundFiles), r.assetsFolder)
	if err := os.MkdirAll(r.assetsFolder, 0o755); err != nil {
		return nil, fmt.Errorf("create assets folder: %w", err)
	}
	for hash, files := range searchResult.FoundFiles {
		if len(files) == 0 {
			continue
		}
		if err := r.restoreSingleFile(files[0], result); err != nil {
			r.logger.Error("Failed to restore %s: %v", hash, err)
			result.FailedFiles = append(result.FailedFiles, hash)
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", hash, err))
		}
	}
	r.logger.Info("Restoration complete: %d restored, %d failed, %d bytes",
		len(result.RestoredFiles), len(result.FailedFiles), result.TotalBytes)
	return result, nil
}

func (r *Restorer) restoreSingleFile(bf BackupFile, result *RestoreResult) error {
	targetPath := filepath.Join(r.assetsFolder, bf.RelativePath)
	if _, err := os.Stat(targetPath); err == nil {
		r.logger.Verbose("Already exists, skipping: %s", targetPath)
		result.RestoredFiles = append(result.RestoredFiles, bf.Hash)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}
	if err := r.copyFile(bf.Path, targetPath); err != nil {
		return err
	}
	result.RestoredFiles = append(result.RestoredFiles, bf.Hash)
	result.TotalBytes += bf.Size
	return nil
}

func (r *Restorer) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy content: %w", err)
	}
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("sync destination: %w", err)
	}
	return nil
}
