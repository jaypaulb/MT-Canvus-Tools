package system

import (
	"testing"
)

func TestSystemCmdStructure(t *testing.T) {
	// Test that SystemCmd is properly initialized
	if SystemCmd == nil {
		t.Fatal("SystemCmd is nil")
	}

	if SystemCmd.Use != "system" {
		t.Errorf("SystemCmd.Use = %q, want %q", SystemCmd.Use, "system")
	}

	if SystemCmd.Short == "" {
		t.Error("SystemCmd.Short is empty")
	}

	// Test that subcommands are registered
	expectedCommands := []string{"info", "license"}
	for _, cmdName := range expectedCommands {
		found := false
		for _, cmd := range SystemCmd.Commands() {
			if cmd.Name() == cmdName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand %q not found in SystemCmd", cmdName)
		}
	}
}

func TestInfoCommandStructure(t *testing.T) {
	if infoCmd == nil {
		t.Fatal("infoCmd is nil")
	}

	if infoCmd.Use != "info" {
		t.Errorf("infoCmd.Use = %q, want %q", infoCmd.Use, "info")
	}

	if infoCmd.Short == "" {
		t.Error("infoCmd.Short is empty")
	}

	if infoCmd.RunE == nil {
		t.Error("infoCmd.RunE is nil")
	}
}

func TestLicenseCommandStructure(t *testing.T) {
	if licenseCmd == nil {
		t.Fatal("licenseCmd is nil")
	}

	if licenseCmd.Use != "license" {
		t.Errorf("licenseCmd.Use = %q, want %q", licenseCmd.Use, "license")
	}

	if licenseCmd.Short == "" {
		t.Error("licenseCmd.Short is empty")
	}

	if licenseCmd.RunE == nil {
		t.Error("licenseCmd.RunE is nil")
	}
}
