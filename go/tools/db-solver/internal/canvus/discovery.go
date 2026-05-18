// Package canvus provides asset discovery via the Canvus Server API.
package canvus

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	canvussdk "github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/logging"
)

// AssetInfo describes a single discovered media asset.
type AssetInfo struct {
	Hash             string `json:"hash"`
	WidgetType       string `json:"widget_type"`
	OriginalFilename string `json:"original_filename"`
	CanvasID         string `json:"canvas_id"`
	CanvasName       string `json:"canvas_name"`
	WidgetID         string `json:"widget_id"`
	WidgetName       string `json:"widget_name"`
}

// DiscoveryResult holds the complete outcome of an asset-discovery run.
type DiscoveryResult struct {
	Assets            []AssetInfo              `json:"assets"`
	AssetsWithoutHash []AssetInfo              `json:"assets_without_hash"`
	Canvases          []canvussdk.Canvas       `json:"canvases"`
	StartTime         time.Time                `json:"start_time"`
	EndTime           time.Time                `json:"end_time"`
	Duration          time.Duration            `json:"duration"`
	Errors            []string                 `json:"errors"`
	ServerValidation  *ServerValidationResult  `json:"server_validation,omitempty"`
}

// ServerValidationResult is a lightweight summary of unique-asset counts.
type ServerValidationResult struct {
	TotalAssets      int      `json:"total_assets"`
	ExistingAssets   int      `json:"existing_assets"`
	MissingAssets    int      `json:"missing_assets"`
	ValidationErrors []string `json:"validation_errors"`
}

// RateLimiter controls the rate of API requests.
type RateLimiter struct {
	requests chan struct{}
	rate     time.Duration
}

// NewRateLimiter creates a token-bucket rate limiter allowing requestsPerSecond calls/s.
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	rate := time.Second / time.Duration(requestsPerSecond)
	rl := &RateLimiter{
		requests: make(chan struct{}, requestsPerSecond),
		rate:     rate,
	}
	go rl.run()
	return rl
}

func (rl *RateLimiter) run() {
	ticker := time.NewTicker(rl.rate)
	defer ticker.Stop()
	for range ticker.C {
		select {
		case rl.requests <- struct{}{}:
		default:
		}
	}
}

// Wait blocks until a request token is available.
func (rl *RateLimiter) Wait() { <-rl.requests }

// DiscoveryOptions customises a discovery run.
type DiscoveryOptions struct {
	SkipArchived bool
}

// DiscoverAllAssets discovers all media assets across every canvas.
func DiscoverAllAssets(session *canvussdk.Session, requestsPerSecond int) (*DiscoveryResult, error) {
	return DiscoverAllAssetsWithOptions(session, requestsPerSecond, DiscoveryOptions{})
}

// DiscoverAllAssetsWithOptions discovers media assets with caller-supplied options.
func DiscoverAllAssetsWithOptions(session *canvussdk.Session, requestsPerSecond int, options DiscoveryOptions) (*DiscoveryResult, error) {
	startTime := time.Now()
	result := &DiscoveryResult{
		StartTime:         startTime,
		Assets:            make([]AssetInfo, 0),
		AssetsWithoutHash: make([]AssetInfo, 0),
		Canvases:          make([]canvussdk.Canvas, 0),
		Errors:            make([]string, 0),
	}

	ctx := context.Background()
	logger := logging.GetLogger()

	// Step 1: Fetch all canvases.
	logger.Info("")
	logger.Info("STEP 1: Fetching canvas list from Canvus Server API...")
	allCanvases, err := session.ListCanvases(ctx, nil)
	if err != nil {
		logger.Error("STEP 1 FAILED: %v", err)
		return nil, fmt.Errorf("list canvases: %w", err)
	}
	logger.Info("STEP 1 COMPLETE: %d canvases returned", len(allCanvases))

	// Step 2: Filter archived canvases.
	logger.Info("STEP 2: Filtering archived canvases...")
	var canvases []canvussdk.Canvas
	archivedCount := 0
	for _, c := range allCanvases {
		if c.InTrash {
			archivedCount++
			continue
		}
		canvases = append(canvases, c)
	}
	logger.Info("STEP 2 COMPLETE: %d active, %d archived skipped", len(canvases), archivedCount)
	result.Canvases = canvases

	// Step 3: Extract media assets from each canvas in parallel.
	logger.Info("STEP 3: Processing %d canvases for media assets...", len(canvases))
	rateLimiter := NewRateLimiter(requestsPerSecond)
	var wg sync.WaitGroup
	var mu sync.Mutex
	semaphore := make(chan struct{}, 10)

	var processedCount, successCount, archivedByServerCount, otherErrorCount int
	total := len(canvases)

	for _, canvas := range canvases {
		wg.Add(1)
		go func(c canvussdk.Canvas) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			rateLimiter.Wait()

			widgetAssets, widgetNoHash, widgetErr := extractMediaAssetsWithError(ctx, session, c)
			bgAssets, bgNoHash := extractBackgroundAssets(ctx, session, c)

			mu.Lock()
			defer mu.Unlock()
			processedCount++
			if widgetErr != nil {
				errStr := widgetErr.Error()
				if strings.Contains(errStr, "archived") {
					archivedByServerCount++
				} else {
					otherErrorCount++
				}
			} else {
				successCount++
				result.Assets = append(result.Assets, widgetAssets...)
				result.Assets = append(result.Assets, bgAssets...)
				result.AssetsWithoutHash = append(result.AssetsWithoutHash, widgetNoHash...)
				result.AssetsWithoutHash = append(result.AssetsWithoutHash, bgNoHash...)
			}
			if processedCount%100 == 0 || processedCount == total {
				logger.Info("Progress: %d/%d canvases processed", processedCount, total)
			}
		}(canvas)
	}
	wg.Wait()

	logger.Info("STEP 3 COMPLETE: %d ok / %d server-archived / %d errors | %d assets / %d without hash",
		successCount, archivedByServerCount, otherErrorCount,
		len(result.Assets), len(result.AssetsWithoutHash))

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Step 4: Count unique assets (server-side download validation is disabled to prevent OOM).
	logger.Info("STEP 4: Counting unique assets...")
	validationResult, err := countUniqueAssets(result.Assets)
	if err != nil {
		logger.Warn("Asset validation failed: %v", err)
		result.Errors = append(result.Errors, fmt.Sprintf("asset validation failed: %v", err))
	} else {
		result.ServerValidation = validationResult
	}
	logger.Info("STEP 4 COMPLETE: %d unique assets | duration %v", validationResult.TotalAssets, result.Duration)

	return result, nil
}

// extractMediaAssetsWithError pulls asset info from every widget in a canvas.
func extractMediaAssetsWithError(ctx context.Context, session *canvussdk.Session, canvas canvussdk.Canvas) ([]AssetInfo, []AssetInfo, error) {
	logger := logging.GetLogger()
	logger.Verbose("Getting widgets for canvas %q (ID: %s)", canvas.Name, canvas.ID)

	widgets, err := session.ListWidgets(ctx, canvas.ID, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list widgets for canvas %s: %w", canvas.ID, err)
	}
	logger.Verbose("Found %d widgets in canvas %q", len(widgets), canvas.Name)

	var assets, noHash []AssetInfo
	for _, widget := range widgets {
		a, n := extractAssetFromWidget(ctx, session, canvas, widget)
		if a != nil {
			assets = append(assets, *a)
		}
		if n != nil {
			noHash = append(noHash, *n)
		}
	}
	return assets, noHash, nil
}

// extractBackgroundAssets checks the canvas background for an image asset.
func extractBackgroundAssets(ctx context.Context, session *canvussdk.Session, canvas canvussdk.Canvas) ([]AssetInfo, []AssetInfo) {
	logger := logging.GetLogger()
	bg, err := session.GetCanvasBackground(ctx, canvas.ID)
	if err != nil {
		logger.Verbose("Failed to get background for canvas %q: %v", canvas.Name, err)
		return nil, nil
	}
	if bg.Image == nil || bg.Image.Hash == "" {
		return nil, nil
	}
	asset := AssetInfo{
		Hash:       bg.Image.Hash,
		WidgetType: "CanvasBackground",
		CanvasID:   canvas.ID,
		CanvasName: canvas.Name,
		WidgetID:   "background",
		WidgetName: "Canvas Background",
	}
	return []AssetInfo{asset}, nil
}

// countUniqueAssets returns a validation summary without downloading any files.
// Full server-side download validation was disabled to prevent OOM on large deployments.
func countUniqueAssets(assets []AssetInfo) (*ServerValidationResult, error) {
	unique := make(map[string]struct{})
	for _, a := range assets {
		if a.Hash != "" {
			unique[a.Hash] = struct{}{}
		}
	}
	n := len(unique)
	return &ServerValidationResult{
		TotalAssets:      n,
		ExistingAssets:   n,
		ValidationErrors: make([]string, 0),
	}, nil
}

// extractAssetFromWidget fetches per-type widget details and returns asset info.
// Returns (withHash, withoutHash) — either or both may be nil.
func extractAssetFromWidget(ctx context.Context, session *canvussdk.Session, canvas canvussdk.Canvas, widget canvussdk.Widget) (*AssetInfo, *AssetInfo) {
	logger := logging.GetLogger()
	logger.Verbose("Widget ID=%s type=%s in canvas %q", widget.ID, widget.WidgetType, canvas.Name)

	var details interface{}
	var err error

	switch widget.WidgetType {
	case "Image":
		details, err = session.GetImage(ctx, canvas.ID, widget.ID)
	case "Pdf":
		details, err = session.GetPDF(ctx, canvas.ID, widget.ID)
	case "Video":
		details, err = session.GetVideo(ctx, canvas.ID, widget.ID)
	default:
		return nil, nil // Not a media type.
	}

	if err != nil {
		logger.Verbose("Failed to get widget details ID=%s type=%s: %v", widget.ID, widget.WidgetType, err)
		return nil, nil
	}

	// Extract fields via reflection so this code works across SDK versions.
	hash, filename, name := reflectAssetFields(details, logger, widget.ID)

	if hash != "" {
		return &AssetInfo{
			Hash:             hash,
			WidgetType:       widget.WidgetType,
			OriginalFilename: filename,
			CanvasID:         canvas.ID,
			CanvasName:       canvas.Name,
			WidgetID:         widget.ID,
			WidgetName:       name,
		}, nil
	}
	if filename != "" {
		return nil, &AssetInfo{
			WidgetType:       widget.WidgetType,
			OriginalFilename: filename,
			CanvasID:         canvas.ID,
			CanvasName:       canvas.Name,
			WidgetID:         widget.ID,
			WidgetName:       name,
		}
	}
	return nil, nil
}

// reflectAssetFields extracts Hash, OriginalFilename, and Title/Name from a
// widget detail struct via reflection. This avoids type-asserting to three
// different concrete types (canvus.Image, canvus.PDF, canvus.Video).
func reflectAssetFields(details interface{}, logger *logging.Logger, widgetID string) (hash, filename, name string) {
	v := reflect.ValueOf(details)
	if !v.IsValid() || v.IsNil() {
		return
	}
	elem := v.Elem()

	if f := elem.FieldByName("Hash"); f.IsValid() && f.CanInterface() {
		if s, ok := f.Interface().(string); ok {
			hash = s
		}
	}
	if f := elem.FieldByName("OriginalFilename"); f.IsValid() && f.CanInterface() {
		if s, ok := f.Interface().(string); ok {
			filename = s
		}
	}
	// Prefer Title, fall back to Name.
	if f := elem.FieldByName("Title"); f.IsValid() && f.CanInterface() {
		if s, ok := f.Interface().(string); ok {
			name = s
		}
	} else if f := elem.FieldByName("Name"); f.IsValid() && f.CanInterface() {
		if s, ok := f.Interface().(string); ok {
			name = s
		}
	}

	logger.Verbose("Widget %s: hash=%q filename=%q name=%q", widgetID, hash, filename, name)
	return
}

// GetUniqueAssets de-duplicates assets by hash.
func (r *DiscoveryResult) GetUniqueAssets() []AssetInfo {
	seen := make(map[string]AssetInfo)
	for _, a := range r.Assets {
		if _, exists := seen[a.Hash]; !exists {
			seen[a.Hash] = a
		}
	}
	out := make([]AssetInfo, 0, len(seen))
	for _, a := range seen {
		out = append(out, a)
	}
	return out
}

// GetAssetsByCanvas groups assets by canvas name.
func (r *DiscoveryResult) GetAssetsByCanvas() map[string][]AssetInfo {
	m := make(map[string][]AssetInfo)
	for _, a := range r.Assets {
		m[a.CanvasName] = append(m[a.CanvasName], a)
	}
	return m
}
