package canvus

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ExportedWidgetSet is the on-disk + in-memory representation of widgets and
// assets exported from a canvas region.
type ExportedWidgetSet struct {
	Widgets []Widget
	Assets  map[string]string
	Region  *Rectangle
}

// ExportWidgetsToFolder exports the specified widgets (and asset files) from a
// canvas to a folder on disk and returns the folder path.
//
// sharedCanvasID, when set, blanks parent_id on widgets whose parent is the
// shared canvas — making the exported set portable to canvases with different
// SharedCanvas IDs.
func (s *Session) ExportWidgetsToFolder(ctx context.Context, canvasID string, widgetIDs []string, region Rectangle, sharedCanvasID, baseFolder string) (string, error) {
	if baseFolder == "" {
		baseFolder = filepath.Join("export", time.Now().Format("20060102_150405"))
	}
	exportFolder := baseFolder
	if err := os.MkdirAll(exportFolder, 0o755); err != nil {
		return "", err
	}

	log := slog.With("op", "export", "canvas", canvasID, "folder", exportFolder)
	var selected []Widget
	assets := make(map[string]string)

	for _, id := range widgetIDs {
		w, err := s.GetWidget(ctx, canvasID, id)
		if err != nil {
			return "", fmt.Errorf("ExportWidgetsToFolder: get widget %s: %w", id, err)
		}
		if sharedCanvasID != "" && w.ParentID == sharedCanvasID {
			w.ParentID = ""
		}
		selected = append(selected, *w)
		widgetType := strings.ToLower(w.WidgetType)
		switch widgetType {
		case "image":
			data, err := s.DownloadImage(ctx, canvasID, w.ID)
			if err != nil {
				return "", fmt.Errorf("ExportWidgetsToFolder: download image %s: %w", w.ID, err)
			}
			filename := "image_" + w.ID + ".jpg"
			if err := os.WriteFile(filepath.Join(exportFolder, filename), data, 0o644); err != nil {
				return "", fmt.Errorf("ExportWidgetsToFolder: write image: %w", err)
			}
			assets[w.ID] = filename
		case "pdf":
			data, err := s.DownloadPDF(ctx, canvasID, w.ID)
			if err != nil {
				return "", fmt.Errorf("ExportWidgetsToFolder: download pdf %s: %w", w.ID, err)
			}
			filename := "pdf_" + w.ID + ".pdf"
			if err := os.WriteFile(filepath.Join(exportFolder, filename), data, 0o644); err != nil {
				return "", fmt.Errorf("ExportWidgetsToFolder: write pdf: %w", err)
			}
			assets[w.ID] = filename
		case "video":
			data, err := s.DownloadVideo(ctx, canvasID, w.ID)
			if err != nil {
				return "", fmt.Errorf("ExportWidgetsToFolder: download video %s: %w", w.ID, err)
			}
			filename := "video_" + w.ID + ".mp4"
			if err := os.WriteFile(filepath.Join(exportFolder, filename), data, 0o644); err != nil {
				return "", fmt.Errorf("ExportWidgetsToFolder: write video: %w", err)
			}
			assets[w.ID] = filename
		}
	}

	exportJSON := struct {
		Widgets []Widget          `json:"widgets"`
		Assets  map[string]string `json:"assets"`
		Region  *Rectangle        `json:"region"`
	}{Widgets: selected, Assets: assets, Region: &region}

	jsonBytes, err := json.MarshalIndent(exportJSON, "", "  ")
	if err != nil {
		return "", fmt.Errorf("ExportWidgetsToFolder: marshal export JSON: %w", err)
	}
	if err := os.WriteFile(filepath.Join(exportFolder, "export.json"), jsonBytes, 0o644); err != nil {
		return "", fmt.Errorf("ExportWidgetsToFolder: write export JSON: %w", err)
	}
	log.Info("export complete", "widget_count", len(selected), "asset_count", len(assets))
	return exportFolder, nil
}
