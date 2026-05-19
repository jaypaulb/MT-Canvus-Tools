package logger

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

var (
	// IsConsole indicates if the app is running from a console/terminal
	IsConsole bool
)

func init() {
	// Check if stdout is connected to a terminal
	// On Windows cmd.exe, stdout might not be detected as a character device
	// So we use a more permissive check

	// First, try the standard check
	fileInfo, err := os.Stdout.Stat()
	if err == nil {
		mode := fileInfo.Mode()
		IsConsole = (mode & os.ModeCharDevice) != 0
	}

	// On Windows, also check if we can write to stderr (which is more reliable)
	// If stdout check failed or we're on Windows, try stderr
	if !IsConsole || runtime.GOOS == "windows" {
		stderrInfo, err := os.Stderr.Stat()
		if err == nil {
			mode := stderrInfo.Mode()
			if (mode & os.ModeCharDevice) != 0 {
				IsConsole = true
			}
		}
	}

	// As a fallback on Windows, check for console environment variable
	// or try to detect if we're running from cmd.exe
	if runtime.GOOS == "windows" && !IsConsole {
		// Check if TERM or other console indicators exist
		// Or check if we can actually write (more permissive)
		// For now, be permissive on Windows - always enable if we can't determine
		// This helps with debugging
		IsConsole = true // Enable by default on Windows for debugging
	}
}

// Log prints a message to console if running from terminal
func Log(message string) {
	if IsConsole {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		fmt.Printf("[%s] %s\n", timestamp, message)
	}
}

// Logf prints a formatted message to console if running from terminal
func Logf(format string, args ...interface{}) {
	if IsConsole {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		message := fmt.Sprintf(format, args...)
		fmt.Printf("[%s] %s\n", timestamp, message)
	}
}

