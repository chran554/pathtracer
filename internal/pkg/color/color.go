package color

import (
	"math"
	"pathtracer/internal/pkg/util"
)

var (
	Black = Color{R: 0, G: 0, B: 0}
	White = Color{R: 1, G: 1, B: 1}
)

type Color struct{ R, G, B float32 }

func NewColor(r, g, b float64) Color {
	return Color{R: float32(r), G: float32(g), B: float32(b)}
}

func NewColorGrey(greyIntensity float64) Color {
	return NewColor(greyIntensity, greyIntensity, greyIntensity)
}

func NewColorKelvin(colorTemperature float64) Color {
	return KelvinTemperatureColor(colorTemperature)
}

// KelvinTemperatureColor gives the rgb color for a planck black body radiator heated to the given temperature in Kelvin.
// This function is valid for Kelvin temperatures in the range [1000,40000]. Other values are clamped to the valid range.
//
// https://tannerhelland.com/2012/09/18/convert-temperature-rgb-algorithm-code.html
func KelvinTemperatureColor(kelvinTemperature float64) Color {
	var r, g, b float64

	// Temperature must fall between 1000 and 40000 degrees
	kelvinTemperature = util.Clamp(1000, 40000, kelvinTemperature)

	// All calculations below require kelvinTemperature to be in hundreds of actual value
	kelvinTemperature /= 100.0

	// Red
	if kelvinTemperature <= 66 {
		r = 255.0
	} else {
		// Note: the R-squared value for this approximation is .988
		r = 329.698727446 * math.Pow(kelvinTemperature-60, -0.1332047592)
	}

	// Green
	if kelvinTemperature <= 66 {
		// Note: the R-squared value for this approximation is .996
		g = 99.4708025861*math.Log(kelvinTemperature) - 161.1195681661
	} else {
		// Note: the R-squared value for this approximation is .987
		g = 288.1221695283 * math.Pow(kelvinTemperature-60, -0.0755148492)
	}

	// Blue
	if kelvinTemperature >= 66 {
		b = 255
	} else if kelvinTemperature <= 19 {
		b = 0
	} else {
		// Note: the R-squared value for this approximation is .998
		b = 138.5177312231*math.Log(kelvinTemperature-10) - 305.0447927307
	}

	// Normalize r,g, and b values to range [0,1] and clamp them to make sure they stay in range.
	r = util.Clamp(0.0, 1.0, r/255.0)
	g = util.Clamp(0.0, 1.0, g/255.0)
	b = util.Clamp(0.0, 1.0, b/255.0)

	return NewColor(r, g, b)
}

func (c *Color) Copy() Color {
	color := *c
	return color
}

func (c *Color) ChannelAdd(color Color) *Color {
	c.R += color.R
	c.G += color.G
	c.B += color.B
	return c
}

func (c *Color) ChannelMultiply(color Color) *Color {
	c.R *= color.R
	c.G *= color.G
	c.B *= color.B
	return c
}

func (c *Color) Divide(divider float32) *Color {
	c.R /= divider
	c.G /= divider
	c.B /= divider
	return c
}

func (c *Color) Multiply(factor float32) *Color {
	c.R *= factor
	c.G *= factor
	c.B *= factor
	return c
}
