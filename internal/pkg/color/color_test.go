package color

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
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

	assert.Equal(t, c, c2)
	assert.True(t, &c != &c2)
}

func Test_RgbToKelvin(t *testing.T) {
	t.Run("RGB to Kelvin", func(t *testing.T) {
		minKelvin := 1000.0
		maxKelvin := 40000.0
		step := 50.0

		for kelvin := minKelvin; kelvin <= maxKelvin; kelvin += step {
			temperatureColor := KelvinTemperatureColor2(kelvin)
			kelvinCct := KelvinTemperatureCCT(temperatureColor)

			fmt.Printf("Temperature %5d:   %5d    delta: %d\n", int(kelvin), int(kelvinCct), int(kelvinCct)-int(kelvin))
		}
	})
}

/*
func Test_KelvinToRgb(t *testing.T) {
	t.Run("kelvin to RGB", func(t *testing.T) {
		minKelvin := 1000
		maxKelvin := 40000
		width := (maxKelvin - minKelvin) / 50
		height := 500
		pic := img.NewFloatImage("kelvin to rgb", width, height)

		for x := 0; x <= width; x++ {
			kelvin := float64(minKelvin) + float64(x*(maxKelvin-minKelvin))/float64(width)
			rgb1 := KelvinTemperatureColor(float64(kelvin))
			rgb2 := KelvinTemperatureColor2(float64(kelvin))

			inset := 1
			if int(kelvin)%5000 <= 39 {
				inset = 10
			} else if int(kelvin)%500 <= 39 {
				inset = 5
			}

			drawLine(pic, x, 0, height/2-inset, &rgb1)
			drawLine(pic, x, height/2+inset, height-1, &rgb2)

			fmt.Printf("Temperature %5d:     RGB1:%+v     RGB2:%+v\n", int(kelvin), rgb1, rgb2)
		}

		img.WriteImage("kelvin2rgb.png", pic)
	})
}

func drawLine(pic *img.FloatImage, x int, y1 int, y2 int, rgb *Color) {
	if x >= 0 && x < pic.Width {
		for y := y1; y <= y2; y++ {
			if y >= 0 && y < pic.Height {
				pic.SetPixel(x, y, rgb)
			}
		}
	}
}
*/
