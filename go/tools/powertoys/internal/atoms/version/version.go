package version

import "fmt"

// Version information
var (
	Version     = "1.0.83"
	BuildDate   = "unknown"
	GitCommit   = "unknown"
	AppName     = "Canvus PowerToys"
	AppID       = "com.canvus.powertoys"
	Description = "Canvus PowerToys - Configuration and management tool for Multitaction Canvus"
)

// GetVersion returns the version string
func GetVersion() string {
	return Version
}

// GetFullVersion returns a full version string with build info
func GetFullVersion() string {
	return fmt.Sprintf("%s (Build: %s, Commit: %s)", Version, BuildDate, GitCommit)
}
