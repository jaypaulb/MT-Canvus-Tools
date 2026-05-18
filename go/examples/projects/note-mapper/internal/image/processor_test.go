package image

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

// makeTestJPEG creates a small solid-color JPEG for testing.
func makeTestJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("makeTestJPEG: %v", err)
	}
	return buf.Bytes()
}

// makeTestPNG creates a small solid-color PNG for testing.
func makeTestPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{R: 50, G: 100, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("makeTestPNG: %v", err)
	}
	return buf.Bytes()
}

func TestProcessImage_smallJPEG(t *testing.T) {
	input := makeTestJPEG(t, 100, 100)
	out, mime, err := ProcessImage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mime != "image/jpeg" {
		t.Errorf("mime: got %q, want image/jpeg", mime)
	}
	if len(out) == 0 {
		t.Error("output is empty")
	}
}

func TestProcessImage_smallPNG(t *testing.T) {
	input := makeTestPNG(t, 200, 150)
	out, mime, err := ProcessImage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mime != "image/png" {
		t.Errorf("mime: got %q, want image/png", mime)
	}
	if len(out) == 0 {
		t.Error("output is empty")
	}
}

func TestProcessImage_largeJPEGResized(t *testing.T) {
	// Image larger than MaxDimension (2048); expect resize.
	input := makeTestJPEG(t, 4000, 3000)
	out, mime, err := ProcessImage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mime != "image/jpeg" {
		t.Errorf("mime: got %q, want image/jpeg", mime)
	}
	// Decode result to verify dimensions are within MaxDimension.
	img, _, err := image.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("decode result: %v", err)
	}
	b := img.Bounds()
	if b.Dx() > MaxDimension || b.Dy() > MaxDimension {
		t.Errorf("result dimensions %dx%d exceed MaxDimension %d",
			b.Dx(), b.Dy(), MaxDimension)
	}
}

func TestProcessImage_invalidInput(t *testing.T) {
	_, _, err := ProcessImage([]byte("not an image"))
	if err == nil {
		t.Error("expected error for invalid input, got nil")
	}
}

func TestResizeDimensions_noResizeNeeded(t *testing.T) {
	w, h := resizeDimensions(100, 200)
	if w != 100 || h != 200 {
		t.Errorf("expected 100x200, got %dx%d", w, h)
	}
}

func TestResizeDimensions_landscapeResize(t *testing.T) {
	// 4000×2000 — wider; should cap at MaxDimension=2048 on the long axis.
	w, h := resizeDimensions(4000, 2000)
	if w != MaxDimension {
		t.Errorf("width: got %d, want %d", w, MaxDimension)
	}
	// Height should be proportional: 2000 * (2048/4000) = 1024.
	if h != 1024 {
		t.Errorf("height: got %d, want 1024", h)
	}
}

func TestResizeDimensions_portraitResize(t *testing.T) {
	// 1000×4000 — taller; cap height at MaxDimension.
	w, h := resizeDimensions(1000, 4000)
	if h != MaxDimension {
		t.Errorf("height: got %d, want %d", h, MaxDimension)
	}
	// Width should be proportional: 1000 * (2048/4000) = 512.
	if w != 512 {
		t.Errorf("width: got %d, want 512", w)
	}
}
