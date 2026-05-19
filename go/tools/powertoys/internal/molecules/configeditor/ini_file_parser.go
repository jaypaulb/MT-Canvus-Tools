package configeditor

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// INIFileParser parses actual mt-canvus.ini files to extract configuration schema.
// It reads both active options and commented-out options (which represent defaults).
type INIFileParser struct {
	filePath string
}

// NewINIFileParser creates a new INI file parser.
func NewINIFileParser(filePath string) *INIFileParser {
	return &INIFileParser{
		filePath: filePath,
	}
}

// Parse parses the INI file and returns a ConfigSchema.
func (p *INIFileParser) Parse() (*ConfigSchema, error) {
	file, err := os.Open(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open INI file: %w", err)
	}
	defer file.Close()

	schema := NewConfigSchema()
	scanner := bufio.NewScanner(file)

	var currentSection string
	var currentDescription strings.Builder
	var pendingOption *PendingOption
	var inEnumList bool // Track if we're parsing enum values

	// Regex patterns
	sectionPattern := regexp.MustCompile(`^\[([^\]]+)\]`)
	commentedKeyPattern := regexp.MustCompile(`^;\s*([^=]+)=(.+)$`)
	activeKeyPattern := regexp.MustCompile(`^([^=;]+)=(.+)$`)
	defaultPattern := regexp.MustCompile(`DEFAULT[^:]*:\s*(.+)$`)
	permittedPattern := regexp.MustCompile(`Permitted values?:\s*(.+)$`)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			// If we have a pending option, save it
			if pendingOption != nil {
				p.savePendingOption(schema, currentSection, pendingOption)
				pendingOption = nil
			}
			currentDescription.Reset()
			continue
		}

		// Detect section headers
		if match := sectionPattern.FindStringSubmatch(trimmed); match != nil {
			// Save any pending option before changing sections
			if pendingOption != nil {
				p.savePendingOption(schema, currentSection, pendingOption)
				pendingOption = nil
			}

			sectionName := match[1]
			// Handle compound sections like [server:name] or [fixed-workspace:1]
			if strings.Contains(sectionName, ":") {
				parts := strings.Split(sectionName, ":")
				currentSection = parts[0]
				// Mark section as compound
				if section := schema.GetSection(currentSection); section != nil {
					section.IsCompound = true
					section.Pattern = currentSection
				}
			} else {
				currentSection = sectionName
			}
			currentDescription.Reset()
			continue
		}

		// Check for DEFAULT line
		if match := defaultPattern.FindStringSubmatch(trimmed); match != nil {
			defaultValue := strings.TrimSpace(match[1])
			// Handle multi-line defaults like "DEFAULT (port 443): ssl"
			if strings.Contains(defaultValue, "(") {
				// Extract just the value part
				parts := strings.Split(defaultValue, ":")
				if len(parts) > 1 {
					defaultValue = strings.TrimSpace(parts[len(parts)-1])
				} else {
					// Extract value before parenthesis
					parenIdx := strings.Index(defaultValue, "(")
					if parenIdx > 0 {
						defaultValue = strings.TrimSpace(defaultValue[:parenIdx])
					}
				}
			}
			if pendingOption != nil {
				pendingOption.Default = defaultValue
			}
			continue
		}

		// Check for permitted values / enum
		if match := permittedPattern.FindStringSubmatch(trimmed); match != nil {
			values := strings.Split(match[1], ",")
			for _, v := range values {
				v = strings.TrimSpace(v)
				// Remove markdown formatting like "* none"
				v = strings.TrimPrefix(v, "*")
				v = strings.TrimSpace(v)
				if v != "" {
					if pendingOption != nil {
						pendingOption.EnumValues = append(pendingOption.EnumValues, v)
					}
				}
			}
			continue
		}

		// Check for enum values in description - look for "Available options:" or "Available choices:"
		if strings.Contains(trimmed, "Available options:") || strings.Contains(trimmed, "Available choices:") {
			// Next lines will have enum values - parse them
			inEnumList = true
			// Don't add this line to description
			continue
		}

		// Parse enum values from lines like ";  true    Description" or ";   en  (English)"
		if inEnumList && strings.HasPrefix(trimmed, ";") {
			// Pattern: ";  value    Description" or ";   value  (Description)"
			desc := strings.TrimPrefix(trimmed, ";")
			desc = strings.TrimSpace(desc)
			// Try to extract value (first word before spaces/tabs)
			parts := strings.Fields(desc)
			if len(parts) > 0 {
				potentialValue := parts[0]
				// Check if it looks like a value (short, no colons, not a full sentence)
				if len(potentialValue) <= 20 && !strings.Contains(potentialValue, ":") &&
				   !strings.Contains(potentialValue, ".") && potentialValue != "" {
					// This is likely an enum value
					if pendingOption != nil {
						pendingOption.EnumValues = append(pendingOption.EnumValues, potentialValue)
					}
				} else {
					// Not an enum value, stop parsing enum list
					inEnumList = false
				}
			} else {
				// Empty line, stop parsing enum list
				inEnumList = false
			}
			// Continue to next line (don't add to description)
			continue
		} else if inEnumList {
			// Not a comment line, stop parsing enum list
			inEnumList = false
		}

		// Check for commented-out option (default value)
		if match := commentedKeyPattern.FindStringSubmatch(trimmed); match != nil {
			// Save previous pending option
			if pendingOption != nil {
				p.savePendingOption(schema, currentSection, pendingOption)
			}

			key := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])

			// Create new pending option
			pendingOption = &PendingOption{
				Key:         key,
				Value:       value,
				Description: currentDescription.String(),
				EnumValues:  []string{},
			}
			currentDescription.Reset()
			inEnumList = false
			continue
		}

		// Check for active (uncommented) option
		if match := activeKeyPattern.FindStringSubmatch(trimmed); match != nil {
			// Save previous pending option
			if pendingOption != nil {
				p.savePendingOption(schema, currentSection, pendingOption)
				pendingOption = nil
			}

			key := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])

			// Create option from active line
			option := p.createOption(currentSection, key, value, currentDescription.String(), []string{})
			schema.AddOption(option)
			currentDescription.Reset()
			continue
		}

		// Accumulate description from comments
		if strings.HasPrefix(trimmed, ";") {
			desc := strings.TrimPrefix(trimmed, ";")
			desc = strings.TrimSpace(desc)
			// Skip lines that are just section markers or empty
			if desc != "" && !strings.HasPrefix(desc, "[") {
				if currentDescription.Len() > 0 {
					currentDescription.WriteString(" ")
				}
				currentDescription.WriteString(desc)
			}
		}
	}

	// Save any remaining pending option
	if pendingOption != nil {
		p.savePendingOption(schema, currentSection, pendingOption)
	}

	return schema, scanner.Err()
}

// PendingOption represents an option that was found commented out (default value).
type PendingOption struct {
	Key         string
	Value       string
	Default     string
	Description string
	EnumValues  []string
}

// savePendingOption saves a pending option to the schema.
func (p *INIFileParser) savePendingOption(schema *ConfigSchema, section string, pending *PendingOption) {
	// Use the value as default if no explicit DEFAULT was found
	defaultValue := pending.Default
	if defaultValue == "" {
		defaultValue = pending.Value
	}

	option := p.createOption(section, pending.Key, "", pending.Description, pending.EnumValues)
	option.Default = defaultValue
	schema.AddOption(option)
}

// createOption creates a ConfigOption from parsed data.
func (p *INIFileParser) createOption(section, key, value, description string, enumValues []string) *ConfigOption {
	option := &ConfigOption{
		Section:     section,
		Key:         key,
		Description: description,
		EnumValues:  enumValues,
	}

	// Determine type
	if len(enumValues) > 0 {
		option.Type = ValueTypeEnum
	} else if value == "true" || value == "false" || strings.Contains(strings.ToLower(description), "true") && strings.Contains(strings.ToLower(description), "false") {
		// Check if it's a boolean option
		if strings.Contains(strings.ToLower(key), "enabled") ||
			strings.Contains(strings.ToLower(key), "allow") ||
			strings.Contains(strings.ToLower(description), "enable") ||
			strings.Contains(strings.ToLower(description), "disable") {
			option.Type = ValueTypeBoolean
		} else if value == "auto" || strings.Contains(description, "auto") {
			option.Type = ValueTypeEnum
			option.EnumValues = []string{"true", "false", "auto"}
		} else {
			option.Type = ValueTypeBoolean
		}
	} else if strings.Contains(strings.ToLower(key), "path") ||
		strings.Contains(strings.ToLower(key), "folder") ||
		strings.Contains(strings.ToLower(key), "file") ||
		strings.Contains(strings.ToLower(key), "root") {
		option.Type = ValueTypeFilePath
	} else if strings.Contains(key, "-size") || strings.Contains(key, "-timeout") ||
		strings.Contains(key, "-port") || strings.Contains(key, "-age") ||
		strings.Contains(key, "-fps") || strings.Contains(key, "-resolution") ||
		strings.Contains(key, "-scale") || strings.Contains(key, "-length") {
		option.Type = ValueTypeNumber
	} else if strings.Contains(key, "-layouts") || strings.Contains(key, "-sites") ||
		strings.Contains(key, "-extensions") || strings.Contains(key, "-codes") ||
		strings.Contains(key, "-volumes") || strings.Contains(key, "-gestures") ||
		strings.Contains(key, "-options") {
		option.Type = ValueTypeCommaList
	} else {
		option.Type = ValueTypeString
	}

	return option
}

