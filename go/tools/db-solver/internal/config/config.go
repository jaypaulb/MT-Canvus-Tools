// Package config loads and validates the db-solver tool configuration.
//
// Configuration precedence (highest to lowest):
//  1. Command-line flag  --config <path>
//  2. Environment variables with CANVUS_ prefix
//  3. YAML config file  (./config.yaml or $CANVUS_CONFIG)
//  4. Built-in defaults
//
// Viper is intentionally not used — a direct YAML + env loader keeps the
// dependency surface small and the precedence rules explicit.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds all settings for the db-solver tool.
type Config struct {
	CanvusServer CanvusServerConfig `yaml:"canvus_server"`
	Paths        PathsConfig        `yaml:"paths"`
	Logging      LoggingConfig      `yaml:"logging"`
	Performance  PerformanceConfig  `yaml:"performance"`
}

// CanvusServerConfig contains Canvus Server connection settings.
type CanvusServerConfig struct {
	// URL is the base URL for the Canvus server, e.g. "https://localhost:443".
	URL         string `yaml:"url"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	Timeout     int    `yaml:"timeout"`     // seconds
	InsecureTLS bool   `yaml:"insecuretls"` // skip TLS certificate verification
}

// PathsConfig contains file system paths.
type PathsConfig struct {
	AssetsFolder     string `yaml:"assetsfolder"`
	BackupRootFolder string `yaml:"backuprootfolder"`
	OutputFolder     string `yaml:"outputfolder"`
}

// LoggingConfig contains logging settings.
type LoggingConfig struct {
	Level     string `yaml:"level"`     // debug | info | warn | error
	Verbose   bool   `yaml:"verbose"`
	LogToFile bool   `yaml:"logtofile"`
	LogFile   string `yaml:"logfile"`
}

// PerformanceConfig contains performance tuning settings.
type PerformanceConfig struct {
	MaxConcurrentAPI     int `yaml:"maxconcurrentapi"`
	MaxConcurrentFiles   int `yaml:"maxconcurrentfiles"`
	APIRequestTimeout    int `yaml:"apirequesttimeout"`    // seconds
	FileOperationTimeout int `yaml:"fileoperationtimeout"` // seconds
}

// getDefaultPaths returns OS-appropriate default paths.
func getDefaultPaths() PathsConfig {
	if runtime.GOOS == "linux" {
		return PathsConfig{
			AssetsFolder:     "/var/lib/mt-canvus-server/assets",
			BackupRootFolder: "/var/lib/mt-canvus-server/backups",
			OutputFolder:     "./reports",
		}
	}
	return PathsConfig{
		AssetsFolder:     `C:\ProgramData\MultiTaction\canvus\assets`,
		BackupRootFolder: `C:\ProgramData\MultiTaction\canvus\backups`,
		OutputFolder:     `.\reports`,
	}
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		CanvusServer: CanvusServerConfig{
			URL:         "https://localhost:443",
			Timeout:     30,
			InsecureTLS: true,
		},
		Paths: getDefaultPaths(),
		Logging: LoggingConfig{
			Level:     "info",
			Verbose:   false,
			LogToFile: true,
			LogFile:   "canvus-server-db-solver.log",
		},
		Performance: PerformanceConfig{
			MaxConcurrentAPI:     10,
			MaxConcurrentFiles:   20,
			APIRequestTimeout:    30,
			FileOperationTimeout: 60,
		},
	}
}

// LoadConfig loads configuration from file and environment variables.
// configFile may be empty; if so the default search paths are used.
func LoadConfig(configFile string) (*Config, error) {
	cfg := DefaultConfig()

	// Determine config file path.
	if configFile == "" {
		configFile = os.Getenv("CANVUS_CONFIG")
	}
	if configFile == "" {
		// Try default locations.
		for _, candidate := range []string{
			"./config.yaml",
			"./config/config.yaml",
			filepath.Join(os.Getenv("HOME"), ".canvus-server-db-solver", "config.yaml"),
		} {
			if _, err := os.Stat(candidate); err == nil {
				configFile = candidate
				break
			}
		}
	}

	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("read config file %s: %w", configFile, err)
		}
		if err == nil {
			fmt.Printf("Using config file: %s\n", configFile)
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("parse config file %s: %w", configFile, err)
			}
		}
	}

	// Layer in environment variables.
	applyEnv(cfg)

	// Fill empty fields with defaults.
	cfg.applyDefaults()

	return cfg, nil
}

// applyEnv overlays CANVUS_* environment variables onto the config.
func applyEnv(cfg *Config) {
	if v := os.Getenv("CANVUS_BASE_URL"); v != "" {
		cfg.CanvusServer.URL = v
	}
	if v := os.Getenv("CANVUS_API_URL"); v != "" {
		// Preferred env var per Phase 4a unification.
		cfg.CanvusServer.URL = v
	}
	if v := os.Getenv("CANVUS_USERNAME"); v != "" {
		cfg.CanvusServer.Username = v
	}
	if v := os.Getenv("CANVUS_PASSWORD"); v != "" {
		cfg.CanvusServer.Password = v
	}
}

// applyDefaults fills zero-valued fields with defaults.
func (c *Config) applyDefaults() {
	d := DefaultConfig()
	if c.CanvusServer.URL == "" {
		c.CanvusServer.URL = d.CanvusServer.URL
	}
	if c.CanvusServer.Timeout == 0 {
		c.CanvusServer.Timeout = d.CanvusServer.Timeout
	}
	if c.Paths.AssetsFolder == "" {
		c.Paths.AssetsFolder = d.Paths.AssetsFolder
	}
	if c.Paths.BackupRootFolder == "" {
		c.Paths.BackupRootFolder = d.Paths.BackupRootFolder
	}
	if c.Paths.OutputFolder == "" {
		c.Paths.OutputFolder = d.Paths.OutputFolder
	}
	if c.Logging.Level == "" {
		c.Logging.Level = d.Logging.Level
	}
	if c.Logging.LogFile == "" {
		c.Logging.LogFile = d.Logging.LogFile
	}
	if c.Performance.MaxConcurrentAPI == 0 {
		c.Performance.MaxConcurrentAPI = d.Performance.MaxConcurrentAPI
	}
	if c.Performance.MaxConcurrentFiles == 0 {
		c.Performance.MaxConcurrentFiles = d.Performance.MaxConcurrentFiles
	}
	if c.Performance.APIRequestTimeout == 0 {
		c.Performance.APIRequestTimeout = d.Performance.APIRequestTimeout
	}
	if c.Performance.FileOperationTimeout == 0 {
		c.Performance.FileOperationTimeout = d.Performance.FileOperationTimeout
	}
}

// Validate validates the configuration, returning an error for invalid values.
func (c *Config) Validate() error {
	if c.CanvusServer.URL == "" {
		c.CanvusServer.URL = "https://localhost:443"
	}
	if c.CanvusServer.Username == "" {
		return fmt.Errorf("canvus server username is required")
	}
	if c.CanvusServer.Password == "" {
		return fmt.Errorf("canvus server password is required")
	}
	if c.Paths.AssetsFolder == "" {
		return fmt.Errorf("assets folder path is required")
	}
	if c.Paths.BackupRootFolder == "" {
		return fmt.Errorf("backup root folder path is required")
	}
	if !pathExists(c.Paths.AssetsFolder) {
		return fmt.Errorf("assets folder does not exist or is not accessible: %s", c.Paths.AssetsFolder)
	}
	if !pathExists(c.Paths.BackupRootFolder) {
		return fmt.Errorf("backup root folder does not exist or is not accessible: %s", c.Paths.BackupRootFolder)
	}
	if err := os.MkdirAll(c.Paths.OutputFolder, 0o755); err != nil {
		return fmt.Errorf("create output folder: %w", err)
	}
	validLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLevels, c.Logging.Level) {
		return fmt.Errorf("invalid logging level %q (must be one of: %s)", c.Logging.Level, strings.Join(validLevels, ", "))
	}
	if c.Performance.MaxConcurrentAPI < 1 {
		return fmt.Errorf("max concurrent API calls must be at least 1")
	}
	if c.Performance.MaxConcurrentFiles < 1 {
		return fmt.Errorf("max concurrent file operations must be at least 1")
	}
	return nil
}

// SaveConfig saves the configuration to a YAML file.
func (c *Config) SaveConfig(filename string) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

// GetCanvusAPIURL returns the full API URL (ensures /api/v1 suffix).
func (c *Config) GetCanvusAPIURL() string {
	url := strings.TrimSuffix(c.CanvusServer.URL, "/")
	if !strings.HasSuffix(url, "/api/v1") {
		url += "/api/v1"
	}
	return url
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
