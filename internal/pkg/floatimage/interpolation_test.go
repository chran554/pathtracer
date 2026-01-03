package floatimage

import (
	"pathtracer/internal/pkg/color"
	"testing"

	"os"
	"path/filepath"

	"github.com/stretchr/testify/assert"
)

func TestInterpolation_Save(t *testing.T) {
	// 1. Create a 5x5 test image with linear color values.
	// Interpolation should be performed in linear space for physically correct results.
	// The WriteImage function will handle gamma encoding when saving the final image.
	srcSize := 5
	src := NewFloatImage("source", srcSize, srcSize)
	for y := 0; y < srcSize; y++ {
		for x := 0; x < srcSize; x++ {
			// Create a simple pattern
			r := float64(x) / float64(srcSize-1)
			g := float64(y) / float64(srcSize-1)
			b := 1.0 - (r+g)/2.0
			src.SetPixel(x, y, color.NewColor(r, g, b))
		}
	}
	// Add a dot in the middle
	src.SetPixel(2, 2, color.NewColor(1, 1, 1))

	dstSize := 50
	tesCases := []struct {
		name          string
		interpolation Interpolation
		sampleOffset  float64
		file          string
	}{
		{"NearestNeighbor", InterpolationNearestNeighbor, 0.0, "interpolation_nearest.png"},
		{"Bilinear", InterpolationBilinear, -0.5, "interpolation_bilinear.png"},
		{"Bicubic", InterpolationBicubic, -0.5, "interpolation_bicubic.png"},
	}

	renderDir := "../../../rendered"
	_ = os.MkdirAll(renderDir, 0755)

	for _, testCase := range tesCases {
		t.Run(testCase.name, func(t *testing.T) {
			dst := NewFloatImage(testCase.name, dstSize, dstSize)
			scale := float64(srcSize+1) / float64(dstSize-1)

			for y := 0; y < dstSize; y++ {
				// fmt.Printf("pos: %0.3f\n", float64(y)*scale-0.5)

				for x := 0; x < dstSize; x++ {
					srcX := float64(x)*scale - 0.5
					srcY := float64(y)*scale - 0.5
					srcX += testCase.sampleOffset
					srcY += testCase.sampleOffset

					c := src.GetInterpolatedPixel(srcX, srcY, testCase.interpolation)
					dst.SetPixel(x, y, c)
				}
			}

			outputPath := filepath.Join(renderDir, testCase.file)
			WriteImage(outputPath, dst)

			_, err := os.Stat(outputPath)
			assert.NoError(t, err, "Output file should exist: %s", outputPath)
		})
	}
}

func TestInterpolation(t *testing.T) {
	fi := NewFloatImage("test", 4, 4)
	// Fill with some pattern
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			fi.SetPixel(x, y, color.NewColor(float64(x)/3.0, float64(y)/3.0, 0))
		}
	}

	t.Run("Bilinear", func(t *testing.T) {
		c := fi.GetInterpolatedPixel(1.5, 1.5, InterpolationBilinear)
		assert.InDelta(t, 0.5, c.R, 0.001)
		assert.InDelta(t, 0.5, c.G, 0.001)
	})

	t.Run("Bicubic", func(t *testing.T) {
		c := fi.GetInterpolatedPixel(1.5, 1.5, InterpolationBicubic)
		// For a linear gradient, bicubic should also return the midpoint if it behaves well
		assert.InDelta(t, 0.5, c.R, 0.001)
		assert.InDelta(t, 0.5, c.G, 0.001)
	})

	t.Run("BicubicBoundary", func(t *testing.T) {
		// Test near boundary to ensure no panic and reasonable values
		c := fi.GetInterpolatedPixel(0.1, 0.1, InterpolationBicubic)
		assert.NotNil(t, c)

		// Test near boundary to ensure no panic and reasonable values
		c = fi.GetInterpolatedPixel(0.0, 0.0, InterpolationBicubic)
		assert.NotNil(t, c)

		// Test near boundary to ensure no panic and reasonable values
		c = fi.GetInterpolatedPixel(float64(fi.Bounds().Dx()-1), float64(fi.Bounds().Dy()-1), InterpolationBicubic)
		assert.NotNil(t, c)

		// Test near boundary to ensure no panic and reasonable values
		c = fi.GetInterpolatedPixel(float64(fi.Bounds().Dx()), float64(fi.Bounds().Dy()), InterpolationBicubic)
		assert.NotNil(t, c)
	})
}
