package floatimage

import (
	"math"
	"pathtracer/internal/pkg/color"
)

type Interpolation string

const (
	InterpolationNearestNeighbor = "nearest neighbour"
	InterpolationBilinear        = "bilinear"
	InterpolationBicubic         = "bicubic"
)

func (fi *FloatImage) GetInterpolatedPixelByNormalizedCoordinates(srcX, srcY float64, interpolation Interpolation) *color.Color {
	imageBounds := fi.Bounds()
	pixelX := srcX*float64((imageBounds.Max.X-1)-imageBounds.Min.X) + float64(imageBounds.Min.X)
	pixelY := (1-srcY)*float64((imageBounds.Max.Y-1)-imageBounds.Min.Y) + float64(imageBounds.Min.Y)
	return fi.GetInterpolatedPixel(pixelX, pixelY, interpolation)
}

func (fi *FloatImage) GetInterpolatedPixel(srcX, srcY float64, interpolation Interpolation) *color.Color {
	switch interpolation {
	case InterpolationBilinear:
		return fi.GetBilinearInterpolatedPixel(srcX, srcY)
	case InterpolationBicubic:
		return fi.GetBicubicInterpolatedPixel(srcX, srcY)
	case InterpolationNearestNeighbor:
		fallthrough
	default:
		x := clampInt(int(srcX), fi.Bounds().Min.X, fi.Bounds().Max.X-1)
		y := clampInt(int(srcY), fi.Bounds().Min.Y, fi.Bounds().Max.Y-1)
		return fi.GetPixel(x, y)
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

func (fi *FloatImage) GetBicubicInterpolatedPixel(srcX, srcY float64) *color.Color {
	x0 := int(math.Floor(srcX))
	y0 := int(math.Floor(srcY))

	// Fractional part
	tx := srcX - float64(x0)
	ty := srcY - float64(y0)

	// Cubic weights
	wx := cubicWeights(tx)
	wy := cubicWeights(ty)

	var interpolatedR, interpolatedG, interpolatedB, interpolatedA float32

	bounds := fi.Bounds()

	for offsetY := -1; offsetY <= 2; offsetY++ {
		y := clampInt(y0+offsetY, bounds.Min.Y, bounds.Max.Y-1)
		rowWeight := float32(wy[offsetY+1])

		for offsetX := -1; offsetX <= 2; offsetX++ {
			x := clampInt(x0+offsetX, bounds.Min.X, bounds.Max.X-1)
			pixelWeight := rowWeight * float32(wx[offsetX+1])

			c := fi.GetPixel(x, y)
			interpolatedR += c.R * pixelWeight
			interpolatedG += c.G * pixelWeight
			interpolatedB += c.B * pixelWeight
			interpolatedA += c.A * pixelWeight
		}
	}

	return &color.Color{R: interpolatedR, G: interpolatedG, B: interpolatedB, A: interpolatedA}
}

func cubicWeights(t float64) [4]float64 {
	t2 := t * t
	t3 := t2 * t
	return [4]float64{
		0.5 * (-t3 + 2*t2 - t),
		0.5 * (3*t3 - 5*t2 + 2),
		0.5 * (-3*t3 + 4*t2 + t),
		0.5 * (t3 - t2),
	}
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
