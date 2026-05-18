package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigSet(t *testing.T) {
	// Create a temporary directory for test config
	tmpDir := t.TempDir()
	tmpConfig := filepath.Join(tmpDir, "config.yaml")

	// Override getConfigFilePath for testing
	originalGetConfigPath := getConfigFilePath
	getConfigFilePath = func() (string, error) {
		return tmpConfig, nil
	}
	defer func() { getConfigFilePath = originalGetConfigPath }()

	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{
			name:    "set valid string value",
			key:     "url",
			value:   "https://canvus.example.com",
			wantErr: false,
		},
		{
			name:    "set valid output format",
			key:     "output",
			value:   "json",
			wantErr: false,
		},
		{
			name:    "set valid boolean",
			key:     "insecure",
			value:   "true",
			wantErr: false,
		},
		{
			name:    "set valid integer",
			key:     "timeout",
			value:   "60",
			wantErr: false,
		},
		{
			name:    "invalid key",
			key:     "invalid_key",
			value:   "value",
			wantErr: true,
		},
		{
			name:    "invalid output format",
			key:     "output",
			value:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := setCmd
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs([]string{tt.key, tt.value})

			err := cmd.RunE(cmd, []string{tt.key, tt.value})

			if (err != nil) != tt.wantErr {
				t.Errorf("setCmd error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				output := buf.String()
				if !strings.Contains(output, "Configuration updated") {
					t.Errorf("expected success message in output, got: %s", output)
				}
			}
		})
	}
}

func TestConfigGet(t *testing.T) {
	// Create a temporary directory for test config
	tmpDir := t.TempDir()
	tmpConfig := filepath.Join(tmpDir, "config.yaml")

	// Override getConfigFilePath for testing
	originalGetConfigPath := getConfigFilePath
	getConfigFilePath = func() (string, error) {
		return tmpConfig, nil
	}
	defer func() { getConfigFilePath = originalGetConfigPath }()

	// Create test config file
	testConfig := `url: https://test.example.com
api_key: test-key-1234567890
output: json
`
	err := os.MkdirAll(filepath.Dir(tmpConfig), 0755)
	if err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	err = os.WriteFile(tmpConfig, []byte(testConfig), 0600)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	tests := []struct {
		name       string
		key        string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "get existing string value",
			key:        "url",
			wantErr:    false,
			wantOutput: "https://test.example.com",
		},
		{
			name:       "get sensitive value (masked)",
			key:        "api_key",
			wantErr:    false,
			wantOutput: "test***********7890", // 19 chars total: test (4) + 11 stars + 7890 (4)
		},
		{
			name:    "get missing key",
			key:     "username",
			wantErr: true,
		},
		{
			name:    "get invalid key",
			key:     "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := getCmd
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs([]string{tt.key})

			err := cmd.RunE(cmd, []string{tt.key})

			if (err != nil) != tt.wantErr {
				t.Errorf("getCmd error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				output := buf.String()
				if !strings.Contains(output, tt.wantOutput) {
					t.Errorf("expected output to contain %q, got: %s", tt.wantOutput, output)
				}
			}
		})
	}
}

func TestConfigList(t *testing.T) {
	// Create a temporary directory for test config
	tmpDir := t.TempDir()
	tmpConfig := filepath.Join(tmpDir, "config.yaml")

	// Override getConfigFilePath for testing
	originalGetConfigPath := getConfigFilePath
	getConfigFilePath = func() (string, error) {
		return tmpConfig, nil
	}
	defer func() { getConfigFilePath = originalGetConfigPath }()

	// Create test config file
	testConfig := `url: https://test.example.com
api_key: test-key-1234567890
output: json
timeout: 60
`
	err := os.MkdirAll(filepath.Dir(tmpConfig), 0755)
	if err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	err = os.WriteFile(tmpConfig, []byte(testConfig), 0600)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cmd := listCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})

	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("listCmd error = %v", err)
	}

	output := buf.String()

	// Check that all keys are present
	expectedKeys := []string{"url", "api_key", "output", "timeout"}
	for _, key := range expectedKeys {
		if !strings.Contains(output, key) {
			t.Errorf("expected output to contain key %q, got: %s", key, output)
		}
	}

	// Check that sensitive value is masked
	if strings.Contains(output, "test-key-1234567890") {
		t.Errorf("expected api_key to be masked, but found unmasked value in output")
	}
	if !strings.Contains(output, "test***********7890") {
		t.Errorf("expected api_key to be masked as 'test***********7890', got: %s", output)
	}
}

func TestMaskValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "empty string",
			value: "",
			want:  "",
		},
		{
			name:  "short value (8 or less)",
			value: "short",
			want:  "short", // Don't mask short values
		},
		{
			name:  "long value",
			value: "test-key-1234567890",
			want:  "test***********7890", // 19 chars: 4 + 11 stars + 4
		},
		{
			name:  "exactly 8 characters",
			value: "12345678",
			want:  "12345678", // Don't mask 8-char values
		},
		{
			name:  "exactly 9 characters",
			value: "123456789",
			want:  "1234*6789", // 9 chars: 4 + 1 star + 4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskValue(tt.value)
			if got != tt.want {
				t.Errorf("maskValue(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
