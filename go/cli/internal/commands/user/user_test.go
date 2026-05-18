package user

import (
	"testing"
)

func TestValidateUserID(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{
			name:    "valid user ID",
			userID:  "user-123",
			wantErr: false,
		},
		{
			name:    "empty user ID",
			userID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUserID(tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUserID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
		},
		{
			name:    "email without @",
			email:   "userexample.com",
			wantErr: true,
		},
		{
			name:    "email starting with @",
			email:   "@example.com",
			wantErr: true,
		},
		{
			name:    "email ending with @",
			email:   "user@",
			wantErr: true,
		},
		{
			name:    "too short",
			email:   "a@",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{
			name:     "valid username",
			username: "jdoe",
			wantErr:  false,
		},
		{
			name:     "empty username",
			username: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name     string
		userName string
		wantErr  bool
	}{
		{
			name:     "valid name",
			userName: "John Doe",
			wantErr:  false,
		},
		{
			name:     "empty name",
			userName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateName(tt.userName)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserCmdStructure(t *testing.T) {
	// Test that UserCmd is properly initialized
	if UserCmd == nil {
		t.Fatal("UserCmd is nil")
	}

	if UserCmd.Use != "user" {
		t.Errorf("UserCmd.Use = %q, want %q", UserCmd.Use, "user")
	}

	if UserCmd.Short == "" {
		t.Error("UserCmd.Short is empty")
	}

	// Test that subcommands are registered
	expectedCommands := []string{"create", "list", "get", "update", "delete"}
	for _, cmdName := range expectedCommands {
		found := false
		for _, cmd := range UserCmd.Commands() {
			if cmd.Name() == cmdName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand %q not found in UserCmd", cmdName)
		}
	}
}
