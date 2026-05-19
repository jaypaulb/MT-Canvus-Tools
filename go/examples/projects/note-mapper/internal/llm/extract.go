package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"

	sharedllm "github.com/jaypaulb/MT-Canvus-Tools/go/internal/llm"
)

const (
	geminiModel   = "gemini-2.5-flash-preview-05-20"
	geminiTimeout = 5 * time.Minute
	geminiEnvKey  = "GOOGLE_GENAI_API_KEY"
)

// ExtractPostitNotes calls Google Gemini to identify Post-it notes in the
// supplied image and returns a slice of Note values with pixel-space positions
// relative to the image's top-left corner.
//
// The GOOGLE_GENAI_API_KEY environment variable must be set.
func ExtractPostitNotes(input ExtractInput) ([]Note, error) {
	apiKey := os.Getenv(geminiEnvKey)
	if apiKey == "" {
		return nil, fmt.Errorf("ExtractPostitNotes: %s environment variable is not set", geminiEnvKey)
	}

	ctx, cancel := context.WithTimeout(context.Background(), geminiTimeout)
	defer cancel()

	shared, err := sharedllm.NewClient(ctx, sharedllm.Config{
		APIKey: apiKey,
		Model:  geminiModel,
	})
	if err != nil {
		return nil, fmt.Errorf("ExtractPostitNotes: create Gemini client: %w", err)
	}
	defer shared.Close()
	client := shared.Raw()

	cfg := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   noteSchema(),
	}

	// Build prompt contents.
	contents := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: extractionPrompt},
				genai.NewPartFromBytes(input.ImageData, input.MimeType),
			},
		},
	}

	slog.Debug("calling Gemini", "model", geminiModel, "image_bytes", len(input.ImageData), "mime", input.MimeType)

	resp, err := client.Models.GenerateContent(ctx, geminiModel, contents, cfg)
	if err != nil {
		return nil, fmt.Errorf("ExtractPostitNotes: GenerateContent: %w", err)
	}

	raw, err := parseGeminiResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("ExtractPostitNotes: parse response: %w", err)
	}

	slog.Info("Gemini extraction complete", "note_count", len(raw))
	return convertOutputs(raw), nil
}

// noteSchema returns the JSON schema constraint passed to Gemini for structured
// output. The schema mirrors extractOutput.
func noteSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeArray,
		Items: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"background_color": {Type: genai.TypeString},
				"location": {
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"x": {Type: genai.TypeInteger},
						"y": {Type: genai.TypeInteger},
					},
					Required: []string{"x", "y"},
				},
				"scale": {Type: genai.TypeNumber},
				"size": {
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"height": {Type: genai.TypeInteger},
						"width":  {Type: genai.TypeInteger},
					},
					Required: []string{"height", "width"},
				},
				"text":        {Type: genai.TypeString},
				"widget_type": {Type: genai.TypeString},
			},
			Required: []string{
				"background_color", "location", "scale", "size", "text", "widget_type",
			},
		},
	}
}

const extractionPrompt = `Analyze the image for Post-it notes. For each note, extract:
- The text content.
- The background color as a hex code (e.g. "#FFEB3B").
- The precise top-left pixel location (x, y) relative to the image top-left corner.
- The width and height in pixels.
- The scale (use 1.0 unless the note is visibly scaled).

Relative spatial positioning and sizing matter: a note that appears in the top-left
of the image should have a lower (x, y) than a note in the center.

Return a JSON array. Each element must match the schema exactly.`

// parseGeminiResponse extracts the JSON array of notes from the Gemini response.
func parseGeminiResponse(resp *genai.GenerateContentResponse) ([]extractOutput, error) {
	for _, c := range resp.Candidates {
		if c.Content == nil {
			continue
		}
		for _, part := range c.Content.Parts {
			if part.Text == "" {
				continue
			}
			jsonStr := stripMarkdownFence(part.Text)
			var outputs []extractOutput
			if err := json.Unmarshal([]byte(jsonStr), &outputs); err == nil && len(outputs) > 0 {
				return outputs, nil
			}
		}
	}
	return nil, errors.New("no valid note array found in Gemini response")
}

// stripMarkdownFence removes optional ```json … ``` or ``` … ``` wrappers.
func stripMarkdownFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

// convertOutputs converts raw Gemini output to the canonical Note type.
func convertOutputs(raw []extractOutput) []Note {
	notes := make([]Note, 0, len(raw))
	for _, o := range raw {
		n := Note{
			Content: o.Text,
			Color:   o.BackgroundColor,
			Scale:   o.Scale,
		}
		if loc := o.Location; loc != nil {
			n.X = loc["x"]
			n.Y = loc["y"]
		}
		if sz := o.Size; sz != nil {
			n.Width = sz["width"]
			n.Height = sz["height"]
		}
		notes = append(notes, n)
	}
	return notes
}
