package paths

import (
	"os"
	"path/filepath"
)

// GetAppDataPath returns the Windows %APPDATA% path or equivalent.
func GetAppDataPath() (string, error) {
	return os.UserConfigDir()
}

// GetProgramDataPath returns the Windows %ProgramData% path or equivalent.
func GetProgramDataPath() (string, error) {
	// On Windows, this is %ProgramData%
	// On Linux, this would be /etc or similar
	if programData := os.Getenv("ProgramData"); programData != "" {
		return programData, nil
	}
	// Linux fallback - use /etc for system-wide config
	return "/etc", nil
}

// GetLocalAppDataPath returns the Windows %LOCALAPPDATA% path or equivalent.
func GetLocalAppDataPath() (string, error) {
	return os.UserCacheDir()
}

// GetCanvusUserConfigPath returns the path to user Canvus config directory.
func GetCanvusUserConfigPath() (string, error) {
	appData, err := GetAppDataPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(appData, "MultiTaction", "canvus"), nil
}

// GetCanvusSystemConfigPath returns the path to system Canvus config directory.
func GetCanvusSystemConfigPath() (string, error) {
	programData, err := GetProgramDataPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(programData, "MultiTaction", "canvus"), nil
}

// GetCanvusLogsPath returns the path to Canvus logs directory.
func GetCanvusLogsPath() (string, error) {
	localAppData, err := GetLocalAppDataPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(localAppData, "MultiTaction", "Canvus", "logs"), nil
}

// JoinPath joins path elements using the appropriate separator for the OS.
func JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}

// FileExists checks if a file exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if a path is a directory.
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
