// Command 04-file-upload uploads a local PNG to a Canvus canvas as an image
// widget, repositions it, and (optionally) cleans up afterwards.
//
// If CANVUS_IMAGE_PATH is unset the program generates a tiny 256x256 solid-
// colour PNG via the image/png stdlib so the example is fully self-contained
// — no external assets required.
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

const (
	defaultSamplePath = "sample.png"
	sampleSize        = 256
)

func main() {
	setupLogging()
	if err := run(); err != nil {
		slog.Error("file-upload failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	baseURL, err := mustEnv("CANVUS_API_URL")
	if err != nil {
		return err
	}
	apiKey, err := mustEnv("CANVUS_API_KEY")
	if err != nil {
		return err
	}

	cfg := &canvus.SessionConfig{BaseURL: baseURL}
	s := canvus.NewSession(cfg, canvus.WithAPIKey(apiKey))
	ctx := context.Background()

	canvasID, err := resolveCanvasID(ctx, s)
	if err != nil {
		return fmt.Errorf("resolve canvas id: %w", err)
	}
	slog.Info("target canvas", "canvas_id", canvasID)

	// Resolve the image: env-supplied path or generated sample.
	imagePath, generated, err := resolveImagePath()
	if err != nil {
		return fmt.Errorf("resolve image path: %w", err)
	}
	slog.Info("image source", "path", imagePath, "generated", generated)

	body, contentType, err := buildMultipart(imagePath)
	if err != nil {
		return fmt.Errorf("build multipart body: %w", err)
	}

	img, err := s.CreateImage(ctx, canvasID, body, contentType)
	if err != nil {
		return fmt.Errorf("CreateImage: %w", err)
	}
	slog.Info("image uploaded",
		"widget_id", img.ID,
		"asset_hash", img.Hash,
		"original_filename", img.OriginalFilename)

	// Defer cleanup unless CANVUS_KEEP_WIDGET=1.
	keep := os.Getenv("CANVUS_KEEP_WIDGET") == "1"
	if !keep {
		defer func() {
			if err := s.DeleteImage(context.Background(), canvasID, img.ID); err != nil {
				slog.Warn("cleanup DeleteImage failed", "widget_id", img.ID, "err", err)
				return
			}
			slog.Info("cleanup — image deleted", "widget_id", img.ID)
		}()
	} else {
		slog.Info("CANVUS_KEEP_WIDGET=1 — leaving widget on canvas")
	}

	// Reposition + scale. The endpoint pins widget_type as a server-set
	// discriminator; sending it in the PATCH body is unnecessary.
	patch := map[string]any{
		"location": map[string]any{"x": 100, "y": 200},
		"scale":    0.5,
	}
	updated, err := s.UpdateImage(ctx, canvasID, img.ID, patch)
	if err != nil {
		return fmt.Errorf("UpdateImage: %w", err)
	}
	loc := pointString(updated.Location)
	fmt.Printf("widget id: %s  position: %s  scale: %.2f\n", updated.ID, loc, updated.Scale)
	slog.Info("image repositioned",
		"widget_id", updated.ID,
		"location", loc,
		"scale", updated.Scale)
	return nil
}

// resolveCanvasID returns CANVUS_CANVAS_ID if set, otherwise picks the first
// canvas with read+write access from ListCanvases.
func resolveCanvasID(ctx context.Context, s *canvus.Session) (string, error) {
	if v := os.Getenv("CANVUS_CANVAS_ID"); v != "" {
		return v, nil
	}
	canvases, err := s.ListCanvases(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("ListCanvases (auto-select): %w", err)
	}
	for _, c := range canvases {
		if c.Access == "rw" && !c.InTrash {
			return c.ID, nil
		}
	}
	return "", errors.New("no canvas with rw access found; set CANVUS_CANVAS_ID explicitly")
}

// resolveImagePath returns the upload path. If CANVUS_IMAGE_PATH is set, it
// is used. Otherwise a 256x256 solid-colour PNG is written to ./sample.png
// and that path returned.
func resolveImagePath() (string, bool, error) {
	if p := os.Getenv("CANVUS_IMAGE_PATH"); p != "" {
		if _, err := os.Stat(p); err != nil {
			return "", false, fmt.Errorf("CANVUS_IMAGE_PATH=%q: %w", p, err)
		}
		return p, false, nil
	}
	if _, err := os.Stat(defaultSamplePath); err == nil {
		return defaultSamplePath, false, nil
	}
	if err := writeSamplePNG(defaultSamplePath); err != nil {
		return "", false, fmt.Errorf("generate sample.png: %w", err)
	}
	return defaultSamplePath, true, nil
}

// writeSamplePNG generates a tiny solid-colour PNG with a one-pixel border
// for visual sanity-checking on the canvas.
func writeSamplePNG(path string) error {
	img := image.NewRGBA(image.Rect(0, 0, sampleSize, sampleSize))
	fill := color.RGBA{R: 0x1d, G: 0x71, B: 0xb8, A: 0xff} // MT blue
	border := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	for y := 0; y < sampleSize; y++ {
		for x := 0; x < sampleSize; x++ {
			if x == 0 || y == 0 || x == sampleSize-1 || y == sampleSize-1 {
				img.Set(x, y, border)
			} else {
				img.Set(x, y, fill)
			}
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return png.Encode(f, img)
}

// buildMultipart packages the file as a multipart/form-data body with the
// canonical {"json": <metadata>, "data": <bytes>} shape the Canvus image
// endpoint expects.
func buildMultipart(path string) (io.Reader, string, error) {
	f, err := os.Open(path) //nolint:gosec // path is user-supplied env-var, intended use.
	if err != nil {
		return nil, "", fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	buf := &bytes.Buffer{}
	mw := multipart.NewWriter(buf)

	// Metadata part — minimum {title} is enough; the server fills the rest.
	jsonPart, err := mw.CreateFormField("json")
	if err != nil {
		return nil, "", fmt.Errorf("create json part: %w", err)
	}
	meta := fmt.Sprintf(`{"title":%q}`, filepath.Base(path))
	if _, err := jsonPart.Write([]byte(meta)); err != nil {
		return nil, "", fmt.Errorf("write json part: %w", err)
	}

	// Binary part — field name is "data" per the API spec.
	dataPart, err := mw.CreateFormFile("data", filepath.Base(path))
	if err != nil {
		return nil, "", fmt.Errorf("create data part: %w", err)
	}
	if _, err := io.Copy(dataPart, f); err != nil {
		return nil, "", fmt.Errorf("copy file into multipart: %w", err)
	}
	if err := mw.Close(); err != nil {
		return nil, "", fmt.Errorf("close multipart writer: %w", err)
	}
	return buf, mw.FormDataContentType(), nil
}

func pointString(p *canvus.Point) string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("(%.1f, %.1f)", p.X, p.Y)
}

// mustEnv returns the value of name or an error if it is empty/unset.
func mustEnv(name string) (string, error) {
	v := os.Getenv(name)
	if v == "" {
		return "", fmt.Errorf("missing required env var %s", name)
	}
	return v, nil
}

// setupLogging configures slog.Default.
func setupLogging() {
	var h slog.Handler
	switch os.Getenv("LOG_FORMAT") {
	case "json":
		h = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	default:
		h = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	slog.SetDefault(slog.New(h))
}
