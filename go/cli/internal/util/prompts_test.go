package util

import (
	"testing"
)

func TestConfirmAction_Force(t *testing.T) {
	// When force is true, should always return true without prompting
	confirmed, err := ConfirmAction("Test message", true)
	if err != nil {
		t.Errorf("ConfirmAction() with force=true returned error: %v", err)
	}
	if !confirmed {
		t.Error("ConfirmAction() with force=true should return true")
	}
}

func TestIsInteractive(t *testing.T) {
	// This test just ensures the function doesn't panic
	// Actual behavior depends on the environment
	_ = isInteractive()
}

func TestPromptForPassword(t *testing.T) {
	// This test ensures the function signature is correct
	// Actual testing would require mocking stdin
	t.Skip("Skipping interactive test - requires stdin mocking")
}
