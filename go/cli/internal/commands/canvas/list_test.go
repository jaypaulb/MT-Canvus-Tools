package canvas

import (
	"testing"
)

func TestListCmd_Registration(t *testing.T) {
	// Verify list command is registered
	found := false
	for _, cmd := range CanvasCmd.Commands() {
		if cmd.Name() == "list" {
			found = true
			break
		}
	}

	if !found {
		t.Error("list command should be registered with CanvasCmd")
	}
}

func TestListCmd_Flags(t *testing.T) {
	// Test that optional flags are defined
	if listCmd.Flags().Lookup("folder") == nil {
		t.Error("list command should have --folder flag")
	}

	if listCmd.Flags().Lookup("filter") == nil {
		t.Error("list command should have --filter flag")
	}
}

func TestListCmd_Args(t *testing.T) {
	// Test that list command accepts no arguments
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "with arguments",
			args:    []string{"unexpected"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := listCmd.Args(listCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("listCmd.Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListCmd_HelpText(t *testing.T) {
	// Verify command has proper help text
	if listCmd.Short == "" {
		t.Error("list command should have Short description")
	}

	if listCmd.Long == "" {
		t.Error("list command should have Long description")
	}

	// Check for examples in help text
	if !containsIgnoreCase(listCmd.Long, "example") {
		t.Error("list command Long description should include examples")
	}
}

func TestListCmd_UseText(t *testing.T) {
	// Verify Use field shows expected format
	expectedUse := "list"
	if listCmd.Use != expectedUse {
		t.Errorf("list command Use = %q, want %q", listCmd.Use, expectedUse)
	}
}
