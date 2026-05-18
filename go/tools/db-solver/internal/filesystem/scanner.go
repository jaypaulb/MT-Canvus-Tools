// Package filesystem provides utilities for scanning the Canvus assets folder.
package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileInfo represents information about a file in the assets folder.
type FileInfo struct {
	Path         string `json:"path"`
	Hash         string `json:"hash"`
	Filename     string `json:"filename"`
	Size         int64  `json:"size"`
	RelativePath string `json:"relative_path"`
}

// ScanResult holds the result of scanning the assets folder.
type ScanResult struct {
	Files     []FileInfo          `json:"files"`
	HashMap   map[string]FileInfo `json:"hash_map"`
	TotalSize int64               `json:"total_size"`
}

// ScanAssetsFolder walks the assets folder and builds a hash → FileInfo map.
func ScanAssetsFolder(assetsPath string) (*ScanResult, error) {
	result := &ScanResult{
		Files:   make([]FileInfo, 0),
		HashMap: make(map[string]FileInfo),
	}
	if _, err := os.Stat(assetsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("assets folder does not exist: %s", assetsPath)
	}
	err := filepath.Walk(assetsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		hash := extractHashFromFilename(info.Name())
		if hash == "" {
			return nil
		}
		relPath, relErr := filepath.Rel(assetsPath, path)
		if relErr != nil {
			relPath = info.Name()
		}
		fi := FileInfo{
			Path:         path,
			Hash:         hash,
			Filename:     info.Name(),
			Size:         info.Size(),
			RelativePath: relPath,
		}
		result.Files = append(result.Files, fi)
		result.HashMap[hash] = fi
		result.TotalSize += info.Size()
		return nil
	})
	if err != nil {
		return result, err
	}
	return result, nil
}

// FindMissingAssets returns hashes from discoveredAssets that are absent from scanResult.
func FindMissingAssets(discoveredAssets []string, scanResult *ScanResult) []string {
	var missing []string
	for _, hash := range discoveredAssets {
		if _, exists := scanResult.HashMap[hash]; !exists {
			missing = append(missing, hash)
		}
	}
	return missing
}

// extractHashFromFilename strips the extension and validates the remaining string
// as a hash-like token (alphanumeric, 8–64 chars).
func extractHashFromFilename(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ""
	}
	hash := strings.TrimSuffix(filename, ext)
	if len(hash) < 8 || len(hash) > 64 {
		return ""
	}
	for _, ch := range hash {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
			return ""
		}
	}
	return hash
}
