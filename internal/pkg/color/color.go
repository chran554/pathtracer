package color

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

var (
	White            = NewColorGrey(1)
	Black            = NewColorGrey(0)
	BlackTransparent = NewColorRGBA(0, 0, 0, 0)
	WhiteTransparent = NewColorRGBA(1, 1, 1, 0)
)

type Color struct{ R, G, B, A float32 }

// RGBA creates a new color
func (c *Color) RGBA() (r, g, b, a uint32) {
	return uint32(c.R * 0xffff), uint32(c.G * 0xffff), uint32(c.B * 0xffff), uint32(c.A * 0xffff)
}

func NewColor(r, g, b float64) Color {
	return Color{R: float32(r), G: float32(g), B: float32(b), A: 1.0}
}

func NewColorRGBA(r, g, b, a float64) Color {
	return Color{R: float32(r), G: float32(g), B: float32(b), A: float32(a)}
}

func NewColorGrey(greyIntensity float64) Color {
	return NewColor(greyIntensity, greyIntensity, greyIntensity)
}

// NewColorHex converts a hex RGB or RGBA string to a color.
// Hex string can contain an initial '#' or '0x' prefix.
// Without prefix the string have to be 6 or 8 hex characters long as RGBA values
// are specified in the range 0x00 (0) to 0xFF (255).
func NewColorHex(colorHex string) Color {
	hex := strings.TrimSpace(strings.Replace(strings.Replace(colorHex, "#", "", -1), "0x", "", -1))
	if (len(hex) != 6) && (len(hex) != 8) {
		panic(fmt.Sprintf("Could not convert hex '%s' to RGB value. Illegal length %d of hex value. Expected length of 6 or 8.", colorHex, len(hex)))
	}

	r, err1 := strconv.ParseInt(hex[0:2], 16, 16)
	g, err2 := strconv.ParseInt(hex[2:4], 16, 16)
	b, err3 := strconv.ParseInt(hex[4:6], 16, 16)

	var a int64 = 255
	var err4 error = nil
	if len(hex) == 8 {
		a, err4 = strconv.ParseInt(hex[6:8], 16, 16)
	}

	if (err1 != nil) || (err2 != nil) || (err3 != nil) || (err4 != nil) {
		panic(fmt.Sprintf("Could not convert hex '%s' to RGB value.", colorHex))
	}

	return NewColorRGBA(float64(r)/255.0, float64(g)/255.0, float64(b)/255.0, float64(a)/255.0)
}

func (c *Color) Copy() *Color {
	col := *c
	return &col
}

func (c *Color) ChannelAdd(color *Color) *Color {
	c.R += color.R
	c.G += color.G
	c.B += color.B
	c.A += color.A
	return c
}

func (c *Color) ChannelMultiply(color *Color) *Color {
	c.R *= color.R
	c.G *= color.G
	c.B *= color.B
	//c.A *= color.A
	return c
}

func (c *Color) Divide(divider float32) *Color {
	c.R /= divider
	c.G /= divider
	c.B /= divider
	// c.A /= divider
	return c
}

func (c *Color) Multiply(factor float32) *Color {
	c.R *= factor
	c.G *= factor
	c.B *= factor
	//c.A *= factor
	return c
}

func (c *Color) Fade(destColor Color, factor float32) *Color {
	dstContribution := destColor.Multiply(factor)
	srcContribution := c.Multiply(1.0 - factor)
	return srcContribution.Copy().ChannelAdd(dstContribution)
}

// CloseRGB checks whether a color is close to another (within a tolerance distance in RGB space).
// No gamma is considered or expanded.
// Alpha channel is not considered, as only the color is compared.
func (c *Color) CloseRGB(compareColor Color, delta float64) bool {
	rD := float64(c.R - compareColor.R)
	gD := float64(c.G - compareColor.G)
	bD := float64(c.B - compareColor.B)
	return math.Sqrt(rD*rD+gD*gD+bD*bD) <= delta
}

// GammaEncode (or gamma compression) converts this color with values in linear space to a new color with values in gamma space.
//
// https://blog.johnnovak.net/2016/09/21/what-every-coder-should-know-about-gamma/
func (c *Color) GammaEncode(gamma float64) *Color {
	return GammaEncodeColor(c, gamma)
}

// GammaDecode (or gamma expansion) converts this color with values in gamma space to a new color with values in linear space.
//
// https://blog.johnnovak.net/2016/09/21/what-every-coder-should-know-about-gamma/
func (c *Color) GammaDecode(gamma float64) *Color {
	return GammaDecodeColor(c, gamma)
}

// GammaEncodeColor (or gamma compression) converts a color with values in linear space to a new color with values in gamma space.
//
// https://blog.johnnovak.net/2016/09/21/what-every-coder-should-know-about-gamma/
func GammaEncodeColor(linearColor *Color, gamma float64) *Color {
	invGamma := 1.0 / gamma
	gammaColor := &Color{
		R: gammaCalculation(linearColor.R, invGamma),
		G: gammaCalculation(linearColor.G, invGamma),
		B: gammaCalculation(linearColor.B, invGamma),
		A: linearColor.A,
	}

	return gammaColor
}

// GammaDecodeColor (or gamma expansion) converts a color with values in gamma space to a new color with values in linear space.
//
// https://blog.johnnovak.net/2016/09/21/what-every-coder-should-know-about-gamma/
func GammaDecodeColor(gammaColor *Color, gamma float64) *Color {
	linearColor := &Color{
		R: gammaCalculation(gammaColor.R, gamma),
		G: gammaCalculation(gammaColor.G, gamma),
		B: gammaCalculation(gammaColor.B, gamma),
		A: gammaColor.A,
	}
	return linearColor
}

func gammaCalculation(value float32, gamma float64) float32 {
	return float32(math.Pow(float64(value), gamma))
}
