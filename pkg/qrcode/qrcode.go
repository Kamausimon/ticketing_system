package qrcode

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/skip2/go-qrcode"
)

// Generator handles QR code generation
type Generator struct {
	size    int
	level   qrcode.RecoveryLevel
	fgColor color.Color
	bgColor color.Color
}

// NewGenerator creates a new QR code generator
func NewGenerator() *Generator {
	return &Generator{
		size:    256,
		level:   qrcode.Medium,
		fgColor: color.Black,
		bgColor: color.White,
	}
}

// WithSize sets the QR code size
func (g *Generator) WithSize(size int) *Generator {
	g.size = size
	return g
}

// WithRecoveryLevel sets the error correction level
func (g *Generator) WithRecoveryLevel(level qrcode.RecoveryLevel) *Generator {
	g.level = level
	return g
}

// WithColors sets custom foreground and background colors
func (g *Generator) WithColors(fg, bg color.Color) *Generator {
	g.fgColor = fg
	g.bgColor = bg
	return g
}

// GenerateBytes generates QR code and returns as PNG bytes
func (g *Generator) GenerateBytes(content string) ([]byte, error) {
	qr, err := qrcode.New(content, g.level)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	qr.ForegroundColor = g.fgColor
	qr.BackgroundColor = g.bgColor

	return qr.PNG(g.size)
}

// GenerateImage generates QR code and returns as image.Image
func (g *Generator) GenerateImage(content string) (image.Image, error) {
	qr, err := qrcode.New(content, g.level)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	qr.ForegroundColor = g.fgColor
	qr.BackgroundColor = g.bgColor

	return qr.Image(g.size), nil
}

// GenerateFile generates QR code and saves to file
func (g *Generator) GenerateFile(content, filename string) error {
	qr, err := qrcode.New(content, g.level)
	if err != nil {
		return fmt.Errorf("failed to create QR code: %w", err)
	}

	qr.ForegroundColor = g.fgColor
	qr.BackgroundColor = g.bgColor

	return qr.WriteFile(g.size, filename)
}

// GenerateWithLogo generates QR code with a logo in the center
func (g *Generator) GenerateWithLogo(content string, logoPath string) ([]byte, error) {
	// Generate base QR code
	qrImg, err := g.GenerateImage(content)
	if err != nil {
		return nil, err
	}

	// Load logo
	logoFile, err := os.Open(logoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open logo: %w", err)
	}
	defer logoFile.Close()

	logo, _, err := image.Decode(logoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode logo: %w", err)
	}

	// Create new image with logo overlay
	bounds := qrImg.Bounds()
	result := image.NewRGBA(bounds)

	// Draw QR code
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			result.Set(x, y, qrImg.At(x, y))
		}
	}

	// Calculate logo position (center)
	logoSize := bounds.Dx() / 5 // Logo is 1/5 of QR code size
	logoBounds := logo.Bounds()
	startX := (bounds.Dx() - logoSize) / 2
	startY := (bounds.Dy() - logoSize) / 2

	// Draw logo
	for y := 0; y < logoSize; y++ {
		for x := 0; x < logoSize; x++ {
			srcX := logoBounds.Min.X + (x * logoBounds.Dx() / logoSize)
			srcY := logoBounds.Min.Y + (y * logoBounds.Dy() / logoSize)
			result.Set(startX+x, startY+y, logo.At(srcX, srcY))
		}
	}

	// Convert to PNG bytes
	var buf bytes.Buffer
	if err := png.Encode(&buf, result); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// Quick generation functions for convenience

// Generate creates a QR code with default settings
func Generate(content string) ([]byte, error) {
	return NewGenerator().GenerateBytes(content)
}

// GenerateToFile creates a QR code and saves to file with default settings
func GenerateToFile(content, filename string) error {
	return NewGenerator().GenerateFile(content, filename)
}

// GenerateCustom creates a QR code with custom size
func GenerateCustom(content string, size int) ([]byte, error) {
	return NewGenerator().WithSize(size).GenerateBytes(content)
}
