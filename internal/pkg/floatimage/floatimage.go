package floatimage

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	img "image"
	col "image/color"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/util"
	"strconv"

	"github.com/vmihailenco/msgpack/v5"
)

var (
	GammaSRGB    = 2.2
	GammaDefault = GammaSRGB
)

type FloatImage struct {
	name   string
	pixels []*color.Color
	Width  int
	Height int
	_hash  string
}

func NewFloatImage(name string, width, height int) *FloatImage {
	floatImage := FloatImage{
		name:   name,
		pixels: make([]*color.Color, width*height),
		Width:  width,
		Height: height,
	}

	for i := range floatImage.pixels {
		floatImage.pixels[i] = color.Black.Copy()
	}

	return &floatImage
}

func (fi *FloatImage) Bounds() img.Rectangle {
	return img.Rect(0, 0, fi.Width, fi.Height)
}

func (fi *FloatImage) At(x, y int) col.Color {
	return fi.GetPixel(x, y)
}

func (fi *FloatImage) Set(x, y int, c col.Color) {
	fi.SetPixel(x, y, c.(*color.Color))
}

func (fi *FloatImage) ColorModel() col.Model { return color.FloatNRGBAModel }

func (fi *FloatImage) String() string {
	return fmt.Sprintf("%s (%dx%d)", fi.name, fi.Width, fi.Height)
}

func (fi *FloatImage) Copy() *FloatImage {
	return &FloatImage{
		name:   fi.name,
		pixels: append([]*color.Color{}, fi.pixels...),
		Width:  fi.Width,
		Height: fi.Height,
	}
}

func (fi *FloatImage) Hash() string {
	if fi._hash == "" {
		data := []byte(fi.name)
		if len(fi.pixels) > 0 {
			data, _ = msgpack.Marshal(fi.pixels)
		}
		sum256 := sha256.Sum256(data)
		fi._hash = base64.URLEncoding.EncodeToString(sum256[:])
	}
	return fi._hash
}

func (fi *FloatImage) Name() string {
	return fi.name
}

func (fi *FloatImage) ContainImageData() bool {
	return (fi.Width > 0) && (fi.Height > 0) && (fi.pixels != nil)
}

// GetPixel returns the pixel at the given coordinates.
// This is a reference to the pixel value, so changing the pixel value will change the image.
func (fi *FloatImage) GetPixel(x, y int) *color.Color {
	if (x >= fi.Width) || (y >= fi.Height) || (x < 0) || (y < 0) {
		fmt.Printf("Illegal pixel access in image \"%s\" of size (%d x %d). There is no pixel at (%d x %d).\n", fi.name, fi.Width, fi.Height, x, y)
	}
	return fi.pixels[y*fi.Width+x]
}

func (fi *FloatImage) SetPixel(x, y int, color *color.Color) {
	fi.pixels[y*fi.Width+x].Set(color)
}

func (fi *FloatImage) Fill(color *color.Color) {
	for i := range fi.pixels {
		fi.pixels[i].Set(color)
	}
}

// EmptyPlaceholderImage returns a new FloatImage with the given filename as the name.
// The image is empty and of no size (w=0, h=0).
// This is useful for images that are referenced during scene creation where the image is not but rather the filename.
func EmptyPlaceholderImage(filename string) (*FloatImage, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("file \"%s\" does not exist", filename)
	} else if err != nil {
		return nil, err
	} else if info.IsDir() {
		return nil, fmt.Errorf("file \"%s\" is a directory", filename)
	} else if info.Size() == 0 {
		return nil, fmt.Errorf("file \"%s\" is empty", filename)
	}

	return NewFloatImage(filename, 0, 0), nil
}

func Load(filename string) (*FloatImage, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	image, err := Read(filename, f)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func LoadOrPanic(filename string) *FloatImage {
	image, err := Load(filename)
	if err != nil {
		message := fmt.Sprintf("image file \"%s\" could not be loaded: %s", filename, err.Error())
		panic(message)
	}
	return image
}

// Read reads an image from the given reader.
// No gama correction is applied.
func Read(imageName string, r io.Reader) (*FloatImage, error) {
	image, _, err := img.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("could not decode image \"%s\": %w", imageName, err)
	}

	floatImage := ConvertImageToFloatImage(imageName, image)
	return floatImage, nil
}

// ConvertImageToFloatImage converts an image.Image to a FloatImage.
// No gama correction is applied.
func ConvertImageToFloatImage(imageName string, textureImage img.Image) *FloatImage {
	width := textureImage.Bounds().Max.X
	height := textureImage.Bounds().Max.Y

	image := NewFloatImage(imageName, width, height)
	colorNormalizationFactor := 1.0 / 0xff
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c1 := textureImage.At(x, y)
			nrgbaColor := col.NRGBAModel.Convert(c1).(col.NRGBA)

			c2 := color.NewColorRGBA(
				float64(nrgbaColor.R)*colorNormalizationFactor,
				float64(nrgbaColor.G)*colorNormalizationFactor,
				float64(nrgbaColor.B)*colorNormalizationFactor,
				float64(nrgbaColor.A)*colorNormalizationFactor,
			)

			image.SetPixel(x, y, c2)
		}
	}
	return image
}

func WriteImage(filename string, floatImage *FloatImage) {
	width := floatImage.Width
	height := floatImage.Height

	tmp := floatImage.Copy()
	tmp.GammaEncode(GammaDefault)

	image := img.NewNRGBA(img.Rect(0, 0, width, height))
	// imageAlpha := img.NewRGBA(img.Rect(0, 0, width, height))

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			pixelValue := tmp.GetPixel(x, y)

			r := uint8(util.ClampFloat64(0, 255, math.Round(float64(pixelValue.R)*255.0)))
			g := uint8(util.ClampFloat64(0, 255, math.Round(float64(pixelValue.G)*255.0)))
			b := uint8(util.ClampFloat64(0, 255, math.Round(float64(pixelValue.B)*255.0)))
			a := uint8(util.ClampFloat64(0, 255, math.Round(float64(pixelValue.A)*255.0)))

			image.Set(x, y, col.NRGBA{R: r, G: g, B: b, A: a})
			// imageAlpha.Set(x, y, col.RGBA{R: a, G: a, B: a, A: 255})
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Oups, no files for you today.")
		os.Exit(1)
	}
	defer f.Close()

	// Encode to `PNG` with `BestCompression` level then save to file
	var encoder png.Encoder
	encoder.CompressionLevel = png.BestCompression
	err = encoder.Encode(f, image)
	if err != nil {
		fmt.Println("Oups, no image encode for you today.")
		os.Exit(1)
	}

	// falpha, _ := os.Create(filename + ".alpha.png")
	// defer falpha.Close()
	// encoder.Encode(falpha, imageAlpha)
}

func (fi *FloatImage) Image() *img.NRGBA {
	width := fi.Width
	height := fi.Height

	tmp := fi.Copy()
	tmp.GammaEncode(GammaDefault)

	newImage := img.NewNRGBA(img.Rect(0, 0, width, height))
	// imageAlpha := img.NewRGBA(img.Rect(0, 0, width, height))

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			pixelValue := tmp.GetPixel(x, y)

			r := uint8(util.ClampFloat64(0, 255, math.Round(float64(pixelValue.R)*255.0)))
			g := uint8(util.ClampFloat64(0, 255, math.Round(float64(pixelValue.G)*255.0)))
			b := uint8(util.ClampFloat64(0, 255, math.Round(float64(pixelValue.B)*255.0)))
			a := uint8(util.ClampFloat64(0, 255, math.Round(float64(pixelValue.A)*255.0)))

			newImage.Set(x, y, col.NRGBA{R: r, G: g, B: b, A: a})
			// imageAlpha.Set(x, y, col.RGBA{R: a, G: a, B: a, A: 255})
		}
	}
	return newImage
}

func WriteRawImage(filename string, image *FloatImage) {
	var byteBuffer bytes.Buffer

	width := image.Width
	height := image.Height

	fileFormatVersionMajor := 1
	fileFormatVersionMinor := 0

	writeBinaryInt32(&byteBuffer, int32(fileFormatVersionMajor))
	writeBinaryInt32(&byteBuffer, int32(fileFormatVersionMinor))
	writeBinaryInt32(&byteBuffer, int32(width))
	writeBinaryInt32(&byteBuffer, int32(height))

	for _, pixel := range image.pixels {
		if err := binary.Write(&byteBuffer, binary.BigEndian, pixel); err != nil {
			fmt.Printf("could not write raw image pixel data: %v\n", err)
		}
	}

	byteData := byteBuffer.Bytes()
	length := byteBuffer.Len()

	err := os.WriteFile(filename, byteData, 0644)
	if err != nil {
		fmt.Println("could not write raw image file:", filename)
		os.Exit(1)
	} else {
		fmt.Println("Wrote raw image file \"" + filename + "\" of size " + util.ByteCountIEC(int64(length)) + " (" + strconv.Itoa(length) + " bytes)")
	}
}

func writeBinaryFloat64(buffer *bytes.Buffer, value float64) {
	if err := binary.Write(buffer, binary.BigEndian, value); err != nil {
		fmt.Println(err)
	}
}

func writeBinaryFloat32(buffer *bytes.Buffer, value float32) {
	if err := binary.Write(buffer, binary.BigEndian, value); err != nil {
		fmt.Println(err)
	}
}

func writeBinaryInt32(buffer *bytes.Buffer, value int32) {
	if err := binary.Write(buffer, binary.BigEndian, value); err != nil {
		fmt.Println(err)
	}
}

// GammaEncode (or gamma compression) converts this image with values in linear space to have values in gamma space.
//
// https://blog.johnnovak.net/2016/09/21/what-every-coder-should-know-about-gamma/
func (fi *FloatImage) GammaEncode(gamma float64) {
	for y := 0; y < fi.Height; y++ {
		for x := 0; x < fi.Width; x++ {
			linearPixelValue := fi.GetPixel(x, y)
			gammaPixelValue := linearPixelValue.GammaEncode(gamma)
			fi.SetPixel(x, y, gammaPixelValue)
		}
	}
}

// GammaDecode (or gamma expansion) converts this image with values in gamma space to have values in linear space.
//
// https://blog.johnnovak.net/2016/09/21/what-every-coder-should-know-about-gamma/
func (fi *FloatImage) GammaDecode(gamma float64) {
	for y := 0; y < fi.Height; y++ {
		for x := 0; x < fi.Width; x++ {
			gammaPixelValue := fi.GetPixel(x, y)
			linearPixelValue := gammaPixelValue.GammaDecode(gamma)
			fi.SetPixel(x, y, linearPixelValue)
		}
	}
}
