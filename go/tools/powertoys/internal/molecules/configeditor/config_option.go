package configeditor

// ValueType represents the type of a configuration option value.
type ValueType string

const (
	ValueTypeBoolean ValueType = "boolean"
	ValueTypeEnum    ValueType = "enum"
	ValueTypeNumber  ValueType = "number"
	ValueTypeString  ValueType = "string"
	ValueTypeFilePath ValueType = "filepath"
	ValueTypeCommaList ValueType = "comma-list"
)

// ConfigOption represents a single configuration option.
type ConfigOption struct {
	Section     string   // Section name (e.g., "system", "canvas")
	Key         string   // Key name (e.g., "lock-config", "multi-user-mode-enabled")
	Type        ValueType // Type of value
	Default     string   // Default value
	Description string   // Description from documentation
	EnumValues  []string // Allowed values for enum type (e.g., ["true", "false", "auto"])
	Required    bool     // Whether this option is required
	IsCompound  bool     // Whether this is part of a compound entry (e.g., [server:name])
	Pattern     string   // Pattern for compound entries (e.g., "server", "remote-desktop", "fixed-workspace")
}

// GetCurrentValue gets the current value from the loaded INI file.
func (co *ConfigOption) GetCurrentValue(iniFile interface{}) string {
	// This will be implemented to read from INI file
	// For now, return default
	return co.Default
}

