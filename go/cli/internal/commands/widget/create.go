package widget

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/output"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <canvas-id>",
	Short: "Create a new widget",
	Long: `Create a new widget in a canvas.

Widgets are content elements that can be added to canvases. Use the --type flag
to specify the widget type and provide type-specific parameters.

Supported widget types:
  - note: Text notes with optional title and background color
  - image: Image widgets from URL or file
  - pdf: PDF document widgets
  - video: Video widgets
  - anchor: Connection point widgets
  - connector: Line connectors between widgets

Examples:
  # Create a note widget
  canvus widget create canvas-123 --type note --text "My note" --x 100 --y 200

  # Create a note with title and background color
  canvus widget create canvas-123 --type note --text "Content" --title "Title" --background-color "#FF0000"

  # Create an image widget from URL
  canvus widget create canvas-123 --type image --url "https://example.com/image.png" --x 300 --y 400

  # Create an image widget from file
  canvus widget create canvas-123 --type image --file "/path/to/image.png" --x 300 --y 400

  # Create a PDF widget
  canvus widget create canvas-123 --type pdf --url "https://example.com/doc.pdf" --title "Document"

  # Create a video widget
  canvus widget create canvas-123 --type video --url "https://example.com/video.mp4" --title "Video"`,
	Args: cobra.ExactArgs(1),
	RunE: runCreate,
}

var (
	createType            string
	createText            string
	createTitle           string
	createBackgroundColor string
	createURL             string
	createFile            string
	createX               float64
	createY               float64
	createWidth           float64
	createHeight          float64
)

func init() {
	createCmd.Flags().StringVar(&createType, "type", "", "Widget type (required): note, image, pdf, video, anchor, connector")
	createCmd.Flags().StringVar(&createText, "text", "", "Text content (for note widgets)")
	createCmd.Flags().StringVar(&createTitle, "title", "", "Title (for note, pdf, video widgets)")
	createCmd.Flags().StringVar(&createBackgroundColor, "background-color", "", "Background color (for note widgets)")
	createCmd.Flags().StringVar(&createURL, "url", "", "URL (for image, pdf, video widgets)")
	createCmd.Flags().StringVar(&createFile, "file", "", "File path (for image widgets)")
	createCmd.Flags().Float64Var(&createX, "x", 0, "X coordinate")
	createCmd.Flags().Float64Var(&createY, "y", 0, "Y coordinate")
	createCmd.Flags().Float64Var(&createWidth, "width", 0, "Width")
	createCmd.Flags().Float64Var(&createHeight, "height", 0, "Height")

	createCmd.MarkFlagRequired("type")
	WidgetCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	canvasID := args[0]

	// Validate canvas ID
	if err := validateCanvasID(canvasID); err != nil {
		return err
	}

	// Validate widget type
	if err := validateWidgetType(createType); err != nil {
		return err
	}

	// Get SDK session from context
	sess, err := session.GetSession(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	ctx := context.Background()

	// Create widget based on type
	var result interface{}
	switch createType {
	case "note":
		result, err = createNoteWidget(ctx, sess, canvasID)
	case "image":
		result, err = createImageWidget(ctx, sess, canvasID)
	case "pdf":
		result, err = createPDFWidget(ctx, sess, canvasID)
	case "video":
		result, err = createVideoWidget(ctx, sess, canvasID)
	case "anchor":
		result, err = createAnchorWidget(ctx, sess, canvasID)
	case "connector":
		result, err = createConnectorWidget(ctx, sess, canvasID)
	default:
		return fmt.Errorf("unsupported widget type: %s", createType)
	}

	if err != nil {
		return fmt.Errorf("failed to create %s widget: %w", createType, err)
	}

	// Format and output the result
	return output.OutputSingle(cmd.Context(), result, "")
}

func createNoteWidget(ctx context.Context, sess *canvus.Session, canvasID string) (*canvus.Note, error) {
	if createText == "" {
		return nil, fmt.Errorf("--text is required for note widgets")
	}

	req := map[string]interface{}{
		"text": createText,
	}

	if createTitle != "" {
		req["title"] = createTitle
	}
	if createBackgroundColor != "" {
		req["background_color"] = createBackgroundColor
	}

	// Add location if specified
	if createX != 0 || createY != 0 {
		req["location"] = map[string]interface{}{
			"x": createX,
			"y": createY,
		}
	}

	// Add size if specified
	if createWidth != 0 || createHeight != 0 {
		req["size"] = map[string]interface{}{
			"width":  createWidth,
			"height": createHeight,
		}
	}

	return sess.CreateNote(ctx, canvasID, req)
}

func createImageWidget(ctx context.Context, sess *canvus.Session, canvasID string) (*canvus.Image, error) {
	if createURL == "" && createFile == "" {
		return nil, fmt.Errorf("either --url or --file is required for image widgets")
	}
	if createURL != "" && createFile != "" {
		return nil, fmt.Errorf("cannot specify both --url and --file")
	}

	var filename string
	var fileReader *bytes.Reader

	if createFile != "" {
		// Read file content
		data, err := os.ReadFile(createFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		fileReader = bytes.NewReader(data)
		filename = createFile
	}

	// For URL-based images, use the URL as the identifier
	if createURL != "" {
		filename = createURL
	}

	return sess.CreateImage(ctx, canvasID, fileReader, filename)
}

func createPDFWidget(ctx context.Context, sess *canvus.Session, canvasID string) (*canvus.PDF, error) {
	if createURL == "" {
		return nil, fmt.Errorf("--url is required for PDF widgets")
	}

	req := map[string]interface{}{
		"url": createURL,
	}

	title := createTitle
	if title == "" {
		title = "PDF Document"
	}

	return sess.CreatePDF(ctx, canvasID, req, title)
}

func createVideoWidget(ctx context.Context, sess *canvus.Session, canvasID string) (*canvus.Video, error) {
	if createURL == "" {
		return nil, fmt.Errorf("--url is required for video widgets")
	}

	req := map[string]interface{}{
		"url": createURL,
	}

	title := createTitle
	if title == "" {
		title = "Video"
	}

	return sess.CreateVideo(ctx, canvasID, req, title)
}

func createAnchorWidget(ctx context.Context, sess *canvus.Session, canvasID string) (*canvus.Anchor, error) {
	req := map[string]interface{}{}

	// Add location if specified
	if createX != 0 || createY != 0 {
		req["location"] = map[string]interface{}{
			"x": createX,
			"y": createY,
		}
	}

	return sess.CreateAnchor(ctx, canvasID, req)
}

func createConnectorWidget(ctx context.Context, sess *canvus.Session, canvasID string) (*canvus.Connector, error) {
	req := map[string]interface{}{}

	// Connectors typically need start and end points or connected widget IDs
	// For simplicity, we'll just create a basic connector

	return sess.CreateConnector(ctx, canvasID, req)
}
