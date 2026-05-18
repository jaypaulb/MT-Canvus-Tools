// Command 08-cross-canvas-clone clones a single widget from one Canvus
// canvas to another using the SDK's CloneWidget helper, which exercises
// the per-type create endpoint with {source_canvas_id, source_widget_id}
// fields per the Phase 3 changelog §1 documentation update.
//
// The legacy POST /api/v1/canvases/{id}/widgets/clone endpoint returns
// 501 Not Implemented and is deliberately not used.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

// widgetTypeToPath maps the lowercase widget_type discriminator returned by
// the server to the URL path segment used by the per-type create endpoint.
// Asset-based and primitive widget types share the same shape: type-name
// plural-ised.
var widgetTypeToPath = map[string]string{
	"note":    "notes",
	"image":   "images",
	"video":   "videos",
	"pdf":     "pdfs",
	"browser": "browsers",
	"anchor":  "anchors",
	"table":   "tables",
}

func main() {
	setupLogging()
	if err := run(); err != nil {
		slog.Error("cross-canvas-clone failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	baseURL, err := mustEnv("CANVUS_BASE_URL")
	if err != nil {
		return err
	}
	apiKey, err := mustEnv("CANVUS_API_KEY")
	if err != nil {
		return err
	}
	srcCanvasID, err := mustEnv("CANVUS_CANVAS_ID")
	if err != nil {
		return err
	}
	dstCanvasID, err := mustEnv("CANVUS_DEST_CANVAS_ID")
	if err != nil {
		return err
	}
	srcWidgetID, err := mustEnv("CANVUS_SOURCE_WIDGET_ID")
	if err != nil {
		return err
	}

	cfg := &canvus.SessionConfig{BaseURL: baseURL}
	s := canvus.NewSession(cfg, canvus.WithAPIKey(apiKey))
	ctx := context.Background()

	// Step 1 — fetch the source widget to learn its type.
	src, err := s.GetWidget(ctx, srcCanvasID, srcWidgetID)
	if err != nil {
		return fmt.Errorf("GetWidget (source): %w", err)
	}
	slog.Info("source widget fetched",
		"widget_id", src.ID,
		"widget_type", src.WidgetType,
		"location", pointString(src.Location))

	// Step 2 — resolve widget type → URL path segment.
	wtype := strings.ToLower(src.WidgetType)
	pathSeg, ok := widgetTypeToPath[wtype]
	if !ok {
		return fmt.Errorf("widget type %q is not cloneable via this example "+
			"(supported: note/image/video/pdf/browser/anchor/table)", src.WidgetType)
	}

	// Step 3 — call the SDK clone helper. Passing nil for location keeps
	// the source widget's original coordinates on the destination.
	cloned, err := s.CloneWidget(ctx, dstCanvasID, srcCanvasID, srcWidgetID, pathSeg, nil)
	if err != nil {
		return fmt.Errorf("CloneWidget: %w", err)
	}
	slog.Info("widget cloned",
		"source_widget_id", srcWidgetID,
		"cloned_widget_id", cloned.ID,
		"dest_canvas_id", dstCanvasID,
		"widget_type", cloned.WidgetType,
		"location", pointString(cloned.Location))

	fmt.Printf("cloned widget id on destination: %s\n", cloned.ID)

	// Step 4 — optional cleanup.
	if os.Getenv("CANVUS_CLEANUP") == "1" {
		slog.Info("CANVUS_CLEANUP=1 — sleeping 2s then deleting cloned widget")
		time.Sleep(2 * time.Second)
		if err := s.DeleteWidget(ctx, dstCanvasID, cloned.ID, wtype); err != nil {
			return fmt.Errorf("DeleteWidget (cleanup): %w", err)
		}
		slog.Info("cloned widget deleted", "cloned_widget_id", cloned.ID)
	} else {
		slog.Info("CANVUS_CLEANUP unset — leaving cloned widget on destination canvas")
	}
	return nil
}

func pointString(p *canvus.Point) string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("(%.1f, %.1f)", p.X, p.Y)
}

func mustEnv(name string) (string, error) {
	v := os.Getenv(name)
	if v == "" {
		return "", fmt.Errorf("missing required env var %s", name)
	}
	return v, nil
}

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
