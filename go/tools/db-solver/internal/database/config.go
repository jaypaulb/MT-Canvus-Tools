// Package database provides PostgreSQL connectivity for the db-solver tool.
package database

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/logging"
	"gopkg.in/ini.v1"
)

// DBConfig holds database connection parameters loaded from mt-canvus-server.ini.
type DBConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	SSLMode  string
}

func getDefaultINIPath() string {
	if runtime.GOOS == "linux" {
		return "/etc/MultiTaction/canvus/mt-canvus-server.ini"
	}
	return `C:\ProgramData\MultiTaction\Canvus\mt-canvus-server.ini`
}

// FindINIFile attempts to find the mt-canvus-server.ini file in well-known locations.
func FindINIFile() (string, error) {
	var candidates []string
	if runtime.GOOS == "linux" {
		candidates = []string{
			"/etc/MultiTaction/canvus/mt-canvus-server.ini",
			"/etc/multitaction/canvus/mt-canvus-server.ini",
			"/opt/MultiTaction/canvus/mt-canvus-server.ini",
			"./mt-canvus-server.ini",
			"mt-canvus-server.ini",
		}
	} else {
		candidates = []string{
			`C:\ProgramData\MultiTaction\Canvus\mt-canvus-server.ini`,
			`C:\Program Files\MultiTaction\Canvus\mt-canvus-server.ini`,
			`.\mt-canvus-server.ini`,
			`mt-canvus-server.ini`,
		}
	}
	for _, p := range candidates {
		expanded := os.ExpandEnv(p)
		if _, err := os.Stat(expanded); err == nil {
			if abs, err := filepath.Abs(expanded); err == nil {
				return abs, nil
			}
		}
	}
	return "", fmt.Errorf("mt-canvus-server.ini not found in common locations")
}

// LoadDBConfigFromINI loads database configuration from a Canvus server INI file.
func LoadDBConfigFromINI(iniPath string) (*DBConfig, error) {
	logger := logging.GetLogger()
	if iniPath == "" {
		iniPath = getDefaultINIPath()
	}
	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("INI file not found: %s", iniPath)
	}
	logger.Verbose("Loading database configuration from: %s", iniPath)

	cfg, err := ini.Load(iniPath)
	if err != nil {
		return nil, fmt.Errorf("load INI file: %w", err)
	}

	dbConfig := &DBConfig{SSLMode: "disable"}

	// Try section names in priority order.
	var section *ini.Section
	for _, name := range []string{"sql", "database", "postgresql", "db", "Database", "PostgreSQL", "DB", "SQL"} {
		if s := cfg.Section(name); s != nil {
			section = s
			logger.Verbose("Found database section: [%s]", name)
			break
		}
	}
	// Fallback: any section that has database-related keys.
	if section == nil {
		for _, s := range cfg.Sections() {
			if s.HasKey("host") || s.HasKey("database") || s.HasKey("dbname") {
				section = s
				logger.Verbose("Found database configuration in section: [%s]", s.Name())
				break
			}
		}
	}
	if section == nil {
		return nil, fmt.Errorf("no database configuration section found in: %s", iniPath)
	}

	keyOr := func(keys ...string) string {
		for _, k := range keys {
			if section.HasKey(k) {
				return section.Key(k).String()
			}
		}
		return ""
	}

	dbConfig.Host = keyOr("host", "hostname", "server")
	dbConfig.Port = keyOr("port")
	dbConfig.Database = keyOr("database", "databasename", "dbname", "name")
	dbConfig.Username = keyOr("username", "user")
	dbConfig.Password = keyOr("password", "pass")
	sslMode := keyOr("sslmode", "ssl_mode")
	if sslMode != "" {
		dbConfig.SSLMode = sslMode
	}

	if dbConfig.Host == "" {
		dbConfig.Host = "localhost"
		logger.Verbose("Database host not specified; defaulting to localhost")
	}
	if dbConfig.Port == "" {
		dbConfig.Port = "5432"
	}
	if dbConfig.Database == "" {
		return nil, fmt.Errorf("database name not found in INI file: %s", iniPath)
	}
	if dbConfig.Username == "" {
		return nil, fmt.Errorf("database username not found in INI file: %s", iniPath)
	}

	logger.Verbose("Database config: host=%s port=%s db=%s user=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.Database, dbConfig.Username)
	return dbConfig, nil
}

// GetConnectionString returns a PostgreSQL DSN string.
func (c *DBConfig) GetConnectionString() string {
	parts := []string{
		fmt.Sprintf("host=%s", c.Host),
		fmt.Sprintf("port=%s", c.Port),
		fmt.Sprintf("dbname=%s", c.Database),
		fmt.Sprintf("user=%s", c.Username),
		fmt.Sprintf("sslmode=%s", c.SSLMode),
	}
	if c.Password != "" {
		parts = append(parts, fmt.Sprintf("password=%s", c.Password))
	}
	return strings.Join(parts, " ")
}
