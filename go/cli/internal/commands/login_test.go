package commands

import (
	"strings"
	"testing"
)

func TestLoginCommandStructure(t *testing.T) {
	cmd := GetLoginCmd()

	if cmd.Use != "login" {
		t.Errorf("expected Use to be 'login', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	// Check that required flags exist
	usernameFlag := cmd.Flags().Lookup("username")
	if usernameFlag == nil {
		t.Error("expected --username flag to exist")
	}

	passwordFlag := cmd.Flags().Lookup("password")
	if passwordFlag == nil {
		t.Error("expected --password flag to exist")
	}
}

func TestLoginCommandHelp(t *testing.T) {
	cmd := GetLoginCmd()

	// Verify help text mentions security considerations
	if !strings.Contains(cmd.Long, "password") {
		t.Error("expected help text to mention password")
	}

	if !strings.Contains(cmd.Long, "username") {
		t.Error("expected help text to mention username")
	}
}

func TestLoginCommandFlags(t *testing.T) {
	cmd := GetLoginCmd()

	tests := []struct {
		name      string
		flagName  string
		shorthand string
	}{
		{
			name:      "username flag",
			flagName:  "username",
			shorthand: "u",
		},
		{
			name:      "password flag",
			flagName:  "password",
			shorthand: "p",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("expected --%s flag to exist", tt.flagName)
				return
			}

			if flag.Shorthand != tt.shorthand {
				t.Errorf("expected shorthand to be %q, got %q", tt.shorthand, flag.Shorthand)
			}
		})
	}
}

// Note: Full integration testing of the login command would require:
// - Mocking the SDK session and Login() method
// - Mocking file I/O for config saving
// - Mocking terminal input for password prompts
// These tests verify the command structure and flags are correct.
// The actual authentication logic is tested via the SDK's own tests.
