package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	img "image"
	col "image/color"
	"image/png"
	"math"
	"os"
	"pathtracer/internal/pkg/scene"
	"strconv"
)

func writeImage(filename string, width int, height int, pixelData []scene.Color) {
	image := img.NewRGBA(img.Rect(0, 0, width, height))

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			pixelValue := pixelData[y*width+x]

			r := uint8(math.Round(pixelValue.R * 255.0))
			g := uint8(math.Round(pixelValue.G * 255.0))
			b := uint8(math.Round(pixelValue.B * 255.0))

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

func writeRawImage(filename string, width int, height int, pixelData []scene.Color) {
	var byteBuffer bytes.Buffer

	fileFormatVersionMajor := 1
	fileFormatVersionMinor := 0

	writeBinaryInt32(&byteBuffer, int32(fileFormatVersionMajor))
	writeBinaryInt32(&byteBuffer, int32(fileFormatVersionMinor))
	writeBinaryInt32(&byteBuffer, int32(width))
	writeBinaryInt32(&byteBuffer, int32(height))

	if err := binary.Write(&byteBuffer, binary.BigEndian, pixelData); err != nil {
		fmt.Println(err)
	}

	//for x := 0; x < width; x++ {
	//	for y := 0; y < height; y++ {
	//		pixelValue := pixelData[y*width+x]
	//		writeBinaryFloat64(&byteBuffer, pixelValue.R)
	//		writeBinaryFloat64(&byteBuffer, pixelValue.G)
	//		writeBinaryFloat64(&byteBuffer, pixelValue.B)
	//	}
	//}

	byteData := byteBuffer.Bytes()
	length := byteBuffer.Len()

	err := os.WriteFile(filename, byteData, 0644)
	if err != nil {
		fmt.Println("Could not write raw image file:", filename)
		os.Exit(1)
	} else {
		fmt.Println("Wrote raw image file \"" + filename + "\" of size " + ByteCountIEC(int64(length)) + " (" + strconv.Itoa(length) + " bytes)")
	}
}

func writeBinaryFloat64(buffer *bytes.Buffer, value float64) {
	if err := binary.Write(buffer, binary.BigEndian, value); err != nil {
		fmt.Println(err)
	}
}

func writeBinaryInt32(buffer *bytes.Buffer, value int32) {
	if err := binary.Write(buffer, binary.BigEndian, value); err != nil {
		fmt.Println(err)
	}
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
