package main

import (
	"fmt"
	img "image"
	col "image/color"
	"image/png"
	"math"
	"os"
	"pathtracer/internal/pkg/scene"
)

func writeImage(filename string, width int, height int, pixeldata []scene.Color) {
	image := img.NewRGBA(img.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			pixelvalue := pixeldata[y*width+x]

			r := uint8(math.Round(float64(pixelvalue.R * 255.0)))
			g := uint8(math.Round(float64(pixelvalue.G * 255.0)))
			b := uint8(math.Round(float64(pixelvalue.B * 255.0)))

			image.Set(x, y, col.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Oups, no files for you today.")
		os.Exit(1)
	}
	defer f.Close()

	// Encode to `PNG` with `DefaultCompression` level then save to file
	err = png.Encode(f, image)
	if err != nil {
		fmt.Println("Oups, no image encode for you today.")
		os.Exit(1)
	}
}
