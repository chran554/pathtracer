package color

import (
	col "image/color"
)

var FloatNRGBAModel col.Model = floatNRGBAModel{}

func (c *Color) RGBA() (r, g, b, a uint32) {
	return uint32(clampFloat32(0, 1.0, c.R) * 0xffff),
		uint32(clampFloat32(0, 1.0, c.G) * 0xffff),
		uint32(clampFloat32(0, 1.0, c.B) * 0xffff),
		uint32(clampFloat32(0, 1.0, c.A) * 0xffff)
}

type floatNRGBAModel struct{}

func (m floatNRGBAModel) Convert(c col.Color) col.Color { return floatNRGBAModelConvert(c) }

func floatNRGBAModelConvert(c col.Color) col.Color {
	if _, ok := c.(*Color); ok {
		return c
	}
	r, g, b, a := c.RGBA()

	const invMax = 1.0 / float32(0xffff)

	// Special case for fully opaque pixels (alpha == 0xffff).
	if a == 0xffff {
		return &Color{
			R: float32(r) * invMax,
			G: float32(g) * invMax,
			B: float32(b) * invMax,
			A: 1.0,
		}
	}

	// Special case for fully transparent pixels (alpha == 0x0000).
	if a == 0x0000 {
		return &Color{R: 0, G: 0, B: 0, A: 0}
	}

	// Since Color.RGBA returns an alpha-premultiplied color, we should have r <= a && g <= a && b <= a.
	invA := 1.0 / float32(a)
	return &Color{
		R: float32(r) * invA,
		G: float32(g) * invA,
		B: float32(b) * invA,
		A: float32(a) * invMax,
	}
}

func clampFloat32(min float32, max float32, value float32) float32 {
	if value < min {
		return min
	} else if value > max {
		return max
	}

	return value
}
