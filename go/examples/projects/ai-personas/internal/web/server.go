// Package web exposes a small HTTP dashboard for ai-personas. It hosts a
// question-submission form, renders a QR code pointing at itself, and
// publishes a /health endpoint suitable for liveness/readiness probes.
package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	fqdn "github.com/Showmax/go-fqdn"
	"github.com/skip2/go-qrcode"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/canvasx"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

//go:embed static/question.html
var staticFS embed.FS

// Version is overridable at build time via -ldflags "-X .../web.Version=...".
var Version = "dev"

// HealthStatus is one of the textual health-state values.
type HealthStatus string

// Known health states.
const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// healthResponse is the body of GET /health.
type healthResponse struct {
	Status  HealthStatus `json:"status"`
	Uptime  string       `json:"uptime"`
	Version string       `json:"version"`
	Details struct {
		CanvusAPI bool `json:"canvus_api"`
	} `json:"details"`
}

// Server is the HTTP dashboard.
type Server struct {
	Session      *canvus.Session
	CanvasID     string
	Port         string
	PublicWebURL string

	startTime time.Time
}

// New returns a Server that talks to the configured Canvus session+canvas.
func New(s *canvus.Session, canvasID, port, publicWebURL string) *Server {
	return &Server{
		Session:      s,
		CanvasID:     canvasID,
		Port:         port,
		PublicWebURL: publicWebURL,
		startTime:    time.Now(),
	}
}

// WebURL returns the URL the QR code should encode.
func (s *Server) WebURL() string {
	if s.PublicWebURL != "" {
		return s.PublicWebURL
	}
	host, err := fqdn.FqdnHostname()
	if err != nil || host == "" {
		host, _ = os.Hostname()
	}
	return "http://" + host + ":" + s.Port + "/"
}

// Start launches the QR watcher goroutine and the HTTP listener. Both run
// until ctx is cancelled.
func (s *Server) Start(ctx context.Context) {
	webURL := s.WebURL()
	go s.qrCodeLoop(ctx, webURL)

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/health", s.handleHealth)

	srv := &http.Server{
		Addr:              ":" + s.Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		slog.Info("web: listening", "port", s.Port, "url", webURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("web: server error", "err", err)
		}
	}()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
}

// handleRoot serves the question form (GET) or processes a submission (POST).
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data, err := staticFS.ReadFile("static/question.html")
		if err != nil {
			http.Error(w, "Question page not available", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(data)
	case http.MethodPost:
		s.handleSubmit(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSubmit creates a New_AI_Question note inside the Remote anchor zone.
func (s *Server) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	question := strings.TrimSpace(r.FormValue("question"))
	if question == "" {
		http.Error(w, "Question required", http.StatusBadRequest)
		return
	}
	if !strings.HasSuffix(question, "?") {
		question += "?"
	}

	ctx := r.Context()
	anchor, err := s.findRemoteAnchor(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	notes, err := s.Session.ListNotes(ctx, s.CanvasID)
	if err != nil {
		http.Error(w, "Failed to list notes", http.StatusInternalServerError)
		return
	}
	images, err := s.Session.ListImages(ctx, s.CanvasID)
	if err != nil {
		http.Error(w, "Failed to list images", http.StatusInternalServerError)
		return
	}
	noteX, noteY, noteW, noteH, scale, err := findFreeSegment(notes, images, anchor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if _, err := s.Session.CreateNote(ctx, s.CanvasID, map[string]any{
		"title":            "New_AI_Question",
		"text":             question,
		"location":         map[string]any{"x": noteX - noteW*scale/2, "y": noteY - noteH*scale/2},
		"size":             map[string]any{"width": noteW, "height": noteH},
		"scale":            scale,
		"background_color": canvasx.ColorWhiteAIQuestion,
	}); err != nil {
		http.Error(w, "Failed to create note: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte("Question submitted!"))
}

// handleHealth answers GET /health with a probe of the Canvus API.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_, err := s.Session.ListNotes(r.Context(), s.CanvasID)
	canvusOK := err == nil
	resp := healthResponse{
		Status:  HealthStatusHealthy,
		Uptime:  formatUptime(time.Since(s.startTime)),
		Version: Version,
	}
	resp.Details.CanvusAPI = canvusOK
	status := http.StatusOK
	if !canvusOK {
		resp.Status = HealthStatusUnhealthy
		status = http.StatusServiceUnavailable
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// findRemoteAnchor returns the "Remote" anchor on the canvas, or an error.
func (s *Server) findRemoteAnchor(ctx context.Context) (*canvus.Anchor, error) {
	anchors, err := s.Session.ListAnchors(ctx, s.CanvasID)
	if err != nil {
		return nil, fmt.Errorf("list anchors: %w", err)
	}
	for i := range anchors {
		if strings.EqualFold(strings.TrimSpace(anchors[i].AnchorName), "Remote") {
			return &anchors[i], nil
		}
	}
	return nil, fmt.Errorf("Remote anchor not found")
}

// findFreeSegment locates a free 5x4 grid cell inside the Remote anchor.
// Cell 0 (top-left) is reserved for the QR code; cells overlapped by existing
// notes/images are considered taken.
func findFreeSegment(notes []canvus.Note, images []canvus.Image, anchor *canvus.Anchor) (noteX, noteY, noteW, noteH, scale float64, err error) {
	if anchor.Location == nil || anchor.Size == nil {
		return 0, 0, 0, 0, 0, fmt.Errorf("Remote anchor missing location/size")
	}
	ax, ay := anchor.Location.X, anchor.Location.Y
	aw, ah := anchor.Size.Width, anchor.Size.Height
	const cols, rows = 5, 4
	segW := aw / float64(cols)
	segH := ah / float64(rows)

	used := make([]bool, cols*rows)
	mark := func(wx, wy, ww, wh float64) {
		for row := 0; row < rows; row++ {
			for col := 0; col < cols; col++ {
				segX := ax + float64(col)*segW
				segY := ay + float64(row)*segH
				if wx < segX+segW && wx+ww > segX && wy < segY+segH && wy+wh > segY {
					used[row*cols+col] = true
				}
			}
		}
	}
	for _, n := range notes {
		if n.Location != nil && n.Size != nil {
			mark(n.Location.X, n.Location.Y, n.Size.Width, n.Size.Height)
		}
	}
	for _, im := range images {
		if im.Location != nil && im.Size != nil {
			mark(im.Location.X, im.Location.Y, im.Size.Width, im.Size.Height)
		}
	}
	used[0] = true // reserve cell 0 for QR

	idx := -1
	for i := 1; i < cols*rows; i++ {
		if !used[i] {
			idx = i
			break
		}
	}
	if idx < 0 {
		return 0, 0, 0, 0, 0, fmt.Errorf("Anchor is full: no free segments available")
	}
	segCol, segRow := idx%cols, idx/cols
	noteX = ax + float64(segCol)*segW + segW/2
	noteY = ay + float64(segRow)*segH + segH/2
	noteW = segW * (2.0 / 3.0)
	noteH = segH * (2.0 / 3.0)
	scale = 1.5 / 3.5
	return
}

// qrCodeLoop creates the QR image on the canvas, subscribes to its widget
// stream, and recreates it whenever it's deleted (e.g. by a user clearing
// the canvas).
func (s *Server) qrCodeLoop(ctx context.Context, webURL string) {
	for {
		if ctx.Err() != nil {
			return
		}
		qrID, err := s.createOrReplaceQRCode(ctx, webURL)
		if err != nil {
			slog.Warn("web: QR setup failed; retrying", "err", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}
		slog.Info("web: QR ready", "id", qrID, "url", webURL)
		s.watchQRImage(ctx, qrID)
	}
}

// watchQRImage subscribes to the QR image widget and returns when ctx is
// cancelled or the image is deleted (the latter triggers a recreate above).
func (s *Server) watchQRImage(ctx context.Context, qrID string) {
	events, err := s.Session.SubscribeImage(ctx, s.CanvasID, qrID)
	if err != nil {
		slog.Warn("web: subscribe QR image failed", "err", err)
		time.Sleep(5 * time.Second)
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-events:
			if !ok {
				slog.Info("web: QR subscription ended")
				return
			}
			if ev.State == "deleted" {
				slog.Info("web: QR image deleted; will recreate", "id", qrID)
				return
			}
		}
	}
}

// createOrReplaceQRCode generates a fresh QR PNG, deletes any existing
// "Remote QR" image, uploads the new one, and returns its widget ID.
func (s *Server) createOrReplaceQRCode(ctx context.Context, webURL string) (string, error) {
	png, err := qrcode.Encode(webURL, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("qrcode encode: %w", err)
	}

	images, err := s.Session.ListImages(ctx, s.CanvasID)
	if err != nil {
		return "", fmt.Errorf("list images: %w", err)
	}
	for _, im := range images {
		if im.Title == "Remote QR" {
			if err := s.Session.DeleteImage(ctx, s.CanvasID, im.ID); err != nil {
				slog.Warn("web: delete old QR failed", "id", im.ID, "err", err)
			}
		}
	}

	anchor, err := s.findRemoteAnchor(ctx)
	if err != nil {
		return "", err
	}
	if anchor.Location == nil || anchor.Size == nil {
		return "", fmt.Errorf("Remote anchor missing location/size")
	}
	qrW := anchor.Size.Width / 20.0
	qrH := anchor.Size.Height / 20.0
	meta := map[string]any{
		"title":    "Remote QR",
		"location": map[string]any{"x": anchor.Location.X, "y": anchor.Location.Y},
		"size":     map[string]any{"width": qrW, "height": qrH},
	}
	body, contentType, err := canvasx.BuildImageMultipart(meta, "remote-qr.png", png)
	if err != nil {
		return "", err
	}
	img, err := s.Session.CreateImage(ctx, s.CanvasID, body, contentType)
	if err != nil {
		return "", fmt.Errorf("upload QR: %w", err)
	}
	return img.ID, nil
}

// formatUptime renders a Duration as Xd Yh Zm Ws (skipping leading zeros).
func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	case hours > 0:
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	case minutes > 0:
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

