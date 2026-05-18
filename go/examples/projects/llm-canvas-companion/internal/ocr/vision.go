// Package ocr performs text extraction via the Google Cloud Vision API.
//
// The single exported function PerformOCR sends image bytes to
// https://vision.googleapis.com/v1/images:annotate and returns the extracted
// text.  If GOOGLE_VISION_API_KEY is empty the call fails immediately with a
// clear error rather than making a network round-trip.
package ocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	visionEndpoint = "https://vision.googleapis.com/v1/images:annotate"
	// minimalPNG is a 1×1 transparent PNG used for API-key validation.
	minimalPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII="
)

type visionRequest struct {
	Requests []visionReqItem `json:"requests"`
}

type visionReqItem struct {
	Image    visionImage    `json:"image"`
	Features []visionFeature `json:"features"`
}

type visionImage struct {
	Content string `json:"content"`
}

type visionFeature struct {
	Type       string `json:"type"`
	MaxResults int    `json:"maxResults"`
}

type visionResponse struct {
	Responses []visionRespItem `json:"responses"`
}

type visionRespItem struct {
	FullTextAnnotation struct {
		Text string `json:"text"`
	} `json:"fullTextAnnotation"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// PerformOCR extracts text from imageData using Google Cloud Vision.
// apiKey must be a valid Cloud Vision API key. Returns an error if no text
// is found or the API call fails.
func PerformOCR(ctx context.Context, apiKey string, imageData []byte) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("PerformOCR: GOOGLE_VISION_API_KEY is not set")
	}

	body := visionRequest{
		Requests: []visionReqItem{
			{
				Image: visionImage{Content: base64.StdEncoding.EncodeToString(imageData)},
				Features: []visionFeature{
					{Type: "DOCUMENT_TEXT_DETECTION", MaxResults: 1},
				},
			},
		},
	}

	text, err := callVision(ctx, apiKey, body)
	if err != nil {
		return "", fmt.Errorf("PerformOCR: %w", err)
	}
	if text == "" {
		return "", fmt.Errorf("PerformOCR: no text found in image")
	}
	return text, nil
}

// ValidateAPIKey makes a minimal Vision API call to confirm the key is active.
func ValidateAPIKey(ctx context.Context, apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("ValidateAPIKey: key is empty")
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	body := visionRequest{
		Requests: []visionReqItem{
			{
				Image: visionImage{Content: minimalPNG},
				Features: []visionFeature{
					{Type: "TEXT_DETECTION", MaxResults: 1},
				},
			},
		},
	}
	_, err := callVision(ctx, apiKey, body)
	// A 200 response (even with empty text) means the key is valid.
	if err != nil {
		return fmt.Errorf("ValidateAPIKey: %w", err)
	}
	return nil
}

// callVision posts a Vision API request and returns the first page's text.
func callVision(ctx context.Context, apiKey string, body visionRequest) (string, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", visionEndpoint, apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("POST vision: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBytes, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("vision API HTTP %d: %s", resp.StatusCode, respBytes)
	}

	var vr visionResponse
	if err := json.Unmarshal(respBytes, &vr); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(vr.Responses) == 0 {
		return "", fmt.Errorf("empty responses array")
	}
	if vr.Responses[0].Error.Message != "" {
		return "", fmt.Errorf("vision error: %s", vr.Responses[0].Error.Message)
	}
	return vr.Responses[0].FullTextAnnotation.Text, nil
}
