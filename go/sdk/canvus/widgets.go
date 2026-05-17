package canvus

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ErrWidgetTypeNotCreatable is returned when a caller attempts to create a
// widget of a type the server refuses to construct via the API. Per
// changelog §2, this currently applies to IP Video and RDP Connection.
var ErrWidgetTypeNotCreatable = errors.New("widget type can only be created from the Canvus desktop client")

// ListWidgets retrieves all widgets for a given canvas. If filter is non-nil,
// results are filtered client-side. If includeAnnotations is true, the
// annotations array is populated on each widget.
//
// The /canvases/{id}/widgets endpoint is read-only: POST, PATCH, DELETE are
// not supported here; use the type-specific helpers.
func (s *Session) ListWidgets(ctx context.Context, canvasID string, filter *Filter, includeAnnotations ...bool) ([]Widget, error) {
	var widgets []Widget
	var qp map[string]string
	if len(includeAnnotations) > 0 && includeAnnotations[0] {
		qp = map[string]string{"annotations": "1"}
	}
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/widgets", canvasID), nil, &widgets, qp, false); err != nil {
		return nil, fmt.Errorf("ListWidgets: %w", err)
	}
	if filter != nil {
		widgets = FilterSlice(widgets, filter)
	}
	return widgets, nil
}

// GetWidget retrieves a widget by ID for a given canvas.
func (s *Session) GetWidget(ctx context.Context, canvasID, widgetID string) (*Widget, error) {
	var widget Widget
	if err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("canvases/%s/widgets/%s", canvasID, widgetID), nil, &widget, nil, false); err != nil {
		return nil, fmt.Errorf("GetWidget: %w", err)
	}
	return &widget, nil
}

// CreateWidget dispatches to the type-specific create helper based on the
// `widget_type` key in req. Multipart types (image/pdf/video) require req to
// be an io.Reader and contentType to be supplied.
//
// IP Video and RDP Connection widget types are rejected per changelog §2:
// the server's createElement whitelist does not include them.
func (s *Session) CreateWidget(ctx context.Context, canvasID string, req any, contentType ...string) (*Widget, error) {
	if m, ok := req.(map[string]any); ok {
		widgetType, ok := m["widget_type"].(string)
		if !ok {
			return nil, fmt.Errorf("CreateWidget: missing or invalid widget_type in request")
		}
		switch strings.ToLower(widgetType) {
		case "note":
			n, err := s.CreateNote(ctx, canvasID, m)
			if err != nil {
				return nil, err
			}
			return widgetFromNote(n), nil
		case "anchor":
			a, err := s.CreateAnchor(ctx, canvasID, m)
			if err != nil {
				return nil, err
			}
			return widgetFromAnchor(a), nil
		case "browser":
			b, err := s.CreateBrowser(ctx, canvasID, m)
			if err != nil {
				return nil, err
			}
			return widgetFromBrowser(b), nil
		case "connector":
			c, err := s.CreateConnector(ctx, canvasID, m)
			if err != nil {
				return nil, err
			}
			return &Widget{ID: c.ID, WidgetType: "connector"}, nil
		case "table":
			t, err := s.CreateTable(ctx, canvasID, m)
			if err != nil {
				return nil, err
			}
			return widgetFromTable(t), nil
		case "ip_video", "ipvideo", "ip-video":
			return nil, fmt.Errorf("CreateWidget(%s): %w", widgetType, ErrWidgetTypeNotCreatable)
		case "rdp_connection", "rdpconnection", "rdp-connection":
			return nil, fmt.Errorf("CreateWidget(%s): %w", widgetType, ErrWidgetTypeNotCreatable)
		case "image", "pdf", "video":
			return nil, fmt.Errorf("CreateWidget: for widget_type %q, req must be a multipart body (io.Reader) and contentType must be set", widgetType)
		default:
			return nil, fmt.Errorf("CreateWidget: unsupported widget_type: %s", widgetType)
		}
	}
	if rdr, ok := req.(io.Reader); ok {
		if len(contentType) == 0 {
			return nil, fmt.Errorf("CreateWidget: contentType must be provided for multipart widget creation")
		}
		ct := contentType[0]
		if img, err := s.CreateImage(ctx, canvasID, rdr, ct); err == nil {
			return widgetFromImage(img), nil
		}
		if pdf, err := s.CreatePDF(ctx, canvasID, rdr, ct); err == nil {
			return &Widget{ID: pdf.ID, WidgetType: "pdf"}, nil
		}
		if vid, err := s.CreateVideo(ctx, canvasID, rdr, ct); err == nil {
			return &Widget{ID: vid.ID, WidgetType: "video"}, nil
		}
		return nil, fmt.Errorf("CreateWidget: failed to create image, pdf, or video")
	}
	return nil, fmt.Errorf("CreateWidget: req must be a map[string]interface{} (for note, anchor, browser, connector, table) or io.Reader (for image, pdf, video)")
}

// UpdateWidget dispatches to the type-specific update helper.
func (s *Session) UpdateWidget(ctx context.Context, canvasID, widgetID string, req map[string]any) (*Widget, error) {
	widgetType, ok := req["widget_type"].(string)
	if !ok {
		return nil, fmt.Errorf("UpdateWidget: missing or invalid widget_type in request")
	}
	switch strings.ToLower(widgetType) {
	case "note":
		n, err := s.UpdateNote(ctx, canvasID, widgetID, req)
		if err != nil {
			return nil, err
		}
		return widgetFromNote(n), nil
	case "anchor":
		a, err := s.UpdateAnchor(ctx, canvasID, widgetID, req)
		if err != nil {
			return nil, err
		}
		return widgetFromAnchor(a), nil
	case "browser":
		b, err := s.UpdateBrowser(ctx, canvasID, widgetID, req)
		if err != nil {
			return nil, err
		}
		return widgetFromBrowser(b), nil
	case "image":
		i, err := s.UpdateImage(ctx, canvasID, widgetID, req)
		if err != nil {
			return nil, err
		}
		return widgetFromImage(i), nil
	case "pdf":
		p, err := s.UpdatePDF(ctx, canvasID, widgetID, req)
		if err != nil {
			return nil, err
		}
		return &Widget{ID: p.ID, WidgetType: "pdf"}, nil
	case "video":
		v, err := s.UpdateVideo(ctx, canvasID, widgetID, req)
		if err != nil {
			return nil, err
		}
		return &Widget{ID: v.ID, WidgetType: "video"}, nil
	case "connector":
		c, err := s.UpdateConnector(ctx, canvasID, widgetID, req)
		if err != nil {
			return nil, err
		}
		return &Widget{ID: c.ID, WidgetType: "connector"}, nil
	case "table":
		t, err := s.UpdateTable(ctx, canvasID, widgetID, req)
		if err != nil {
			return nil, err
		}
		return widgetFromTable(t), nil
	case "ip_video", "ipvideo", "ip-video":
		v, err := s.UpdateIPVideo(ctx, canvasID, widgetID, req)
		if err != nil {
			return nil, err
		}
		return &Widget{ID: v.ID, WidgetType: "ip_video"}, nil
	case "rdp_connection", "rdpconnection", "rdp-connection":
		r, err := s.UpdateRDPConnection(ctx, canvasID, widgetID, req)
		if err != nil {
			return nil, err
		}
		return &Widget{ID: r.ID, WidgetType: "rdp_connection"}, nil
	default:
		return nil, fmt.Errorf("UpdateWidget: unsupported widget_type: %s", widgetType)
	}
}

// DeleteWidget dispatches deletion to the type-specific endpoint.
// widgetType must be supplied (not inferred from server response).
func (s *Session) DeleteWidget(ctx context.Context, canvasID, widgetID, widgetType string) error {
	switch strings.ToLower(widgetType) {
	case "note":
		return s.DeleteNote(ctx, canvasID, widgetID)
	case "anchor":
		return s.DeleteAnchor(ctx, canvasID, widgetID)
	case "browser":
		return s.DeleteBrowser(ctx, canvasID, widgetID)
	case "image":
		return s.DeleteImage(ctx, canvasID, widgetID)
	case "pdf":
		return s.DeletePDF(ctx, canvasID, widgetID)
	case "video":
		return s.DeleteVideo(ctx, canvasID, widgetID)
	case "connector":
		return s.DeleteConnector(ctx, canvasID, widgetID)
	case "table":
		return s.DeleteTable(ctx, canvasID, widgetID)
	case "ip_video", "ipvideo", "ip-video":
		return s.DeleteIPVideo(ctx, canvasID, widgetID)
	case "rdp_connection", "rdpconnection", "rdp-connection":
		return s.DeleteRDPConnection(ctx, canvasID, widgetID)
	}
	return fmt.Errorf("DeleteWidget: unsupported widget_type: %s", widgetType)
}

// PatchParentID updates the parent ID of a widget (re-parenting).
//
// Note: this uses the read-only /widgets/{id} path. The server accepts PATCH
// here as a generic shortcut even though it isn't documented in the public
// spec. Prefer the type-specific update helpers when possible.
func (s *Session) PatchParentID(ctx context.Context, canvasID, widgetID, parentID string) (*Widget, error) {
	var widget Widget
	req := map[string]any{"parent_id": parentID}
	if err := s.doRequest(ctx, http.MethodPatch, fmt.Sprintf("canvases/%s/widgets/%s", canvasID, widgetID), req, &widget, nil, false); err != nil {
		return nil, fmt.Errorf("PatchParentID: %w", err)
	}
	return &widget, nil
}

// CloneWidget performs a cross-canvas widget clone using the standard type-
// specific create endpoint, per changelog §1.
//
// The caller supplies the destination canvas, the source canvas+widget IDs,
// the widget's path segment (one of "notes", "images", "videos", "pdfs",
// "browsers", "anchors", "tables"), and an optional override location. The
// server clones the widget into the destination canvas, regenerating IDs and
// state.
//
// For asset-based types (images/videos/pdfs), the source asset must be
// available; this is checked server-side.
func (s *Session) CloneWidget(ctx context.Context, destCanvasID, sourceCanvasID, sourceWidgetID, widgetTypePath string, location *Point) (*Widget, error) {
	if destCanvasID == "" || sourceCanvasID == "" || sourceWidgetID == "" || widgetTypePath == "" {
		return nil, fmt.Errorf("CloneWidget: destCanvasID, sourceCanvasID, sourceWidgetID and widgetTypePath are required")
	}
	body := map[string]any{
		"source_canvas_id": sourceCanvasID,
		"source_widget_id": sourceWidgetID,
	}
	if location != nil {
		body["location"] = map[string]any{"x": location.X, "y": location.Y}
	}
	var widget Widget
	endpoint := fmt.Sprintf("canvases/%s/%s", destCanvasID, strings.ToLower(widgetTypePath))
	if err := s.doRequest(ctx, http.MethodPost, endpoint, body, &widget, nil, false); err != nil {
		return nil, fmt.Errorf("CloneWidget: %w", err)
	}
	return &widget, nil
}

// WidgetMatch represents a widget search hit across canvases.
type WidgetMatch struct {
	CanvasID string
	WidgetID string
	Widget   Widget
}

// WidgetsLister is the dependency contract for FindWidgetsAcrossCanvases.
type WidgetsLister interface {
	ListCanvases(ctx context.Context, filter *Filter) ([]Canvas, error)
	ListWidgets(ctx context.Context, canvasID string, filter *Filter, includeAnnotations ...bool) ([]Widget, error)
}

// FindWidgetsAcrossCanvases searches all canvases for widgets matching query.
func FindWidgetsAcrossCanvases(ctx context.Context, lister WidgetsLister, query map[string]any) ([]WidgetMatch, error) {
	canvases, err := lister.ListCanvases(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("FindWidgetsAcrossCanvases: failed to list canvases: %w", err)
	}
	filter := &Filter{Criteria: query}
	var matches []WidgetMatch
	for _, canvas := range canvases {
		widgets, err := lister.ListWidgets(ctx, canvas.ID, filter)
		if err != nil {
			return nil, fmt.Errorf("FindWidgetsAcrossCanvases: failed to list widgets for canvas %s: %w", canvas.ID, err)
		}
		for _, w := range widgets {
			matches = append(matches, WidgetMatch{CanvasID: canvas.ID, WidgetID: w.ID, Widget: w})
		}
	}
	return matches, nil
}

// WidgetZone bundles a container widget with the widgets fully inside it.
type WidgetZone struct {
	CanvasID       string
	SharedCanvasID string
	Container      Widget
	Contents       []Widget
}

// WidgetsContainId returns a WidgetZone with widget as Container and all
// widgets fully contained inside it as Contents.
func WidgetsContainId(ctx context.Context, s *Session, canvasID, widgetID string, widget *Widget, tolerance float64) (WidgetZone, error) {
	var srcWidget Widget
	if widget != nil {
		srcWidget = *widget
	} else {
		if widgetID == "" {
			return WidgetZone{}, fmt.Errorf("WidgetsContainId: widgetID must be provided if widget is nil")
		}
		w, err := s.GetWidget(ctx, canvasID, widgetID)
		if err != nil {
			return WidgetZone{}, fmt.Errorf("WidgetsContainId: failed to fetch widget: %w", err)
		}
		srcWidget = *w
	}

	widgets, err := s.ListWidgets(ctx, canvasID, nil)
	if err != nil {
		return WidgetZone{}, fmt.Errorf("WidgetsContainId: failed to list widgets: %w", err)
	}

	var sharedCanvasID string
	for _, w := range widgets {
		if w.WidgetType == "SharedCanvas" {
			sharedCanvasID = w.ID
			break
		}
	}

	srcRect := WidgetBoundingBox(srcWidget)
	srcRect.X -= tolerance
	srcRect.Y -= tolerance
	srcRect.Width += 2 * tolerance
	srcRect.Height += 2 * tolerance

	var contained []Widget
	for _, w := range widgets {
		if w.ID == srcWidget.ID || w.WidgetType == "SharedCanvas" {
			continue
		}
		if Contains(srcRect, WidgetBoundingBox(w)) {
			if sharedCanvasID != "" && w.ParentID == sharedCanvasID {
				w.ParentID = ""
			}
			contained = append(contained, w)
		}
	}
	if sharedCanvasID != "" && srcWidget.ParentID == sharedCanvasID {
		srcWidget.ParentID = ""
	}
	return WidgetZone{CanvasID: canvasID, SharedCanvasID: sharedCanvasID, Container: srcWidget, Contents: contained}, nil
}

// --- internal converters ---

func widgetFromNote(n *Note) *Widget {
	return &Widget{
		ID: n.ID, WidgetType: n.WidgetType, ParentID: n.ParentID,
		Location: n.Location, Size: n.Size, Pinned: n.Pinned,
		Scale: n.Scale, State: n.State, Depth: n.Depth,
	}
}
func widgetFromAnchor(a *Anchor) *Widget {
	return &Widget{
		ID: a.ID, WidgetType: a.WidgetType, ParentID: a.ParentID,
		Location: a.Location, Size: a.Size, Pinned: a.Pinned,
		Scale: a.Scale, State: a.State, Depth: a.Depth,
	}
}
func widgetFromBrowser(b *Browser) *Widget {
	return &Widget{
		ID: b.ID, WidgetType: b.WidgetType, ParentID: b.ParentID,
		Location: b.Location, Size: b.Size, Pinned: b.Pinned,
		Scale: b.Scale, State: b.State, Depth: b.Depth,
	}
}
func widgetFromImage(i *Image) *Widget {
	return &Widget{
		ID: i.ID, WidgetType: i.WidgetType, ParentID: i.ParentID,
		Location: i.Location, Size: i.Size, Pinned: i.Pinned,
		Scale: i.Scale, State: i.State, Depth: i.Depth,
	}
}
func widgetFromTable(t *Table) *Widget {
	return &Widget{
		ID: t.ID, WidgetType: t.WidgetType, ParentID: t.ParentID,
		Location: t.Location, Size: t.Size, Pinned: t.Pinned,
		Scale: t.Scale, State: t.State, Depth: t.Depth,
	}
}
