package screenxml

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Resolution represents a display resolution.
type Resolution struct {
	Width  int
	Height int
	Name   string
}

// EffectiveWidth returns the width considering orientation (swaps for portrait).
func (r Resolution) EffectiveWidth(orientation Orientation) int {
	if orientation == OrientationPortrait {
		return r.Height
	}
	return r.Width
}

// EffectiveHeight returns the height considering orientation (swaps for portrait).
func (r Resolution) EffectiveHeight(orientation Orientation) int {
	if orientation == OrientationPortrait {
		return r.Width
	}
	return r.Height
}

// Common resolutions
var CommonResolutions = []Resolution{
	{1920, 1080, "1920x1080 (Full HD)"},
	{2560, 1440, "2560x1440 (QHD)"},
	{3840, 2160, "3840x2160 (4K UHD)"},
	{1280, 720, "1280x720 (HD)"},
	{1600, 900, "1600x900"},
	{2560, 1080, "2560x1080 (Ultrawide)"},
	{3440, 1440, "3440x1440 (Ultrawide QHD)"},
}

// ResolutionHandler manages resolution selection and detection.
type ResolutionHandler struct {
	resolutionSelect *widget.Select
	currentRes       Resolution
	onResolutionChange func(Resolution)
}

// NewResolutionHandler creates a new resolution handler.
func NewResolutionHandler() *ResolutionHandler {
	rh := &ResolutionHandler{
		currentRes: CommonResolutions[0], // Default to 1920x1080
	}

	// Create resolution dropdown
	options := make([]string, len(CommonResolutions))
	for i, res := range CommonResolutions {
		options[i] = res.Name
	}

	rh.resolutionSelect = widget.NewSelect(options, func(selected string) {
		for _, res := range CommonResolutions {
			if res.Name == selected {
				rh.currentRes = res
				if rh.onResolutionChange != nil {
					rh.onResolutionChange(res)
				}
				break
			}
		}
	})
	rh.resolutionSelect.SetSelected(CommonResolutions[0].Name)

	return rh
}

// CreateUI creates the UI for resolution selection.
func (rh *ResolutionHandler) CreateUI() fyne.CanvasObject {
	label := widget.NewLabel("Resolution per Output:")
	note := widget.NewLabel("(Default: 1920x1080 - can be overridden per output)")

	form := container.NewVBox(
		label,
		rh.resolutionSelect,
		note,
	)

	return form
}

// GetCurrentResolution returns the currently selected resolution.
func (rh *ResolutionHandler) GetCurrentResolution() Resolution {
	return rh.currentRes
}

// SetOnResolutionChange sets the callback for resolution changes.
func (rh *ResolutionHandler) SetOnResolutionChange(fn func(Resolution)) {
	rh.onResolutionChange = fn
}

// DetectResolution attempts to detect resolution for a GPU output.
// This is a placeholder - actual implementation would query Windows WMI/DXGI or Linux xrandr.
func (rh *ResolutionHandler) DetectResolution(gpuOutput string) (Resolution, error) {
	// TODO: Implement actual resolution detection
	// For now, return default resolution
	return CommonResolutions[0], nil
}

