package canvus

import (
	"log/slog"
	"os"
	"sync"
)

// APIWarning represents a known API limitation or issue surfaced by the SDK.
type APIWarning struct {
	Code        string
	Description string
	Workaround  string
	IssueURL    string
}

// Known API limitations and issues.
var (
	// WarningNoteTitleNotExposed indicates Note widget titles are not API-visible.
	WarningNoteTitleNotExposed = APIWarning{
		Code:        "NOTE_TITLE_NOT_EXPOSED",
		Description: "Note widget 'title' field is not exposed by the Canvus API. Title values in requests are ignored and responses will not include the title.",
		Workaround:  "Use the 'name' field instead for identifying notes.",
		IssueURL:    "https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/38",
	}
	// WarningVideoInputTitleNotExposed indicates VideoInput titles are not API-visible.
	WarningVideoInputTitleNotExposed = APIWarning{
		Code:        "VIDEOINPUT_TITLE_NOT_EXPOSED",
		Description: "VideoInput widget 'title' field is not exposed by the Canvus API.",
		Workaround:  "No workaround available. Await Canvus API fix.",
		IssueURL:    "https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/13",
	}
	// WarningPDFSizeBug indicates PDF resize creates a visual disconnect.
	WarningPDFSizeBug = APIWarning{
		Code:        "PDF_SIZE_BUG",
		Description: "PDF widget size changes via PATCH update the bounding box but the actual PDF content stays at its original size.",
		Workaround:  "Avoid resizing PDFs via API. Delete and recreate if different size is needed.",
		IssueURL:    "https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/15",
	}
	// WarningImageAspectRatioNotPreserved indicates Image resize ignores aspect ratio.
	WarningImageAspectRatioNotPreserved = APIWarning{
		Code:        "IMAGE_ASPECT_RATIO_NOT_PRESERVED",
		Description: "Image widget size changes via PATCH do not preserve aspect ratio.",
		Workaround:  "Calculate correct aspect-ratio-preserving dimensions before calling UpdateImage.",
		IssueURL:    "https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/39",
	}
	// WarningVideoAspectRatioNotPreserved indicates Video resize ignores aspect ratio.
	WarningVideoAspectRatioNotPreserved = APIWarning{
		Code:        "VIDEO_ASPECT_RATIO_NOT_PRESERVED",
		Description: "Video widget size changes via PATCH do not preserve aspect ratio.",
		Workaround:  "Calculate correct aspect-ratio-preserving dimensions before calling UpdateVideo.",
		IssueURL:    "https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/39",
	}
	// WarningTableGridSizeImmutable indicates grid_size is silently ignored on PATCH.
	// Documented per changelog §5.
	WarningTableGridSizeImmutable = APIWarning{
		Code:        "TABLE_GRID_SIZE_IMMUTABLE",
		Description: "Table widget 'grid_size' is set at creation time and silently ignored on PATCH.",
		Workaround:  "Omit grid_size from PATCH requests; recreate the table to change dimensions.",
		IssueURL:    "https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/-",
	}
)

var (
	warningsOnce    sync.Once // retained for backwards compat (no users)
	warningsIssued  = make(map[string]bool)
	warningsMu      sync.Mutex
	warningsEnabled = true
)

func init() {
	if os.Getenv("CANVUS_SDK_DISABLE_WARNINGS") == "1" {
		warningsEnabled = false
	}
}

// DisableAPIWarnings silences all SDK warnings for the remainder of the process.
func DisableAPIWarnings() {
	warningsMu.Lock()
	defer warningsMu.Unlock()
	warningsEnabled = false
}

// EnableAPIWarnings re-enables SDK warnings if they were disabled.
func EnableAPIWarnings() {
	warningsMu.Lock()
	defer warningsMu.Unlock()
	warningsEnabled = true
}

// warnOnce emits a warning the first time it's seen during the lifetime of the
// process. Drift remediation: routed through slog at Warn level so embedders'
// configured handler decides destination.
func warnOnce(warning APIWarning) {
	warningsMu.Lock()
	defer warningsMu.Unlock()

	if !warningsEnabled {
		return
	}
	if warningsIssued[warning.Code] {
		return
	}
	warningsIssued[warning.Code] = true

	slog.Warn("canvus SDK warning",
		"code", warning.Code,
		"description", warning.Description,
		"workaround", warning.Workaround,
		"issue", warning.IssueURL,
	)
}

// warnAlways emits a warning on every call (rare; used where reminder matters).
func warnAlways(warning APIWarning) {
	warningsMu.Lock()
	defer warningsMu.Unlock()
	if !warningsEnabled {
		return
	}
	slog.Warn("canvus SDK warning",
		"code", warning.Code,
		"description", warning.Description,
	)
}

// ResetWarnings clears the record of issued warnings (primarily for tests).
func ResetWarnings() {
	warningsMu.Lock()
	defer warningsMu.Unlock()
	warningsIssued = make(map[string]bool)
}

// SuppressWarningsOnce is reserved for future single-shot suppression patterns.
func SuppressWarningsOnce() { warningsOnce.Do(func() {}) }
