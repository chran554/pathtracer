package floatimage

import (
	"math"
	"pathtracer/internal/pkg/color"
)

type Interpolation int

const (
	InterpolationNearestNeighbor = iota
	InterpolationBilinear
)

func (fi *FloatImage) GetInterpolatedPixel(srcX, srcY float64, interpolation Interpolation) *color.Color {
	switch interpolation {
	case InterpolationBilinear:
		return fi.GetBilinearInterpolatedPixel(srcX, srcY)
	case InterpolationNearestNeighbor:
		fallthrough
	default:
		return fi.GetPixel(int(srcX), int(srcY))
	}
}

func (fi *FloatImage) GetBilinearInterpolatedPixel(srcX, srcY float64) *color.Color {
	x0, y0 := int(math.Floor(srcX)), int(math.Floor(srcY))
	x1, y1 := x0+1, y0+1

	// Ensure coordinates stay within image bounds
	bounds := fi.Bounds()
	x0 = clampInt(x0, bounds.Min.X, bounds.Max.X-1)
	y0 = clampInt(y0, bounds.Min.Y, bounds.Max.Y-1)
	x1 = clampInt(x1, bounds.Min.X, bounds.Max.X-1)
	y1 = clampInt(y1, bounds.Min.Y, bounds.Max.Y-1)

	// Calculate fractional weights
	fx := float32(srcX) - float32(x0)
	fy := float32(srcY) - float32(y0)
	fx1, fy1 := 1-fx, 1-fy

	// Fetch the color values from neighboring pixels
	c00 := fi.GetPixel(x0, y0)
	c10 := fi.GetPixel(x1, y0)
	c01 := fi.GetPixel(x0, y1)
	c11 := fi.GetPixel(x1, y1)

	// Compute the weighted sum of the RGB channels
	r := fx1*fy1*c00.R +
		fx*fy1*c10.R +
		fx1*fy*c01.R +
		fx*fy*c11.R

	g := fx1*fy1*c00.G +
		fx*fy1*c10.G +
		fx1*fy*c01.G +
		fx*fy*c11.G

	b := fx1*fy1*c00.B +
		fx*fy1*c10.B +
		fx1*fy*c01.B +
		fx*fy*c11.B

	a := fx1*fy1*c00.A +
		fx*fy1*c10.A +
		fx1*fy*c01.A +
		fx*fy*c11.A

	// Return the interpolated color
	return &color.Color{R: r, G: g, B: b, A: a}
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
