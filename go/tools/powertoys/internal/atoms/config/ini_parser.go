package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

// INIParser handles INI file parsing and writing.
type INIParser struct{}

// NewINIParser creates a new INI parser.
func NewINIParser() *INIParser {
	return &INIParser{}
}

// Read reads an INI file and returns the configuration.
// Commented lines (starting with ;) are properly ignored.
func (p *INIParser) Read(filePath string) (*ini.File, error) {
	// Use LoadOptions to ensure comments are properly handled
	// The ini library should handle comments by default, but we'll be explicit
	cfg, err := ini.LoadSources(ini.LoadOptions{
		AllowPythonMultilineValues: false,
		SpaceBeforeInlineComment:   true,
		UnescapeValueDoubleQuotes:  true,
		UnescapeValueCommentSymbols: true,
		IgnoreInlineComment:        false,
	}, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load INI file: %w", err)
	}

	// Additional safety: Remove any keys that might have been parsed from commented lines
	// This is a defensive measure in case the parser still picks up commented keys
	p.removeCommentedKeys(cfg, filePath)

	return cfg, nil
}

// removeCommentedKeys removes keys and sections that were parsed from commented lines.
// This reads the original file to identify which keys and sections are actually commented out.
func (p *INIParser) removeCommentedKeys(cfg *ini.File, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		// If we can't read the file, skip cleanup
		return
	}
	defer file.Close()

	// Read the file to identify commented keys and sections
	commentedKeys := make(map[string]map[string]bool) // section -> key -> true
	commentedSections := make(map[string]bool)        // section name -> true
	currentSection := ini.DEFAULT_SECTION             // Default section for root-level keys

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Check for section headers (including commented ones)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			sectionName := strings.Trim(trimmed, "[]")
			// Use empty string for default section, otherwise use the section name
			if sectionName == "" || sectionName == ini.DEFAULT_SECTION {
				currentSection = ini.DEFAULT_SECTION
			} else {
				currentSection = sectionName
			}
			continue
		}

		// Check if this is a commented section header (starts with ; and has [])
		if strings.HasPrefix(trimmed, ";") {
			uncommented := strings.TrimSpace(strings.TrimPrefix(trimmed, ";"))
			if strings.HasPrefix(uncommented, "[") && strings.HasSuffix(uncommented, "]") {
				// This is a commented section header
				sectionName := strings.Trim(uncommented, "[]")
				if sectionName != "" && sectionName != ini.DEFAULT_SECTION {
					commentedSections[sectionName] = true
				}
				continue
			}
		}

		// Check if this is a commented key (starts with ; and has =)
		if strings.HasPrefix(trimmed, ";") {
			// Remove the semicolon and check if it's a key=value line
			uncommented := strings.TrimSpace(strings.TrimPrefix(trimmed, ";"))
			if strings.Contains(uncommented, "=") {
				parts := strings.SplitN(uncommented, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					if key != "" {
						if commentedKeys[currentSection] == nil {
							commentedKeys[currentSection] = make(map[string]bool)
						}
						commentedKeys[currentSection][key] = true
					}
				}
			}
		}
	}

	// Remove commented sections from the config
	for sectionName := range commentedSections {
		cfg.DeleteSection(sectionName)
	}

	// Remove commented keys from the config
	for sectionName, keys := range commentedKeys {
		section := cfg.Section(sectionName)
		if section != nil {
			for keyName := range keys {
				section.DeleteKey(keyName)
			}
		}
	}
}

// Write writes an INI configuration to a file.
func (p *INIParser) Write(cfg *ini.File, filePath string) error {
	if err := cfg.SaveTo(filePath); err != nil {
		return fmt.Errorf("failed to save INI file: %w", err)
	}
	return nil
}

// GetSection gets a section from the INI file.
func (p *INIParser) GetSection(cfg *ini.File, sectionName string) (*ini.Section, error) {
	section := cfg.Section(sectionName)
	if section == nil {
		return nil, fmt.Errorf("section '%s' not found", sectionName)
	}
	return section, nil
}

// GetValue gets a value from a section.
func (p *INIParser) GetValue(section *ini.Section, key string) string {
	return section.Key(key).String()
}

// SetValue sets a value in a section.
func (p *INIParser) SetValue(section *ini.Section, key, value string) {
	section.Key(key).SetValue(value)
}

// CreateFile creates a new INI file.
func (p *INIParser) CreateFile(filePath string) (*ini.File, error) {
	cfg := ini.Empty()

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return cfg, nil
}
