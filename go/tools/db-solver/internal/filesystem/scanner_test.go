package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractHashFromFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantHash string
	}{
		{"valid hash with extension", "abcd1234ef567890.jpg", "abcd1234ef567890"},
		{"no extension", "abcd1234ef567890", ""},
		{"hash too short", "abc.jpg", ""},
		{"hash too long", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.jpg", ""},
		{"invalid char in hash", "abcd-1234ef567890.jpg", ""},
		{"valid 64-char hash", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.jpg", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHashFromFilename(tt.input)
			assert.Equal(t, tt.wantHash, got)
		})
	}
}

func TestScanAssetsFolder(t *testing.T) {
	dir := t.TempDir()

	// Create a valid hash-named file in a two-char subfolder.
	sub := filepath.Join(dir, "ab")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	validHash := "abcdef1234567890"
	require.NoError(t, os.WriteFile(filepath.Join(sub, validHash+".jpg"), []byte("data"), 0o644))

	// Create an invalid file that should be skipped.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("x"), 0o644))

	result, err := ScanAssetsFolder(dir)
	require.NoError(t, err)
	assert.Equal(t, 1, len(result.Files))
	_, found := result.HashMap[validHash]
	assert.True(t, found, "valid hash should be indexed")
}

func TestFindMissingAssets(t *testing.T) {
	result := &ScanResult{
		HashMap: map[string]FileInfo{
			"hash1": {},
			"hash2": {},
		},
	}
	discovered := []string{"hash1", "hash2", "hash3", "hash4"}
	missing := FindMissingAssets(discovered, result)
	assert.ElementsMatch(t, []string{"hash3", "hash4"}, missing)
}

func TestScanAssetsFolderNotExist(t *testing.T) {
	_, err := ScanAssetsFolder("/nonexistent/path/that/cannot/exist")
	assert.Error(t, err)
}
