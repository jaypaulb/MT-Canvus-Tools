package canvus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
)

// ImportWidgetsToRegion imports widgets/assets from an ExportedWidgetSet into
// a region of a canvas. Widgets are scaled+translated to fit the region.
func (s *Session) ImportWidgetsToRegion(ctx context.Context, canvasID string, exported *ExportedWidgetSet, targetRegion Rectangle) ([]string, error) {
	if exported == nil || len(exported.Widgets) == 0 {
		return nil, nil
	}
	if exported.Region == nil {
		return nil, fmt.Errorf("ImportWidgetsToRegion: original region is nil")
	}
	scaleX := targetRegion.Width / exported.Region.Width
	scaleY := targetRegion.Height / exported.Region.Height
	dx := targetRegion.X - exported.Region.X*scaleX
	dy := targetRegion.Y - exported.Region.Y*scaleY

	var newIDs []string
	for _, w := range exported.Widgets {
		if w.Location != nil {
			w.Location.X = w.Location.X*scaleX + dx
			w.Location.Y = w.Location.Y*scaleY + dy
		}
		if w.Size != nil {
			w.Size.Width *= scaleX
			w.Size.Height *= scaleY
		}
		var createdID string
		widgetType := strings.ToLower(w.WidgetType)
		switch widgetType {
		case "image":
			data, ok := exported.Assets[w.ID]
			if !ok {
				return nil, fmt.Errorf("ImportWidgetsToRegion: missing image data for widget %s", w.ID)
			}
			meta := importMeta(w, widgetType)
			body, contentType, err := buildMultipartBody(meta, "data", "imported_image.jpg", []byte(data))
			if err != nil {
				return nil, fmt.Errorf("ImportWidgetsToRegion: image multipart: %w", err)
			}
			img, err := s.CreateImage(ctx, canvasID, body, contentType)
			if err != nil {
				return nil, fmt.Errorf("ImportWidgetsToRegion: create image: %w", err)
			}
			createdID = img.ID
		case "pdf":
			data, ok := exported.Assets[w.ID]
			if !ok {
				return nil, fmt.Errorf("ImportWidgetsToRegion: missing PDF data for widget %s", w.ID)
			}
			meta := importMeta(w, widgetType)
			body, contentType, err := buildMultipartBody(meta, "data", "imported.pdf", []byte(data))
			if err != nil {
				return nil, fmt.Errorf("ImportWidgetsToRegion: pdf multipart: %w", err)
			}
			pdf, err := s.CreatePDF(ctx, canvasID, body, contentType)
			if err != nil {
				return nil, fmt.Errorf("ImportWidgetsToRegion: create pdf: %w", err)
			}
			createdID = pdf.ID
		case "video":
			data, ok := exported.Assets[w.ID]
			if !ok {
				return nil, fmt.Errorf("ImportWidgetsToRegion: missing video data for widget %s", w.ID)
			}
			meta := importMeta(w, widgetType)
			body, contentType, err := buildMultipartBody(meta, "data", "imported_video.mp4", []byte(data))
			if err != nil {
				return nil, fmt.Errorf("ImportWidgetsToRegion: video multipart: %w", err)
			}
			video, err := s.CreateVideo(ctx, canvasID, body, contentType)
			if err != nil {
				return nil, fmt.Errorf("ImportWidgetsToRegion: create video: %w", err)
			}
			createdID = video.ID
		default:
			created, err := s.CreateWidget(ctx, canvasID, widgetToMap(w))
			if err != nil {
				return nil, fmt.Errorf("ImportWidgetsToRegion: create %s: %w", widgetType, err)
			}
			createdID = created.ID
		}
		newIDs = append(newIDs, createdID)
	}
	return newIDs, nil
}

func importMeta(w Widget, widgetType string) map[string]any {
	meta := map[string]any{
		"title":       w.ID,
		"widget_type": widgetType,
	}
	if w.Location != nil {
		meta["location"] = map[string]any{"x": w.Location.X, "y": w.Location.Y}
	}
	if w.Size != nil {
		meta["size"] = map[string]any{"width": w.Size.Width, "height": w.Size.Height}
	}
	return meta
}

func widgetToMap(w Widget) map[string]any {
	m := map[string]any{
		"widget_type": strings.ToLower(w.WidgetType),
		"parent_id":   w.ParentID,
		"pinned":      w.Pinned,
		"scale":       w.Scale,
		"state":       w.State,
		"depth":       w.Depth,
	}
	if w.Location != nil {
		m["location"] = map[string]any{"x": w.Location.X, "y": w.Location.Y}
	}
	if w.Size != nil {
		m["size"] = map[string]any{"width": w.Size.Width, "height": w.Size.Height}
	}
	return m
}

func buildMultipartBody(meta map[string]any, fieldName, filename string, fileData []byte) (io.Reader, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if meta != nil {
		jsonBytes, err := json.Marshal(meta)
		if err != nil {
			return nil, "", err
		}
		jsonPart, err := w.CreateFormField("json")
		if err != nil {
			return nil, "", err
		}
		if _, err := jsonPart.Write(jsonBytes); err != nil {
			return nil, "", err
		}
	}
	if fileData != nil && fieldName != "" && filename != "" {
		filePart, err := w.CreateFormFile(fieldName, filename)
		if err != nil {
			return nil, "", err
		}
		if _, err := filePart.Write(fileData); err != nil {
			return nil, "", err
		}
	}
	_ = w.Close()
	return &buf, w.FormDataContentType(), nil
}
