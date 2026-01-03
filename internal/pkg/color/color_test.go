package color

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewColor(t *testing.T) {
	var c *Color

	c = NewColor(0.0, 0.0, 0.0)
	assert.Equal(t, float32(0.0), c.R)
	assert.Equal(t, float32(0.0), c.G)
	assert.Equal(t, float32(0.0), c.B)

	c = NewColor(1.0, 1.0, 1.0)
	assert.Equal(t, float32(1.0), c.R)
	assert.Equal(t, float32(1.0), c.G)
	assert.Equal(t, float32(1.0), c.B)

	c = NewColor(0.5, 0.5, 0.5)
	assert.Equal(t, float32(0.5), c.R)
	assert.Equal(t, float32(0.5), c.G)
	assert.Equal(t, float32(0.5), c.B)

	c = NewColor(0.25, 0.50, 0.75)
	assert.Equal(t, float32(0.25), c.R)
	assert.Equal(t, float32(0.50), c.G)
	assert.Equal(t, float32(0.75), c.B)
}

func Test_NewGreyColor(t *testing.T) {
	var c *Color

	c = NewColorGrey(0.0)
	assert.Equal(t, float32(0.0), c.R)
	assert.Equal(t, float32(0.0), c.G)
	assert.Equal(t, float32(0.0), c.B)

	c = NewColorGrey(1.0)
	assert.Equal(t, float32(1.0), c.R)
	assert.Equal(t, float32(1.0), c.G)
	assert.Equal(t, float32(1.0), c.B)

	c = NewColorGrey(0.5)
	assert.Equal(t, float32(0.5), c.R)
	assert.Equal(t, float32(0.5), c.G)
	assert.Equal(t, float32(0.5), c.B)
}

func Test_NewHexColor(t *testing.T) {
	var c *Color

	c = NewColorHex("#000000")
	assert.Equal(t, float32(0.0), c.R)
	assert.Equal(t, float32(0.0), c.G)
	assert.Equal(t, float32(0.0), c.B)

	c = NewColorHex("#FFFFFF")
	assert.Equal(t, float32(1.0), c.R)
	assert.Equal(t, float32(1.0), c.G)
	assert.Equal(t, float32(1.0), c.B)

	c = NewColorHex("#AAAAAA")
	assert.Equal(t, float32(2.0/3), c.R)
	assert.Equal(t, float32(2.0/3), c.G)
	assert.Equal(t, float32(2.0/3), c.B)

	c = NewColorHex("AAAAAA")
	assert.Equal(t, float32(2.0/3), c.R)
	assert.Equal(t, float32(2.0/3), c.G)
	assert.Equal(t, float32(2.0/3), c.B)
}

func Test_Copy(t *testing.T) {
	c := NewColor(0.25, 0.50, 0.75)
	c2 := c.Copy()

	assert.Equal(t, float32(0.25), c.R)
	assert.Equal(t, float32(0.50), c.G)
	assert.Equal(t, float32(0.75), c.B)

	assert.Equal(t, float32(0.25), c2.R)
	assert.Equal(t, float32(0.50), c2.G)
	assert.Equal(t, float32(0.75), c2.B)

	assert.Equal(t, c, c2)  // Content is equal
	assert.True(t, c != c2) // References are not equal
}

func TestGammaSRGB_roundtrip(t *testing.T) {
	// Test round-trip for various values
	values := []float64{0.0, 0.001, 0.0031308, 0.005, 0.04045, 0.1, 0.5, 0.8, 1.0}

	for _, v := range values {
		c := NewColorGrey(v)
		encoded := c.GammaEncodeSRGB()
		decoded := encoded.GammaDecodeSRGB()

		assert.InDelta(t, v, float64(decoded.R), 1e-6, "Value %f failed round-trip", v)
		assert.InDelta(t, v, float64(decoded.G), 1e-6, "Value %f failed round-trip", v)
		assert.InDelta(t, v, float64(decoded.B), 1e-6, "Value %f failed round-trip", v)
	}
}

func TestGammaSRGB_values(t *testing.T) {
	// sRGB conversion test values
	testCases := []struct {
		linear float32
		sRGB   float32
		source string
	}{
		// Sources:
		// [1] https://en.wikipedia.org/wiki/SRGB#The_forward_transformation_(condition_EN_61966-2-1:1999)
		// [2] https://www.w3.org/Graphics/Color/sRGB.html
		// [3] http://www.brucelindbloom.com/index.html?ColorCalculator.html

		{0.0, 0.0, "[1] Black"},
		{1.0, 1.0, "[1] White"},
		{0.0031308, 0.04045, "[1] Boundary value"},
		{0.18, 0.461356, "[3] 18% gray (0.18 linear -> 0.461356 sRGB)"},
		{0.5, 0.735357, "[3] 0.5 linear -> 0.735357 sRGB"},
		{0.001, 0.01292, "[1] Low value (linear <= 0.0031308)"},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Encode/decode sRGB gamma: linear=%0.3f <--[convert]--> sRGB=%0.3f", testCase.linear, testCase.sRGB), func(t *testing.T) {
			c := NewColorGrey(float64(testCase.linear))
			encoded := c.GammaEncodeSRGB()
			assert.InDelta(t, float64(testCase.sRGB), float64(encoded.R), 1e-5, "Linear to sRGB failed for %f (%s)", testCase.linear, testCase.source)

			c2 := NewColorGrey(float64(testCase.sRGB))
			decoded := c2.GammaDecodeSRGB()
			assert.InDelta(t, float64(testCase.linear), float64(decoded.R), 1e-5, "sRGB to Linear failed for %f (%s)", testCase.sRGB, testCase.source)
		})
	}
}
