package color

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_RgbToKelvin(t *testing.T) {
	minKelvin := 1000.0
	maxKelvin := 40000.0
	step := 25.0

	for kelvin := minKelvin; kelvin <= maxKelvin; kelvin += step {
		temperatureColor := KelvinTemperatureColor2(kelvin)
		kelvinCct := KelvinTemperatureCCT(temperatureColor)

		kelvinDiff := int(kelvinCct - kelvin)
		if kelvinDiff != 0 {
			fmt.Printf("Temperature (in Kelvin): Original=%5d  Derived=%5d    Difference=%d\n", int(kelvin), int(kelvinCct), kelvinDiff)
		}
	}
}

func Test_KelvinToRGBTable(t *testing.T) {
	//t.Skip("Generates a table on stdout. Run manually.")

	minKelvin := 0.0
	maxKelvin := 40000.0
	step := 25.0

	for kelvin := minKelvin; kelvin <= maxKelvin; kelvin += step {
		c := KelvinTemperatureColor(kelvin)
		fmt.Printf("%d, %f, %f, %f\n", int(kelvin), c.R, c.G, c.B)
	}
}

func Test_KelvinToRgb(t *testing.T) {
	t.Skip("Generates an image. Run manually.")

	minKelvin := 0
	maxKelvin := 40000
	width := (maxKelvin - minKelvin) / 25
	height := 500

	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	for x := 0; x <= width; x++ {
		kelvin := float64(minKelvin) + float64(x*(maxKelvin-minKelvin))/float64(width)
		rgb1 := KelvinTemperatureColor(kelvin)
		rgb2 := KelvinTemperatureColor2(kelvin)

		inset := 1
		scaleColor := Black
		if math.Abs(kelvin-6500) <= 10 {
			inset = 10
			scaleColor = NewColor(1.0, 0.3, 1.0)
		} else if int(kelvin)%1000 <= 10 {
			inset = 10
		} else if int(kelvin)%500 <= 10 {
			inset = 5
		}

		drawLine(img, x, 0, height/2, rgb1)
		drawLine(img, x, height/2, height-1, rgb2)

		// Draw a temperature scale line
		scalesSeparation := 15
		drawLine(img, x, height/2-(inset-1)-scalesSeparation, height/2+(inset-1)-scalesSeparation, scaleColor)
		drawLine(img, x, height/2-(inset-1)+scalesSeparation, height/2+(inset-1)+scalesSeparation, scaleColor)

		fmt.Printf("Temperature %5d:     RGB1:%+v     RGB2:%+v\n", int(kelvin), rgb1, rgb2)
	}

	pngFile, err := os.OpenFile("kelvin to rgb.png", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	assert.NoError(t, err)
	defer pngFile.Close()

	err = png.Encode(pngFile, img)
	assert.NoError(t, err)
}

func drawLine(img *image.NRGBA, x int, y1 int, y2 int, rgb Color) {
	if x >= 0 && x < img.Bounds().Dx() {
		for y := y1; y <= y2; y++ {
			if y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, &rgb)
			}
		}
	}
}
