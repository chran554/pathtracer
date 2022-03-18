package image

import (
	"bytes"
	"encoding/binary"
	"fmt"
	_ "image/jpeg"
	"image/png"
	"pathtracer/internal/pkg/color"

	img "image"
	col "image/color"
	"math"
	"os"
	"strconv"
)

type FloatImage struct {
	pixels []color.Color
	Width  int
	Height int
}

func NewFloatImage(width, height int) *FloatImage {
	floatImage := FloatImage{
		pixels: make([]color.Color, width*height),
		Width:  width,
		Height: height,
	}
	return &floatImage
}

func (image *FloatImage) Clear() {
	image.pixels = nil
	image.Width = 0
	image.Height = 0
}

func (image *FloatImage) ContainImage() bool {
	return image.Width > 0 && image.Height > 0 && image.pixels != nil
}

func (image *FloatImage) GetPixel(x, y int) *color.Color {
	return &image.pixels[y*image.Width+x]
}

func (image *FloatImage) SetPixel(x, y int, color *color.Color) {
	image.pixels[y*image.Width+x] = *color
}

func LoadImageData(filename string) *FloatImage {
	textureImage, err := getImageFromFilePath(filename)
	if err != nil {
		fmt.Println("Ouupps, no image file could be loaded for parallel image projection.", filename)
		os.Exit(1)
	}

	width := textureImage.Bounds().Max.X
	height := textureImage.Bounds().Max.Y

	image := NewFloatImage(width, height)
	colorNormalizationFactor := 1.0 / 0xffff
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := textureImage.At(x, y).RGBA()

			c := color.Color{
				R: float64(r) * colorNormalizationFactor,
				G: float64(g) * colorNormalizationFactor,
				B: float64(b) * colorNormalizationFactor,
			}

			image.SetPixel(x, y, &c)
		}
	}

	return image
}

func getImageFromFilePath(filePath string) (img.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	image, _, err := img.Decode(f)

	// fmt.Println("Read image:", filePath)

	return image, err
}

func WriteImage(filename string, width int, height int, floatImage *FloatImage) {
	image := img.NewRGBA(img.Rect(0, 0, width, height))

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			pixelValue := floatImage.GetPixel(x, y)

			r := uint8(clamp(0, 255, math.Round(pixelValue.R*255.0)))
			g := uint8(clamp(0, 255, math.Round(pixelValue.G*255.0)))
			b := uint8(clamp(0, 255, math.Round(pixelValue.B*255.0)))

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

func clamp(min float64, max float64, value float64) float64 {
	if value < min {
		return min
	} else if value > max {
		return max
	} else {
		return value
	}

}

func WriteRawImage(filename string, width int, height int, pixelData *FloatImage) {
	var byteBuffer bytes.Buffer

	fileFormatVersionMajor := 1
	fileFormatVersionMinor := 0

	writeBinaryInt32(&byteBuffer, int32(fileFormatVersionMajor))
	writeBinaryInt32(&byteBuffer, int32(fileFormatVersionMinor))
	writeBinaryInt32(&byteBuffer, int32(width))
	writeBinaryInt32(&byteBuffer, int32(height))

	if err := binary.Write(&byteBuffer, binary.BigEndian, pixelData.pixels); err != nil {
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
