package custommenu

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/assets"
)

// GetIconSetPath returns the path to the icon set directory.
// If embedded icons are available, extracts them to a temp directory.
// Otherwise, returns the path to assets/icons relative to the executable.
func GetIconSetPath() (string, error) {
	// Try to use embedded icons first
	// Check if we can read from the embedded filesystem
	if _, err := fs.Stat(assets.Icons, "icons"); err == nil {
		// Extract embedded icons to a temporary directory
		tempDir := filepath.Join(os.TempDir(), "canvus-powertoys-icons")
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return "", err
		}

		// Walk embedded filesystem and extract files
		// The embed root is "icons", so we walk from "icons"
		err := fs.WalkDir(assets.Icons, "icons", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Remove "icons/" prefix for target path
			relPath := path
			if len(path) > 6 && path[:6] == "icons/" {
				relPath = path[6:]
			}

			targetPath := filepath.Join(tempDir, relPath)
			if d.IsDir() {
				return os.MkdirAll(targetPath, 0755)
			}

			// Read embedded file
			data, err := assets.Icons.ReadFile(path)
			if err != nil {
				return err
			}

			// Write to temp directory
			return os.WriteFile(targetPath, data, 0644)
		})

		if err == nil {
			return tempDir, nil
		}
	}

	// Fallback: try to find assets/icons relative to executable
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	exeDir := filepath.Dir(exePath)
	iconPath := filepath.Join(exeDir, "assets", "icons")
	if _, err := os.Stat(iconPath); err == nil {
		return iconPath, nil
	}

	// Try relative to current working directory
	cwd, _ := os.Getwd()
	iconPath = filepath.Join(cwd, "assets", "icons")
	if _, err := os.Stat(iconPath); err == nil {
		return iconPath, nil
	}

	return "", os.ErrNotExist
}
