package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/logging"
	_ "github.com/lib/pq" // Postgres driver — imported for side effects.
)

// AssetFileRecord represents a row from the asset_files table.
type AssetFileRecord struct {
	PublicHash       string
	PrivateHash      string
	OriginalFilename string
}

// Client handles PostgreSQL database operations.
type Client struct {
	db     *sql.DB
	logger *logging.Logger
}

// NewClient opens a connection to the PostgreSQL database described by config.
func NewClient(config *DBConfig) (*Client, error) {
	logger := logging.GetLogger()
	connStr := config.GetConnectionString()
	logger.Verbose("Connecting to PostgreSQL: %s@%s:%s/%s",
		config.Username, config.Host, config.Port, config.Database)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open database connection: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	logger.Info("Successfully connected to PostgreSQL database")
	return &Client{db: db, logger: logger}, nil
}

// Close closes the underlying database connection.
func (c *Client) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// FindAssetByOriginalFilename searches the asset_files table for records whose
// original filename matches exactly. It tries several common column-name
// variants used across Canvus server versions.
func (c *Client) FindAssetByOriginalFilename(originalFilename string) ([]AssetFileRecord, error) {
	if originalFilename == "" {
		return nil, fmt.Errorf("original filename cannot be empty")
	}
	queries := []string{
		`SELECT public_hash, private_hash, original_filename FROM asset_files WHERE original_filename = $1`,
		`SELECT public_hash, private_hash, filename FROM asset_files WHERE filename = $1`,
		`SELECT public_hash, private_hash, name FROM asset_files WHERE name = $1`,
	}
	return c.runQueries(queries, originalFilename)
}

// FindAssetByOriginalFilenameCaseInsensitive performs a case-insensitive search.
func (c *Client) FindAssetByOriginalFilenameCaseInsensitive(originalFilename string) ([]AssetFileRecord, error) {
	if originalFilename == "" {
		return nil, fmt.Errorf("original filename cannot be empty")
	}
	queries := []string{
		`SELECT public_hash, private_hash, original_filename FROM asset_files WHERE LOWER(original_filename) = LOWER($1)`,
		`SELECT public_hash, private_hash, filename FROM asset_files WHERE LOWER(filename) = LOWER($1)`,
		`SELECT public_hash, private_hash, name FROM asset_files WHERE LOWER(name) = LOWER($1)`,
	}
	return c.runQueries(queries, originalFilename)
}

// runQueries tries each query in turn, returning the first non-empty result set.
func (c *Client) runQueries(queries []string, arg string) ([]AssetFileRecord, error) {
	var lastErr error
	for _, q := range queries {
		c.logger.Verbose("Trying query: %s", q)
		rows, err := c.db.Query(q, arg)
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "column") {
				lastErr = err
				continue
			}
			return nil, fmt.Errorf("query failed: %w", err)
		}
		var records []AssetFileRecord
		for rows.Next() {
			var r AssetFileRecord
			var pub, priv, fname sql.NullString
			if scanErr := rows.Scan(&pub, &priv, &fname); scanErr != nil {
				_ = rows.Close()
				lastErr = scanErr
				break
			}
			if pub.Valid {
				r.PublicHash = pub.String
			}
			if priv.Valid {
				r.PrivateHash = priv.String
			}
			if fname.Valid {
				r.OriginalFilename = fname.String
			}
			records = append(records, r)
		}
		_ = rows.Close()
		if len(records) > 0 {
			c.logger.Verbose("Found %d records for: %s", len(records), arg)
			return records, nil
		}
	}
	if lastErr != nil {
		return nil, fmt.Errorf("no records found; last error: %w", lastErr)
	}
	return nil, fmt.Errorf("no records found for: %s", arg)
}
