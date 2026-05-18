// Package canvasx contains Canvus-side helpers that compose on top of the SDK:
// multipart upload bodies, connector payloads, persona colors, and the small
// grid math used to lay out the persona Q&A canvas.
package canvasx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
)

// BuildImageMultipart constructs a multipart body matching the Canvus image
// upload contract: one "json" field carrying the metadata, one "data" file
// part carrying the image bytes.
func BuildImageMultipart(meta map[string]any, filename string, fileData []byte) (io.Reader, string, error) {
	jsonBytes, err := json.Marshal(meta)
	if err != nil {
		return nil, "", fmt.Errorf("BuildImageMultipart: encode meta: %w", err)
	}
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	jsonPart, err := mw.CreateFormField("json")
	if err != nil {
		return nil, "", fmt.Errorf("BuildImageMultipart: create json field: %w", err)
	}
	if _, err := jsonPart.Write(jsonBytes); err != nil {
		return nil, "", fmt.Errorf("BuildImageMultipart: write json field: %w", err)
	}

	filePart, err := mw.CreateFormFile("data", filename)
	if err != nil {
		return nil, "", fmt.Errorf("BuildImageMultipart: create data field: %w", err)
	}
	if _, err := filePart.Write(fileData); err != nil {
		return nil, "", fmt.Errorf("BuildImageMultipart: write data field: %w", err)
	}

	if err := mw.Close(); err != nil {
		return nil, "", fmt.Errorf("BuildImageMultipart: close: %w", err)
	}
	return &buf, mw.FormDataContentType(), nil
}
