package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfirmAction prompts the user for confirmation of an action.
// Returns true if the user confirms (y/yes), false otherwise.
// If force is true, the confirmation is skipped and true is returned.
func ConfirmAction(message string, force bool) (bool, error) {
	// Skip prompt if force flag is set
	if force {
		return true, nil
	}

	// Check if running in non-interactive environment
	if !isInteractive() {
		return false, fmt.Errorf("cannot prompt for confirmation in non-interactive environment (use --force to skip)")
	}

	// Prompt user for confirmation
	fmt.Printf("%s [y/N]: ", message)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}

	// Parse response
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

// isInteractive checks if the program is running in an interactive terminal.
func isInteractive() bool {
	// Check if stdin is a terminal
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// PromptForPassword prompts the user to enter a password without echoing it.
// This is a placeholder - actual implementation would use golang.org/x/term
func PromptForPassword(message string) (string, error) {
	fmt.Printf("%s: ", message)
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return strings.TrimSpace(password), nil
}
