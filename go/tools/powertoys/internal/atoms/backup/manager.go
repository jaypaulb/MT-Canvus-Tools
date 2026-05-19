package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	// MaxBackups is the maximum number of backups to keep
	MaxBackups = 5
	// BackupSuffix is the suffix for backup files
	BackupSuffix = ".bak"
)

// Manager handles file backups with rotation.
type Manager struct {
	backupDir string
}

// NewManager creates a new backup manager.
func NewManager(backupDir string) *Manager {
	return &Manager{
		backupDir: backupDir,
	}
}

// CreateBackup creates a backup of a file with date suffix in the same folder.
// Returns the backup path if created, empty string if no backup needed.
func (m *Manager) CreateBackup(filePath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// No file to backup
		return "", nil
	}

	// Read original file
	originalData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Generate backup filename with date suffix in same folder
	dir := filepath.Dir(filePath)
	baseName := filepath.Base(filePath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := baseName[:len(baseName)-len(ext)]

	// Format: filename.2024-01-15.ext
	dateStr := time.Now().Format("2006-01-02")
	backupName := fmt.Sprintf("%s.%s%s", nameWithoutExt, dateStr, ext)
	backupPath := filepath.Join(dir, backupName)

	// Check if backup already exists for today
	if _, err := os.Stat(backupPath); err == nil {
		// Backup for today already exists, skip
		return "", nil
	}

	// Create backup file
	if err := os.WriteFile(backupPath, originalData, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	return backupPath, nil
}


// RestoreBackup restores a file from a backup.
func (m *Manager) RestoreBackup(backupPath, targetPath string) error {
	src, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy backup: %w", err)
	}

	return nil
}
