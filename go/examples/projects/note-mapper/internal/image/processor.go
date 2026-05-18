// Package image provides image preprocessing utilities for note-mapper.
//
// ProcessImage resizes and re-encodes images so they fit within the size and
// dimension limits imposed by the Google Gemini multimodal API.
package image

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log/slog"
	"strings"

	"github.com/nfnt/resize"
)

const (
	// MaxImageBytes is the upper size limit sent to the Gemini API (20 MB).
	MaxImageBytes = 20 * 1024 * 1024
	// MaxDimension is the maximum pixel extent on either axis.
	MaxDimension = 2048
	// jpegQuality is the initial JPEG encoding quality.
	jpegQuality = 85
)

// ProcessImage takes raw image bytes and returns optimized bytes along with
// the MIME type ("image/jpeg" or "image/png") suitable for passing to Gemini.
//
// Behaviour:
//   - Decodes the image (JPEG or PNG).
//   - Resizes proportionally if either dimension exceeds MaxDimension.
//   - Re-encodes as JPEG for JPEG inputs (with adaptive quality reduction if
//     the result still exceeds MaxImageBytes); PNG for all other formats.
//
// Returns an error only when the image cannot be decoded or encoded.
func ProcessImage(input []byte) ([]byte, string, error) {
	img, format, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return nil, "", fmt.Errorf("ProcessImage: decode: %w", err)
	}

	bounds := img.Bounds()
	origW, origH := bounds.Dx(), bounds.Dy()
	slog.Debug("image decoded", "format", format, "width", origW, "height", origH)

	// Resize if necessary.
	newW, newH := resizeDimensions(origW, origH)
	if newW != origW || newH != origH {
		slog.Debug("resizing image", "new_width", newW, "new_height", newH)
		img = resize.Resize(uint(newW), uint(newH), img, resize.Lanczos3)
	}

	// Encode.
	out, mimeType, err := encode(img, format, jpegQuality)
	if err != nil {
		return nil, "", fmt.Errorf("ProcessImage: encode: %w", err)
	}

	// If the JPEG result is still too large, reduce quality iteratively.
	if len(out) > MaxImageBytes && isJPEG(format) {
		q := jpegQuality
		for len(out) > MaxImageBytes && q > 10 {
			q -= 10
			slog.Debug("reducing JPEG quality", "quality", q)
			out, _, err = encode(img, format, q)
			if err != nil {
				return nil, "", fmt.Errorf("ProcessImage: re-encode at quality %d: %w", q, err)
			}
		}
	}

	slog.Debug("image processed", "bytes", len(out), "mime_type", mimeType)
	return out, mimeType, nil
}

// resizeDimensions returns new dimensions that fit within MaxDimension while
// preserving aspect ratio. Returns the originals if no resize is needed.
func resizeDimensions(w, h int) (int, int) {
	if w <= MaxDimension && h <= MaxDimension {
		return w, h
	}
	if w > h {
		return MaxDimension, int(float64(h) * float64(MaxDimension) / float64(w))
	}
	return int(float64(w) * float64(MaxDimension) / float64(h)), MaxDimension
}

// encode encodes img in the appropriate format.
func encode(img image.Image, format string, jpegQ int) ([]byte, string, error) {
	var buf bytes.Buffer
	if isJPEG(format) {
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: jpegQ}); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "image/jpeg", nil
	}
	if err := png.Encode(&buf, img); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), "image/png", nil
}

func isJPEG(format string) bool {
	f := strings.ToLower(format)
	return f == "jpeg" || f == "jpg"
}
