package test

import (
	"fmt"
	"pathtracer/internal/pkg/color"
	img "pathtracer/internal/pkg/image"
	"testing"
)

func Test_RgbToKelvin(t *testing.T) {
	t.Run("RGB to Kelvin", func(t *testing.T) {
		minKelvin := 1000.0
		maxKelvin := 40000.0
		step := 50.0

		for kelvin := minKelvin; kelvin <= maxKelvin; kelvin += step {
			temperatureColor := color.KelvinTemperatureColor2(kelvin)
			kelvinCct := color.KelvinTemperatureCCT(temperatureColor)

			fmt.Printf("Temperature %5d:   %5d    delta: %d\n", int(kelvin), int(kelvinCct), int(kelvinCct)-int(kelvin))
		}
	})
}

func Test_KelvinToRgb(t *testing.T) {
	t.Run("kelvin to RGB", func(t *testing.T) {
		minKelvin := 1000
		maxKelvin := 40000
		width := (maxKelvin - minKelvin) / 50
		height := 500
		pic := img.NewFloatImage("kelvin to rgb", width, height)

		for x := 0; x <= width; x++ {
			kelvin := float64(minKelvin) + float64(x*(maxKelvin-minKelvin))/float64(width)
			rgb1 := color.KelvinTemperatureColor(float64(kelvin))
			rgb2 := color.KelvinTemperatureColor2(float64(kelvin))

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

func drawLine(pic *img.FloatImage, x int, y1 int, y2 int, rgb *color.Color) {
	if x >= 0 && x < pic.Width {
		for y := y1; y <= y2; y++ {
			if y >= 0 && y < pic.Height {
				pic.SetPixel(x, y, rgb)
			}
		}
	}
}
