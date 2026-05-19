package custommenu

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/atoms/logger"
)

// IconConverter handles icon conversion to 937x937 PNG format.
type IconConverter struct {
	targetSize   int
	contentSize  int
	transparent  bool
}

// NewIconConverter creates a new icon converter.
func NewIconConverter() *IconConverter {
	return &IconConverter{
		targetSize:  937,  // Target canvas size
		contentSize: 250,  // Recommended content area
		transparent: true, // Use transparent background
	}
}

// ConvertIcon converts an image to 937x937 PNG with centered content.
func (c *IconConverter) ConvertIcon(inputPath, outputPath string) error {
	// Open input file
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	// Decode image
	srcImg, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	logger.Logf("Converting %s (format: %s) to 937x937 PNG", filepath.Base(inputPath), format)

	// Create 937x937 canvas with transparent background
	dst := image.NewRGBA(image.Rect(0, 0, c.targetSize, c.targetSize))

	// Fill with transparent background (already transparent by default in RGBA)
	// No need to fill if we want transparency

	// Calculate scaling to fit content within ~250x250 center area
	srcBounds := srcImg.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	// Scale to fit within contentSize
	scale := float64(c.contentSize) / float64(max(srcWidth, srcHeight))
	newWidth := int(float64(srcWidth) * scale)
	newHeight := int(float64(srcHeight) * scale)

	// Calculate position to center the content
	offsetX := (c.targetSize - newWidth) / 2
	offsetY := (c.targetSize - newHeight) / 2

	// Resize and draw the image centered
	dstRect := image.Rect(offsetX, offsetY, offsetX+newWidth, offsetY+newHeight)

	// Use simple nearest-neighbor scaling (good enough for icons)
	scaled := resize(srcImg, newWidth, newHeight)
	draw.Draw(dst, dstRect, scaled, image.Point{}, draw.Over)

	// Create output directory if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save as PNG
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, dst); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	logger.Logf("Successfully converted icon to %s", outputPath)
	return nil
}

// resize performs simple nearest-neighbor image resizing.
func resize(src image.Image, width, height int) image.Image {
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	xRatio := float64(srcWidth) / float64(width)
	yRatio := float64(srcHeight) / float64(height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := int(float64(x) * xRatio)
			srcY := int(float64(y) * yRatio)
			dst.Set(x, y, src.At(srcX+srcBounds.Min.X, srcY+srcBounds.Min.Y))
		}
	}

	return dst
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ConvertIconWithOptions converts an icon with custom options.
func (c *IconConverter) ConvertIconWithOptions(inputPath, outputPath string, targetSize, contentSize int) error {
	c.targetSize = targetSize
	c.contentSize = contentSize
	return c.ConvertIcon(inputPath, outputPath)
}
