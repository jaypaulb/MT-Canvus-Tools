package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/logging"
)

// HashSearchResult holds the result of searching for a specific hash on disk.
type HashSearchResult struct {
	Found     bool
	FilePath  string
	Hash      string
	Extension string
	Size      int64
}

// SearchHashInAssetsFolder searches for a hash in the expected two-char subfolder
// structure: {assets_folder}/{hash[:2]}/{hash}.{ext}.
func SearchHashInAssetsFolder(assetsPath string, hash string) (*HashSearchResult, error) {
	logger := logging.GetLogger()
	if hash == "" {
		return &HashSearchResult{Found: false}, fmt.Errorf("hash cannot be empty")
	}
	if len(hash) < 2 {
		return &HashSearchResult{Found: false}, fmt.Errorf("hash must be at least 2 characters")
	}
	subfolderPath := filepath.Join(assetsPath, hash[:2])
	logger.Verbose("Searching for hash %q in: %s", hash, subfolderPath)
	if _, err := os.Stat(subfolderPath); os.IsNotExist(err) {
		return &HashSearchResult{Found: false}, nil
	}
	entries, err := os.ReadDir(subfolderPath)
	if err != nil {
		return &HashSearchResult{Found: false}, fmt.Errorf("read subfolder: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := filepath.Ext(name)
		fileHash := strings.TrimSuffix(name, ext)
		if fileHash == hash {
			fullPath := filepath.Join(subfolderPath, name)
			info, err := entry.Info()
			if err != nil {
				logger.Verbose("Failed to stat %s: %v", fullPath, err)
				continue
			}
			logger.Verbose("Found hash %q at: %s", hash, fullPath)
			return &HashSearchResult{
				Found:     true,
				FilePath:  fullPath,
				Hash:      hash,
				Extension: ext,
				Size:      info.Size(),
			}, nil
		}
	}
	return &HashSearchResult{Found: false}, nil
}

// SearchHashInAssetsFolderRecursive is a fallback that walks the entire tree.
func SearchHashInAssetsFolderRecursive(assetsPath string, hash string) (*HashSearchResult, error) {
	logger := logging.GetLogger()
	if hash == "" {
		return &HashSearchResult{Found: false}, fmt.Errorf("hash cannot be empty")
	}
	logger.Verbose("Recursively searching for hash %q in: %s", hash, assetsPath)
	var found *HashSearchResult
	err := filepath.Walk(assetsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(info.Name())
		fileHash := strings.TrimSuffix(info.Name(), ext)
		if fileHash == hash {
			found = &HashSearchResult{
				Found:     true,
				FilePath:  path,
				Hash:      hash,
				Extension: ext,
				Size:      info.Size(),
			}
			return filepath.SkipAll
		}
		return nil
	})
	if found != nil {
		return found, nil
	}
	if err != nil {
		return &HashSearchResult{Found: false}, err
	}
	return &HashSearchResult{Found: false}, nil
}
