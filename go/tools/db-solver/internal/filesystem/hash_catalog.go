package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/logging"
)

// HashCatalogEntry represents one file in the hash catalog.
type HashCatalogEntry struct {
	Hash         string
	FilePath     string
	Extension    string
	Size         int64
	ModifiedTime time.Time
}

// HashCatalog provides O(1) hash → file lookups built by a one-time directory scan.
type HashCatalog struct {
	entries map[string]HashCatalogEntry
	logger  *logging.Logger
}

// NewHashCatalog creates an empty HashCatalog.
func NewHashCatalog() *HashCatalog {
	return &HashCatalog{
		entries: make(map[string]HashCatalogEntry),
		logger:  logging.GetLogger(),
	}
}

// BuildCatalogFromFolder recursively scans folderPath and indexes every file by
// its filename stem (the hash). When duplicates are found the newer file wins.
func (c *HashCatalog) BuildCatalogFromFolder(folderPath string) error {
	c.logger.Info("Building hash catalog from: %s", folderPath)
	start := time.Now()

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			c.logger.Verbose("Error accessing %s: %v", path, err)
			return nil
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(info.Name())
		hash := strings.TrimSuffix(info.Name(), ext)
		if hash == "" {
			return nil
		}
		entry := HashCatalogEntry{
			Hash:         hash,
			FilePath:     path,
			Extension:    ext,
			Size:         info.Size(),
			ModifiedTime: info.ModTime(),
		}
		if existing, ok := c.entries[hash]; ok {
			if entry.ModifiedTime.After(existing.ModifiedTime) {
				c.entries[hash] = entry
			}
		} else {
			c.entries[hash] = entry
		}
		return nil
	})

	c.logger.Info("Catalog built: %d unique hashes indexed in %v", len(c.entries), time.Since(start))
	return err
}

// Lookup returns the catalog entry for a hash, if present.
func (c *HashCatalog) Lookup(hash string) (HashCatalogEntry, bool) {
	entry, ok := c.entries[hash]
	return entry, ok
}

// Size returns the number of entries in the catalog.
func (c *HashCatalog) Size() int { return len(c.entries) }

// Merge merges other into c, keeping the newer entry on collision.
func (c *HashCatalog) Merge(other *HashCatalog) {
	for hash, entry := range other.entries {
		if existing, ok := c.entries[hash]; ok {
			if entry.ModifiedTime.After(existing.ModifiedTime) {
				c.entries[hash] = entry
			}
		} else {
			c.entries[hash] = entry
		}
	}
}
