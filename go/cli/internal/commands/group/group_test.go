package group

import (
	"testing"
)

func TestValidateGroupID(t *testing.T) {
	tests := []struct {
		name    string
		groupID string
		wantErr bool
	}{
		{
			name:    "valid group ID",
			groupID: "group-123",
			wantErr: false,
		},
		{
			name:    "empty group ID",
			groupID: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGroupID(tt.groupID)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGroupID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGroupName(t *testing.T) {
	tests := []struct {
		name      string
		groupName string
		wantErr   bool
	}{
		{
			name:      "valid group name",
			groupName: "Development Team",
			wantErr:   false,
		},
		{
			name:      "empty group name",
			groupName: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGroupName(tt.groupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGroupName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGroupCmdStructure(t *testing.T) {
	// Test that GroupCmd is properly initialized
	if GroupCmd == nil {
		t.Fatal("GroupCmd is nil")
	}

	if GroupCmd.Use != "group" {
		t.Errorf("GroupCmd.Use = %q, want %q", GroupCmd.Use, "group")
	}

	if GroupCmd.Short == "" {
		t.Error("GroupCmd.Short is empty")
	}

	// Test that subcommands are registered
	expectedCommands := []string{"create", "list", "get", "delete", "add-user", "remove-user"}
	for _, cmdName := range expectedCommands {
		found := false
		for _, cmd := range GroupCmd.Commands() {
			if cmd.Name() == cmdName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand %q not found in GroupCmd", cmdName)
		}
	}
}
