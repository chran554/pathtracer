package cie

import (
	"github.com/stretchr/testify/assert"
	"pathtracer/internal/pkg/color"
	"testing"
)

func TestXYZtoRGB(t *testing.T) {
	c := color.Color{R: 1.0, G: 1.0, B: 1.0, A: 1.0}

	xyz := RGBColorToXYZ(c, SRGB_D65_RGBtoXYZ, 2.4, 100)
	c1 := xyz.RGB(SRGB_D65_XYZtoRGB, 2.4)
	c2 := xyz.D65_SRGB()

	cie1931Xyz := IlluminantD65.CIE1931_XYZ(Observer2Deg)
	c3 := cie1931Xyz.RGB(SRGB_D65_XYZtoRGB, 2.4)
	c4 := cie1931Xyz.D65_SRGB()

	assert.True(t, c.CloseRGB(c1, 0.001))
	assert.True(t, c.CloseRGB(c2, 0.001))
	assert.True(t, c.CloseRGB(c3, 0.001))
	assert.True(t, c.CloseRGB(c4, 0.001))
}
