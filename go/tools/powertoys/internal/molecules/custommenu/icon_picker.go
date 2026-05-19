package custommenu

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/logger"
)

// IconPicker provides an icon picker dialog.
type IconPicker struct {
	iconPaths []string
	icons     map[string]*canvas.Image
	gridSize  int
}

// NewIconPicker creates a new icon picker.
func NewIconPicker() (*IconPicker, error) {
	picker := &IconPicker{
		icons:    make(map[string]*canvas.Image),
		gridSize: 64, // Thumbnail size
	}

	// Load icons from dev-docs/canvus-custom-menu/icons/
	if err := picker.loadIcons(); err != nil {
		logger.Logf("Warning: Failed to load icons: %v", err)
		return picker, err
	}

	return picker, nil
}

// loadIcons loads all icons from the dev-docs directory.
func (p *IconPicker) loadIcons() error {
	// Try to find the icons directory
	iconDirs := []string{
		"/home/jaypaulb/Documents/gh/CanvusPowerToys/dev-docs/canvus-custom-menu/icons",
		"dev-docs/canvus-custom-menu/icons",
		"../dev-docs/canvus-custom-menu/icons",
	}

	var iconDir string
	for _, dir := range iconDirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			iconDir = dir
			break
		}
	}

	if iconDir == "" {
		return os.ErrNotExist
	}

	// Walk the icon directory
	err := filepath.Walk(iconDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			return nil
		}

		// Check if it's a PNG file
		if strings.ToLower(filepath.Ext(path)) == ".png" {
			p.iconPaths = append(p.iconPaths, path)
		}

		return nil
	})

	// Sort paths naturally (1.png, 2.png, 10.png, 16.png instead of 1.png, 10.png, 16.png, 2.png)
	sort.Slice(p.iconPaths, func(i, j int) bool {
		return naturalLess(p.iconPaths[i], p.iconPaths[j])
	})

	logger.Logf("Loaded %d icons from %s", len(p.iconPaths), iconDir)
	return err
}

// naturalLess compares two strings using natural sort order (numeric parts sorted numerically).
func naturalLess(a, b string) bool {
	// Extract numeric parts
	re := regexp.MustCompile(`\d+`)

	aBase := filepath.Base(a)
	bBase := filepath.Base(b)

	aMatches := re.FindAllStringIndex(aBase, -1)
	bMatches := re.FindAllStringIndex(bBase, -1)

	aPos, bPos := 0, 0

	for len(aMatches) > 0 && len(bMatches) > 0 {
		aMatch := aMatches[0]
		bMatch := bMatches[0]

		// Compare text before numbers
		aText := aBase[aPos:aMatch[0]]
		bText := bBase[bPos:bMatch[0]]

		if aText != bText {
			return aText < bText
		}

		// Compare numbers numerically
		aNum, _ := strconv.Atoi(aBase[aMatch[0]:aMatch[1]])
		bNum, _ := strconv.Atoi(bBase[bMatch[0]:bMatch[1]])

		if aNum != bNum {
			return aNum < bNum
		}

		// Move to next match
		aPos = aMatch[1]
		bPos = bMatch[1]
		aMatches = aMatches[1:]
		bMatches = bMatches[1:]
	}

	// If all numbers matched, compare remaining text
	return aBase[aPos:] < bBase[bPos:]
}

// Show displays the icon picker dialog.
func (p *IconPicker) Show(parent fyne.Window, onSelect func(string)) {
	if len(p.iconPaths) == 0 {
		dialog.ShowError(os.ErrNotExist, parent)
		return
	}

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search icons...")

	// Create grid of icons
	grid := container.NewGridWrap(fyne.NewSize(80, 80))

	var filteredPaths []string

	updateGrid := func(filter string) {
		grid.Objects = nil
		filteredPaths = nil

		for _, iconPath := range p.iconPaths {
			filename := filepath.Base(iconPath)
			if filter == "" || strings.Contains(strings.ToLower(filename), strings.ToLower(filter)) {
				filteredPaths = append(filteredPaths, iconPath)

				// Create placeholder thumbnail (will load async)
				placeholder := canvas.NewRectangle(nil)
				placeholder.SetMinSize(fyne.NewSize(float32(p.gridSize), float32(p.gridSize)))

				btn := widget.NewButton("Select", func(path string) func() {
					return func() {
						onSelect(path)
					}
				}(iconPath))

				card := container.NewVBox(
					placeholder,
					widget.NewLabel(filepath.Base(iconPath)),
					btn,
				)

				grid.Add(card)

				// Load thumbnail asynchronously - capture the card directly
				go func(path string, cardContainer *fyne.Container) {
					thumbnail := p.createThumbnail(path)
					if thumbnail != nil && len(cardContainer.Objects) > 0 {
						// Update the first object (placeholder) with the loaded thumbnail
						cardContainer.Objects[0] = thumbnail
						cardContainer.Refresh()
					}
				}(iconPath, card)
			}
		}

		grid.Refresh()
	}

	searchEntry.OnChanged = updateGrid

	// Initial load
	updateGrid("")

	// Add New button
	addNewBtn := widget.NewButton("Add Custom Icon", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			selectedPath := reader.URI().Path()

			// Check if conversion is needed
			if filepath.Ext(selectedPath) == ".png" {
				dialog.ShowConfirm("Convert Icon", "Convert this icon to 937x937 PNG format?", func(convert bool) {
					if convert {
						converter := NewIconConverter()
						outputDir := filepath.Join(filepath.Dir(selectedPath), "converted_icons")
						os.MkdirAll(outputDir, 0755)
						outputPath := filepath.Join(outputDir, filepath.Base(selectedPath))

						if err := converter.ConvertIcon(selectedPath, outputPath); err != nil {
							dialog.ShowError(err, parent)
						} else {
							onSelect(outputPath)
							dialog.ShowInformation("Success", "Icon converted and selected", parent)
						}
					} else {
						onSelect(selectedPath)
					}
				}, parent)
			} else {
				onSelect(selectedPath)
			}
		}, parent)
	})

	scrollable := container.NewScroll(grid)
	scrollable.SetMinSize(fyne.NewSize(600, 400))

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Select an icon (937x937 PNG recommended):"),
			searchEntry,
			addNewBtn,
		),
		nil, nil, nil,
		scrollable,
	)

	dialog.ShowCustom("Icon Picker", "Close", content, parent)
}

// createThumbnail creates a thumbnail image for the icon.
func (p *IconPicker) createThumbnail(iconPath string) *canvas.Image {
	// Check cache
	if img, ok := p.icons[iconPath]; ok {
		// Return a copy to avoid shared state issues
		newImg := canvas.NewImageFromImage(img.Image)
		newImg.FillMode = canvas.ImageFillOriginal
		newImg.SetMinSize(fyne.NewSize(float32(p.gridSize), float32(p.gridSize)))
		return newImg
	}

	// Load image
	file, err := os.Open(iconPath)
	if err != nil {
		logger.Logf("Failed to open icon %s: %v", iconPath, err)
		return nil
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		logger.Logf("Failed to decode icon %s: %v", iconPath, err)
		return nil
	}

	// Create Fyne image with zoom to show center content better
	// Icons are 937x937 with content in center ~250x250
	// We use ImageFillOriginal with larger min size to "zoom in"
	thumbnail := canvas.NewImageFromImage(img)
	thumbnail.FillMode = canvas.ImageFillOriginal // Show actual pixels, zoomed
	thumbnail.SetMinSize(fyne.NewSize(float32(p.gridSize)*3, float32(p.gridSize)*3)) // 3x zoom

	// Cache the decoded image (not the widget)
	p.icons[iconPath] = thumbnail

	return thumbnail
}
