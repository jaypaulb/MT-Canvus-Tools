//go:build windows
// +build windows

package shortcut

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CreateDesktopShortcut creates a Windows desktop shortcut with the specified target and arguments.
func CreateDesktopShortcut(name, target, arguments, workingDir string) error {
	// Get desktop path
	desktopPath, err := getDesktopPath()
	if err != nil {
		return fmt.Errorf("failed to get desktop path: %w", err)
	}

	shortcutPath := filepath.Join(desktopPath, name+".lnk")

	// Use PowerShell to create shortcut
	// PowerShell script to create a shortcut
	psScript := fmt.Sprintf(`
$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut("%s")
$Shortcut.TargetPath = "%s"
$Shortcut.Arguments = "%s"
$Shortcut.WorkingDirectory = "%s"
$Shortcut.Save()
`, strings.ReplaceAll(shortcutPath, `\`, `\\`),
		strings.ReplaceAll(target, `\`, `\\`),
		strings.ReplaceAll(arguments, `"`, `\"`),
		strings.ReplaceAll(workingDir, `\`, `\\`))

	// Execute PowerShell
	cmd := exec.Command("powershell", "-Command", psScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create shortcut: %w", err)
	}

	return nil
}

// getDesktopPath returns the Windows desktop path.
func getDesktopPath() (string, error) {
	// Try common desktop paths
	userProfile := os.Getenv("USERPROFILE")
	if userProfile != "" {
		desktop := filepath.Join(userProfile, "Desktop")
		if _, err := os.Stat(desktop); err == nil {
			return desktop, nil
		}
	}

	// Try alternative location
	desktop := os.Getenv("USERPROFILE") + "\\Desktop"
	if _, err := os.Stat(desktop); err == nil {
		return desktop, nil
	}

	return "", fmt.Errorf("desktop path not found")
}

