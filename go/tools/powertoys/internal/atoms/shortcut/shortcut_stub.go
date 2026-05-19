//go:build !windows
// +build !windows

package shortcut

import "fmt"

// CreateDesktopShortcut creates a desktop shortcut (stub for non-Windows).
func CreateDesktopShortcut(name, target, arguments, workingDir string) error {
	return fmt.Errorf("desktop shortcuts are only supported on Windows")
}

