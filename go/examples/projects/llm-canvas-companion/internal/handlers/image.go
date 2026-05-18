package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/llm-canvas-companion/internal/ai"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/llm-canvas-companion/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// handleImageGeneration generates an AI image and uploads it to the canvas
// adjacent to the triggering widget.
//
// The image bytes are streamed directly into the multipart body — no temp
// file is needed. (cfg is passed for symmetry with the other handlers and
// to support future config-driven knobs like model selection.)
func handleImageGeneration(ctx context.Context, s *canvus.Session, cfg *config.Config, canvasID string, w canvus.Widget, prompt string) error {
	imageURL, err := ai.GenerateImageURL(ctx, cfg, prompt)
	if err != nil {
		return fmt.Errorf("handleImageGeneration: generate URL: %w", err)
	}

	imageData, err := downloadURL(ctx, imageURL)
	if err != nil {
		return fmt.Errorf("handleImageGeneration: download image: %w", err)
	}

	// Calculate position offset from triggering widget.
	x, y := 0.0, 0.0
	if w.Location != nil {
		x = w.Location.X
		y = w.Location.Y
	}
	width, height := 400.0, 300.0
	if w.Size != nil {
		width = w.Size.Width
		height = w.Size.Height
	}

	meta := map[string]any{
		"title": fmt.Sprintf("AI Generated Image for %s", w.ID),
		"location": map[string]any{
			"x": x + width*0.8,
			"y": y + height*0.8,
		},
		"size":  map[string]any{"width": width, "height": height},
		"depth": w.Depth + 10,
		"scale": w.Scale / 3,
	}

	body, ct, err := buildImageMultipart(meta, "image.jpg", imageData)
	if err != nil {
		return fmt.Errorf("handleImageGeneration: build multipart: %w", err)
	}

	slog.Info("handleImageGeneration: uploading image", "widget_id", w.ID)
	if _, err := s.CreateImage(ctx, canvasID, body, ct); err != nil {
		return fmt.Errorf("handleImageGeneration: CreateImage: %w", err)
	}
	return nil
}

// downloadURL fetches a URL and returns its body bytes.
func downloadURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("downloadURL: new request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloadURL: GET: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloadURL: HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 50*1024*1024))
}

// buildImageMultipart constructs a multipart body with a 'json' metadata part
// and a 'data' file part — the format required by the Canvus images endpoint.
func buildImageMultipart(meta map[string]any, filename string, fileData []byte) (io.Reader, string, error) {
	jsonBytes, err := json.Marshal(meta)
	if err != nil {
		return nil, "", fmt.Errorf("buildImageMultipart: encode meta: %w", err)
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	jsonPart, err := mw.CreateFormField("json")
	if err != nil {
		return nil, "", fmt.Errorf("buildImageMultipart: create json part: %w", err)
	}
	if _, err := jsonPart.Write(jsonBytes); err != nil {
		return nil, "", fmt.Errorf("buildImageMultipart: write json part: %w", err)
	}

	filePart, err := mw.CreateFormFile("data", filename)
	if err != nil {
		return nil, "", fmt.Errorf("buildImageMultipart: create data part: %w", err)
	}
	if _, err := filePart.Write(fileData); err != nil {
		return nil, "", fmt.Errorf("buildImageMultipart: write data part: %w", err)
	}

	if err := mw.Close(); err != nil {
		return nil, "", fmt.Errorf("buildImageMultipart: close writer: %w", err)
	}
	return &buf, mw.FormDataContentType(), nil
}
