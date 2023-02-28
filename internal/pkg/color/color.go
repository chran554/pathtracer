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

// KelvinTemperatureColor2 gives rgb-value for a Kelvin temperature
// using a more accurate version algorithm based on a different curve fit to the original RGB to Kelvin data.
//
// https://github.com/neilbartlett/color-temperature/blob/master/index.js
func KelvinTemperatureColor2(kelvinTemperature float64) Color {
	// Temperature must fall between 1000 and 40000 degrees
	kelvinTemperature = util.Clamp(1000, 40000, kelvinTemperature)

	var temperature = kelvinTemperature / 100.0
	var red, green, blue float64

	if temperature < 66.0 {
		red = 255
	} else {
		// a + b x + c Log[x] /.
		// {a -> 351.97690566805693`,
		// b -> 0.114206453784165`,
		// c -> -40.25366309332127
		//x -> (kelvin/100) - 55}
		red = temperature - 55.0
		red = 351.97690566805693 + 0.114206453784165*red - 40.25366309332127*math.Log(red)
	}

	if temperature < 66.0 {
		// a + b x + c Log[x] /.
		// {a -> -155.25485562709179`,
		// b -> -0.44596950469579133`,
		// c -> 104.49216199393888`,
		// x -> (kelvin/100) - 2}
		green = temperature - 2
		green = -155.25485562709179 - 0.44596950469579133*green + 104.49216199393888*math.Log(green)

	} else {
		// a + b x + c Log[x] /.
		// {a -> 325.4494125711974`,
		// b -> 0.07943456536662342`,
		// c -> -28.0852963507957`,
		// x -> (kelvin/100) - 50}
		green = temperature - 50.0
		green = 325.4494125711974 + 0.07943456536662342*green - 28.0852963507957*math.Log(green)
	}

	if temperature >= 66.0 {
		blue = 255
	} else {

		if temperature <= 20.0 {
			blue = 0
		} else {
			// a + b x + c Log[x] /.
			// {a -> -254.76935184120902`,
			// b -> 0.8274096064007395`,
			// c -> 115.67994401066147`,
			// x -> kelvin/100 - 10}
			blue = temperature - 10
			blue = -254.76935184120902 + 0.8274096064007395*blue + 115.67994401066147*math.Log(blue)
		}
	}

	red = util.Clamp(0.0, 1.0, red/255.0)
	green = util.Clamp(0.0, 1.0, green/255.0)
	blue = util.Clamp(0.0, 1.0, blue/255.0)

	return NewColor(red, green, blue)
}

func KelvinTemperatureCCT(c Color) float64 {
	var temperature float64

	epsilon := 0.4
	minTemperature := 1000.0
	maxTemperature := 40000.0

	for (maxTemperature - minTemperature) > epsilon {
		temperature = (maxTemperature + minTemperature) / 2
		testRGB := KelvinTemperatureColor2(temperature)

		if (testRGB.B / testRGB.R) >= (c.B / c.R) {
			maxTemperature = temperature
		} else {
			minTemperature = temperature
		}
	}

	return temperature
}
