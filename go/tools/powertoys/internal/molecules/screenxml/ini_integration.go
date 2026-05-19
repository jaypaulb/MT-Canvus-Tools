package screenxml

import (
	"fmt"
	"strings"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/config"
)

// INIIntegration handles integration with mt-canvus.ini.
type INIIntegration struct {
	iniParser *config.INIParser
}

// NewINIIntegration creates a new INI integration handler.
func NewINIIntegration() *INIIntegration {
	return &INIIntegration{
		iniParser: config.NewINIParser(),
	}
}

// OutputCell represents a cell that needs an output section in the INI file.
type OutputCell struct {
	Row        int
	Col        int
	GPUOutput  string
	Resolution Resolution
}

// DetectVideoOutputs detects cells without layout that need output sections.
func (ii *INIIntegration) DetectVideoOutputs(grid *GridWidget) []OutputCell {
	var outputCells []OutputCell

	rows := grid.GetRows()
	cols := grid.GetCols()
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := grid.GetCell(row, col)
			// Cell needs output section if it has GPU output but no layout (Index is empty)
			if cell != nil && cell.GPUOutput != "" && cell.Index == "" {
				outputCells = append(outputCells, OutputCell{
					Row:        row,
					Col:        col,
					GPUOutput:  cell.GPUOutput,
					Resolution: cell.Resolution,
				})
			}
		}
	}

	return outputCells
}

// generateOutputName generates an auto-generated output name from GPU output.
func (ii *INIIntegration) generateOutputName(gpuOutput string, index int) string {
	// Convert from UI format (1:1, 1:2, etc.) to output name (gpu1output1, gpu1output2, etc.)
	parts := strings.Split(gpuOutput, ":")
	if len(parts) == 2 {
		return fmt.Sprintf("gpu%so%s", parts[0], parts[1])
	}
	// Fallback to indexed name
	return fmt.Sprintf("output%d", index+1)
}

// UpdateMtCanvusIni updates mt-canvus.ini with output sections for cells without layout.
func (ii *INIIntegration) UpdateMtCanvusIni(filePath string, outputCells []OutputCell) error {
	// Read existing INI file
	cfg, err := ii.iniParser.Read(filePath)
	if err != nil {
		return fmt.Errorf("failed to read mt-canvus.ini: %w", err)
	}

	// Remove existing [output:*] sections to avoid duplicates
	sectionsToDelete := []string{}
	for _, section := range cfg.Sections() {
		if strings.HasPrefix(section.Name(), "output:") {
			sectionsToDelete = append(sectionsToDelete, section.Name())
		}
	}
	for _, sectionName := range sectionsToDelete {
		cfg.DeleteSection(sectionName)
	}

	// Create [output:...] sections for each cell without layout
	for i, cell := range outputCells {
		// Generate output name
		outputName := ii.generateOutputName(cell.GPUOutput, i)
		sectionName := fmt.Sprintf("output:%s", outputName)

		// Create or get section (delete if exists to recreate fresh)
		existingSection := cfg.Section(sectionName)
		if existingSection != nil {
			cfg.DeleteSection(sectionName)
		}
		section, err := cfg.NewSection(sectionName)
		if err != nil {
			return fmt.Errorf("failed to create section %s: %w", sectionName, err)
		}

		// Calculate location in pixels from top-left (0,0)
		// Location = (col * resolution.width, row * resolution.height)
		locationX := cell.Col * cell.Resolution.Width
		locationY := cell.Row * cell.Resolution.Height
		location := fmt.Sprintf("%d %d", locationX, locationY)

		// Size = resolution (width x height)
		size := fmt.Sprintf("%d %d", cell.Resolution.Width, cell.Resolution.Height)

		// Set location and size
		section.Key("location").SetValue(location)
		section.Key("size").SetValue(size)
	}

	// Write back to file
	if err := ii.iniParser.Write(cfg, filePath); err != nil {
		return fmt.Errorf("failed to write mt-canvus.ini: %w", err)
	}

	return nil
}

// ShouldUpdateIni checks if output sections should be updated in mt-canvus.ini.
func (ii *INIIntegration) ShouldUpdateIni(grid *GridWidget) bool {
	outputCells := ii.DetectVideoOutputs(grid)
	return len(outputCells) > 0
}


