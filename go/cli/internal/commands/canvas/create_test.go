package canvas

import (
	"context"
	"fmt"
	"strings"
	"testing"

	cmdtest "github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/testing"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func TestCreateCmd_SuccessfulCreation(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		folderID     string
		expectedReq  canvus.CreateCanvasRequest
		mockResponse *canvus.Canvas
		mockError    error
	}{
		{
			name: "create canvas with name only",
			args: []string{"Test Canvas"},
			expectedReq: canvus.CreateCanvasRequest{
				Name: "Test Canvas",
			},
			mockResponse: &canvus.Canvas{
				ID:   "canvas-123",
				Name: "Test Canvas",
			},
		},
		{
			name:     "create canvas with folder",
			args:     []string{"Test Canvas"},
			folderID: "folder-456",
			expectedReq: canvus.CreateCanvasRequest{
				Name:     "Test Canvas",
				FolderID: "folder-456",
			},
			mockResponse: &canvus.Canvas{
				ID:       "canvas-123",
				Name:     "Test Canvas",
				FolderID: "folder-456",
			},
		},
		{
			name: "create canvas with special characters in name",
			args: []string{"Test Canvas: Project #1 (v2.0)"},
			expectedReq: canvus.CreateCanvasRequest{
				Name: "Test Canvas: Project #1 (v2.0)",
			},
			mockResponse: &canvus.Canvas{
				ID:   "canvas-789",
				Name: "Test Canvas: Project #1 (v2.0)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset folder flag
			createFolderID = tt.folderID

			// Create mock session
			var capturedReq canvus.CreateCanvasRequest
			sdkCalled := false
			mockSess := &cmdtest.MockSession{
				CreateCanvasFunc: func(ctx context.Context, req any) (*canvus.Canvas, error) {
					capturedReq = req.(canvus.CreateCanvasRequest)
					sdkCalled = true
					return tt.mockResponse, tt.mockError
				},
			}

			// Create context with mock session provider and config
			ctx := context.Background()
			ctx = session.WithSessionProvider(ctx, mockSess)
			cfg := &config.Config{Output: "json"}
			ctx = config.WithConfig(ctx, cfg)

			// Set up command
			createCmd.SetContext(ctx)

			// Execute command
			err := createCmd.RunE(createCmd, tt.args)

			// Verify no error
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			// Verify SDK was called
			if !sdkCalled {
				t.Fatal("expected SDK CreateCanvas to be called")
			}

			// Verify SDK was called with correct request
			if capturedReq.Name != tt.expectedReq.Name {
				t.Errorf("expected name %q, got %q", tt.expectedReq.Name, capturedReq.Name)
			}
			if capturedReq.FolderID != tt.expectedReq.FolderID {
				t.Errorf("expected folderID %q, got %q", tt.expectedReq.FolderID, capturedReq.FolderID)
			}

			// Reset folder flag
			createFolderID = ""
		})
	}
}

func TestCreateCmd_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		mockError     error
		expectedError string
	}{
		{
			name:          "SDK error - authentication failed",
			args:          []string{"Test Canvas"},
			mockError:     fmt.Errorf("authentication failed: invalid API key"),
			expectedError: "authentication failed",
		},
		{
			name:          "SDK error - permission denied",
			args:          []string{"Test Canvas"},
			mockError:     fmt.Errorf("permission denied"),
			expectedError: "permission denied",
		},
		{
			name:          "SDK error - network error",
			args:          []string{"Test Canvas"},
			mockError:     fmt.Errorf("connection refused"),
			expectedError: "connection refused",
		},
		{
			name:          "SDK error - conflict",
			args:          []string{"Test Canvas"},
			mockError:     fmt.Errorf("canvas with this name already exists"),
			expectedError: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset folder flag
			createFolderID = ""

			// Create mock session that returns error
			sdkCalled := false
			mockSess := &cmdtest.MockSession{
				CreateCanvasFunc: func(ctx context.Context, req any) (*canvus.Canvas, error) {
					sdkCalled = true
					return nil, tt.mockError
				},
			}

			// Create context with mock session provider and config
			ctx := context.Background()
			ctx = session.WithSessionProvider(ctx, mockSess)
			cfg := &config.Config{Output: "json"}
			ctx = config.WithConfig(ctx, cfg)

			// Set up command
			createCmd.SetContext(ctx)

			// Execute command
			err := createCmd.RunE(createCmd, tt.args)

			// Verify SDK was called
			if !sdkCalled {
				t.Fatal("expected SDK CreateCanvas to be called")
			}

			// Verify error occurred
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			// Verify error message
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("expected error containing %q, got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestCreateCmd_InputValidation(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{
			name:          "no arguments",
			args:          []string{},
			expectedError: "accepts 1 arg(s), received 0",
		},
		{
			name:          "too many arguments",
			args:          []string{"Canvas1", "Canvas2"},
			expectedError: "accepts 1 arg(s), received 2",
		},
		{
			name:          "empty canvas name",
			args:          []string{""},
			expectedError: "canvas name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test argument validation
			if len(tt.args) != 1 {
				err := createCmd.Args(createCmd, tt.args)
				if err == nil {
					t.Fatal("expected error for invalid args, got nil")
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got: %v", tt.expectedError, err)
				}
				return
			}

			// Test name validation
			err := validateCanvasName(tt.args[0])
			if err == nil {
				t.Fatal("expected error for invalid name, got nil")
			}
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("expected error containing %q, got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestCreateCmd_FlagParsing(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		folderFlag     string
		expectedName   string
		expectedFolder string
	}{
		{
			name:           "with folder flag",
			args:           []string{"My Canvas"},
			folderFlag:     "folder-123",
			expectedName:   "My Canvas",
			expectedFolder: "folder-123",
		},
		{
			name:           "without folder flag",
			args:           []string{"My Canvas"},
			folderFlag:     "",
			expectedName:   "My Canvas",
			expectedFolder: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set folder flag
			createFolderID = tt.folderFlag

			// Create mock session
			var capturedReq canvus.CreateCanvasRequest
			sdkCalled := false
			mockSess := &cmdtest.MockSession{
				CreateCanvasFunc: func(ctx context.Context, req any) (*canvus.Canvas, error) {
					rr := req.(canvus.CreateCanvasRequest)
					capturedReq = rr
					sdkCalled = true
					return &canvus.Canvas{
						ID:       "canvas-123",
						Name:     rr.Name,
						FolderID: rr.FolderID,
					}, nil
				},
			}

			// Create context with mock session provider and config
			ctx := context.Background()
			ctx = session.WithSessionProvider(ctx, mockSess)
			cfg := &config.Config{Output: "json"}
			ctx = config.WithConfig(ctx, cfg)

			// Set up command
			createCmd.SetContext(ctx)

			// Execute command
			err := createCmd.RunE(createCmd, tt.args)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			// Verify SDK was called
			if !sdkCalled {
				t.Fatal("expected SDK CreateCanvas to be called")
			}

			// Verify request parameters
			if capturedReq.Name != tt.expectedName {
				t.Errorf("expected name %q, got %q", tt.expectedName, capturedReq.Name)
			}
			if capturedReq.FolderID != tt.expectedFolder {
				t.Errorf("expected folderID %q, got %q", tt.expectedFolder, capturedReq.FolderID)
			}

			// Reset folder flag
			createFolderID = ""
		})
	}
}

func TestCreateCmd_SDKInteraction(t *testing.T) {
	// Test that verifies proper SDK interaction patterns
	tests := []struct {
		name          string
		args          []string
		mockResponse  *canvus.Canvas
		mockError     error
		expectSDKCall bool
		expectError   bool
	}{
		{
			name: "successful SDK call returns canvas",
			args: []string{"Test Canvas"},
			mockResponse: &canvus.Canvas{
				ID:   "canvas-123",
				Name: "Test Canvas",
			},
			mockError:     nil,
			expectSDKCall: true,
			expectError:   false,
		},
		{
			name:          "SDK error propagates correctly",
			args:          []string{"Test Canvas"},
			mockResponse:  nil,
			mockError:     fmt.Errorf("API error: server unavailable"),
			expectSDKCall: true,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			createFolderID = ""

			sdkCalled := false
			mockSess := &cmdtest.MockSession{
				CreateCanvasFunc: func(ctx context.Context, req any) (*canvus.Canvas, error) {
					sdkCalled = true
					return tt.mockResponse, tt.mockError
				},
			}

			ctx := context.Background()
			ctx = session.WithSessionProvider(ctx, mockSess)
			cfg := &config.Config{Output: "json"}
			ctx = config.WithConfig(ctx, cfg)

			createCmd.SetContext(ctx)

			err := createCmd.RunE(createCmd, tt.args)

			if sdkCalled != tt.expectSDKCall {
				t.Errorf("expected SDK called=%v, got %v", tt.expectSDKCall, sdkCalled)
			}

			if (err != nil) != tt.expectError {
				t.Errorf("expected error=%v, got error=%v", tt.expectError, err)
			}
		})
	}
}

// Legacy tests for command registration and structure
func TestCreateCmd_Registration(t *testing.T) {
	// Verify create command is registered
	found := false
	for _, cmd := range CanvasCmd.Commands() {
		if cmd.Name() == "create" {
			found = true
			break
		}
	}

	if !found {
		t.Error("create command should be registered with CanvasCmd")
	}
}

func TestCreateCmd_Flags(t *testing.T) {
	// Test that required flags are defined
	if createCmd.Flags().Lookup("folder") == nil {
		t.Error("create command should have --folder flag")
	}
}

func TestCreateCmd_Args(t *testing.T) {
	// Test argument validation
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid single argument",
			args:    []string{"My Canvas"},
			wantErr: false,
		},
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many arguments",
			args:    []string{"Canvas1", "Canvas2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createCmd.Args(createCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("createCmd.Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateCmd_HelpText(t *testing.T) {
	// Verify command has proper help text
	if createCmd.Short == "" {
		t.Error("create command should have Short description")
	}

	if createCmd.Long == "" {
		t.Error("create command should have Long description")
	}

	// Check for examples in help text
	if !containsIgnoreCase(createCmd.Long, "example") {
		t.Error("create command Long description should include examples")
	}
}

func TestCreateCmd_UseText(t *testing.T) {
	// Verify Use field shows expected format
	expectedUse := "create <name>"
	if createCmd.Use != expectedUse {
		t.Errorf("create command Use = %q, want %q", createCmd.Use, expectedUse)
	}
}
