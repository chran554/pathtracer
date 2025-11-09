package color

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewColor(t *testing.T) {
	var c Color

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
	var c Color

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
	var c Color

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
	var c Color
	c = NewColor(0.25, 0.50, 0.75)
	c2 := c.Copy()

	assert.Equal(t, float32(0.25), c.R)
	assert.Equal(t, float32(0.50), c.G)
	assert.Equal(t, float32(0.75), c.B)

	assert.Equal(t, float32(0.25), c2.R)
	assert.Equal(t, float32(0.50), c2.G)
	assert.Equal(t, float32(0.75), c2.B)

	assert.Equal(t, c, *c2)
	assert.True(t, &c != c2)
}
