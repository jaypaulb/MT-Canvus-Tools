package configeditor

// ConfigSchema represents the complete configuration schema.
type ConfigSchema struct {
	Sections map[string]*ConfigSection // Map of section name to section
	Options  []*ConfigOption           // Flat list of all options
}

// ConfigSection represents a configuration section.
type ConfigSection struct {
	Name        string                  // Section name
	Description string                  // Section description
	Options     []*ConfigOption         // Options in this section
	IsCompound  bool                    // Whether this section supports compound entries
	Pattern     string                  // Pattern for compound entries (e.g., "server", "remote-desktop")
	CompoundEntries map[string][]*ConfigOption // Map of compound entry name to its options
}

// NewConfigSchema creates a new empty configuration schema.
func NewConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Sections: make(map[string]*ConfigSection),
		Options:  []*ConfigOption{},
	}
}

// AddOption adds an option to the schema.
func (cs *ConfigSchema) AddOption(option *ConfigOption) {
	cs.Options = append(cs.Options, option)

	// Add to section
	section, exists := cs.Sections[option.Section]
	if !exists {
		section = &ConfigSection{
			Name:            option.Section,
			Options:         []*ConfigOption{},
			CompoundEntries: make(map[string][]*ConfigOption),
		}
		cs.Sections[option.Section] = section
	}

	section.Options = append(section.Options, option)
}

// GetSection returns a section by name.
func (cs *ConfigSchema) GetSection(name string) *ConfigSection {
	return cs.Sections[name]
}

