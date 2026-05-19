package webui

import (
	"os"
	"runtime"
)

// GetDeviceName returns the device hostname for use as installation_name fallback.
// On Windows: Returns computer name from os.Hostname()
// On Linux: Returns hostname from os.Hostname()
func GetDeviceName() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return hostname, nil
}

// GetOSInfo returns OS-specific information for debugging.
func GetOSInfo() string {
	return runtime.GOOS
}
