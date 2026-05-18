// Package backup provides utilities for searching and restoring Canvus assets
// from administrator-supplied backup directories.
package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/logging"
)

// BackupFile represents a backup file that matches a missing asset hash.
type BackupFile struct {
	Path         string
	Hash         string
	Extension    string
	ModifiedTime time.Time
	Size         int64
	RelativePath string
}

// SearchResult holds the outcome of a backup search operation.
type SearchResult struct {
	FoundFiles    map[string][]BackupFile // hash → files, newest first after SortBackupFiles
	MissingHashes []string
	TotalSearched int
	TotalFiles    int
}

// Searcher walks a backup root directory looking for files whose name stems
// match the hashes of missing Canvus assets.
type Searcher struct {
	backupRootFolder string
	logger           *logging.Logger
}

// NewSearcher creates a Searcher rooted at backupRootFolder.
func NewSearcher(backupRootFolder string) *Searcher {
	return &Searcher{
		backupRootFolder: backupRootFolder,
		logger:           logging.GetLogger(),
	}
}

// SearchForAssets recursively searches the backup root for files whose filename
// stem matches one of the provided hashes.
func (s *Searcher) SearchForAssets(missingHashes []string) (*SearchResult, error) {
	result := &SearchResult{
		FoundFiles:    make(map[string][]BackupFile),
		MissingHashes: make([]string, 0),
	}
	if len(missingHashes) == 0 {
		s.logger.Info("No missing assets to search for")
		return result, nil
	}
	s.logger.Info("Searching for %d missing assets in: %s", len(missingHashes), s.backupRootFolder)

	if _, err := os.Stat(s.backupRootFolder); os.IsNotExist(err) {
		s.logger.Warn("Backup folder does not exist: %s", s.backupRootFolder)
		return result, nil
	}

	missingSet := make(map[string]bool, len(missingHashes))
	for _, h := range missingHashes {
		missingSet[h] = true
	}

	if err := s.searchRecursive(s.backupRootFolder, missingSet, result); err != nil {
		s.logger.Error("Error searching backup: %v", err)
		return result, fmt.Errorf("search backup: %w", err)
	}
	result.TotalSearched = 1

	for hash := range missingSet {
		if _, found := result.FoundFiles[hash]; !found {
			result.MissingHashes = append(result.MissingHashes, hash)
		}
	}

	s.logger.Info("Backup search completed: %d found / %d still missing",
		len(result.FoundFiles), len(result.MissingHashes))
	return result, nil
}

func (s *Searcher) searchRecursive(root string, missingSet map[string]bool, result *SearchResult) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			s.logger.Verbose("Error accessing %s: %v", path, err)
			return nil
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(info.Name())
		hash := strings.TrimSuffix(info.Name(), ext)
		if !missingSet[hash] {
			return nil
		}
		relPath, relErr := filepath.Rel(root, path)
		if relErr != nil {
			relPath = info.Name()
		}
		result.FoundFiles[hash] = append(result.FoundFiles[hash], BackupFile{
			Path:         path,
			Hash:         hash,
			Extension:    ext,
			ModifiedTime: info.ModTime(),
			Size:         info.Size(),
			RelativePath: relPath,
		})
		result.TotalFiles++
		return nil
	})
}

// SortBackupFiles sorts each hash's file list newest-first.
func (s *Searcher) SortBackupFiles(result *SearchResult) {
	for hash, files := range result.FoundFiles {
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModifiedTime.After(files[j].ModifiedTime)
		})
		result.FoundFiles[hash] = files
	}
}

// GetBestBackupFile returns the newest backup file for a hash, or nil.
func (s *Searcher) GetBestBackupFile(result *SearchResult, hash string) *BackupFile {
	files, ok := result.FoundFiles[hash]
	if !ok || len(files) == 0 {
		return nil
	}
	return &files[0]
}
