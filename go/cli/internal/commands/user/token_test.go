package user

import (
	"testing"
)

func TestTokenCmdStructure(t *testing.T) {
	// Test that tokenCmd is properly initialized
	if tokenCmd == nil {
		t.Fatal("tokenCmd is nil")
	}

	if tokenCmd.Use != "token" {
		t.Errorf("tokenCmd.Use = %q, want %q", tokenCmd.Use, "token")
	}

	if tokenCmd.Short == "" {
		t.Error("tokenCmd.Short is empty")
	}

	// Test that token subcommands are registered
	expectedCommands := []string{"create", "list"}
	for _, cmdName := range expectedCommands {
		found := false
		for _, cmd := range tokenCmd.Commands() {
			if cmd.Name() == cmdName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected token subcommand %q not found", cmdName)
		}
	}
}

func TestTokenCommandsRegistered(t *testing.T) {
	// Verify token command is registered under UserCmd
	found := false
	for _, cmd := range UserCmd.Commands() {
		if cmd.Name() == "token" {
			found = true
			break
		}
	}
	if !found {
		t.Error("token command not registered under UserCmd")
	}
}
