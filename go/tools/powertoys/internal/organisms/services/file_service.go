package services

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/paths"
)

// FileService handles file detection and path operations.
type FileService struct {
	userConfigPath   string
	systemConfigPath string
}

// NewFileService creates a new file service instance.
func NewFileService() (*FileService, error) {
	userPath, err := paths.GetCanvusUserConfigPath()
	if err != nil {
		return nil, err
	}

	systemPath, err := paths.GetCanvusSystemConfigPath()
	if err != nil {
		return nil, err
	}

	return &FileService{
		userConfigPath:   userPath,
		systemConfigPath: systemPath,
	}, nil
}

// DetectMtCanvusIni attempts to find mt-canvus.ini in standard locations.
// Returns the path if found, empty string if not found.
func (fs *FileService) DetectMtCanvusIni() string {
	// Check user config location first
	userIni := filepath.Join(fs.userConfigPath, "mt-canvus.ini")
	if paths.FileExists(userIni) {
		return userIni
	}

	// Check system config location
	systemIni := filepath.Join(fs.systemConfigPath, "mt-canvus.ini")
	if paths.FileExists(systemIni) {
		return systemIni
	}

	return ""
}

// DetectScreenXml attempts to find screen.xml in standard locations.
// Returns the path if found, empty string if not found.
func (fs *FileService) DetectScreenXml() string {
	// Check user config location first
	userXml := filepath.Join(fs.userConfigPath, "screen.xml")
	if paths.FileExists(userXml) {
		return userXml
	}

	// Check system config location
	systemXml := filepath.Join(fs.systemConfigPath, "screen.xml")
	if paths.FileExists(systemXml) {
		return systemXml
	}

	return ""
}

// DetectMenuYml attempts to find menu.yml in standard locations.
// Returns the path if found, empty string if not found.
func (fs *FileService) DetectMenuYml() string {
	// Check CustomMenu directory first
	userCustomMenu := filepath.Join(fs.userConfigPath, "CustomMenu", "menu.yml")
	if paths.FileExists(userCustomMenu) {
		return userCustomMenu
	}

	// Check user config location first
	userYml := filepath.Join(fs.userConfigPath, "menu.yml")
	if paths.FileExists(userYml) {
		return userYml
	}

	// Check system config CustomMenu directory
	systemCustomMenu := filepath.Join(fs.systemConfigPath, "CustomMenu", "menu.yml")
	if paths.FileExists(systemCustomMenu) {
		return systemCustomMenu
	}

	// Check legacy system config location
	systemYml := filepath.Join(fs.systemConfigPath, "menu.yml")
	if paths.FileExists(systemYml) {
		return systemYml
	}

	return ""
}

// GetUserConfigPath returns the user config directory path.
func (fs *FileService) GetUserConfigPath() string {
	return fs.userConfigPath
}

// GetSystemConfigPath returns the system config directory path.
func (fs *FileService) GetSystemConfigPath() string {
	return fs.systemConfigPath
}

// GetDefaultScreenXmlPath returns the default screen.xml path.
// On Windows: %APPDATA%\MultiTouch\screen.xml
// On Linux: ~/.MultiTouch/screen.xml
func (fs *FileService) GetDefaultScreenXmlPath() (string, error) {
	appData, err := paths.GetAppDataPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(appData, "MultiTouch", "screen.xml"), nil
}

// DetectExampleIni attempts to find the example mt-canvus.ini file.
// On Windows: "C:\Program Files\MT Canvus\Examples\mt-canvus.ini"
// On Linux: "/usr/share/MT Canvus/Examples/mt-canvus.ini" or similar
// Returns the path if found, empty string if not found.
func (fs *FileService) DetectExampleIni() string {
	// Try Windows path first
	windowsPath := `C:\Program Files\MT Canvus\Examples\mt-canvus.ini`
	if paths.FileExists(windowsPath) {
		return windowsPath
	}

	// Try common Linux paths
	linuxPaths := []string{
		"/usr/share/MT Canvus/Examples/mt-canvus.ini",
		"/opt/MT Canvus/Examples/mt-canvus.ini",
		"/usr/local/share/MT Canvus/Examples/mt-canvus.ini",
	}
	for _, p := range linuxPaths {
		if paths.FileExists(p) {
			return p
		}
	}

	// Try relative to executable location (for portable installations)
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		examplePath := filepath.Join(execDir, "Examples", "mt-canvus.ini")
		if paths.FileExists(examplePath) {
			return examplePath
		}
		// Also try parent directory
		parentExamplePath := filepath.Join(filepath.Dir(execDir), "Examples", "mt-canvus.ini")
		if paths.FileExists(parentExamplePath) {
			return parentExamplePath
		}
	}

	return ""
}

// EnsureDirectory creates a directory if it doesn't exist.
func (fs *FileService) EnsureDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// ReadJSONFile reads a JSON file and unmarshals it into the provided value.
func (fs *FileService) ReadJSONFile(filePath string, v interface{}) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist - return empty/default value
			return nil
		}
		return err
	}

	if len(data) == 0 {
		// Empty file - return empty/default value
		return nil
	}

	return json.Unmarshal(data, v)
}

// WriteJSONFile writes a value to a JSON file.
func (fs *FileService) WriteJSONFile(filePath string, v interface{}) error {
	// Create parent directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}
