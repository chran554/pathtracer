package floatimage

import (
	"bytes"
	"fmt"
	"image"
	col "image/color"
	"image/png"
	"os"
	"pathtracer/internal/pkg/color"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFloatImage(t *testing.T) {
	fi := NewFloatImage("test", 10, 20)
	assert.Equal(t, "test", fi.Name())
	assert.Equal(t, 10, fi.Width)
	assert.Equal(t, 20, fi.Height)
	assert.Equal(t, 10*20, len(fi.pixels))
	assert.True(t, fi.ContainImageData())
}

func TestFloatImage_Bounds(t *testing.T) {
	fi := NewFloatImage("test", 10, 20)
	bounds := fi.Bounds()
	assert.Equal(t, image.Rect(0, 0, 10, 20), bounds)
}

func TestFloatImage_PixelAccess(t *testing.T) {
	fi := NewFloatImage("test", 2, 2)

	tests := []struct {
		x, y  int
		color *color.Color
	}{
		{0, 0, color.NewColor(1, 0, 0)},
		{1, 1, color.NewColor(0, 1, 0)},
		{0, 1, color.White},
		{1, 0, color.Black},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Pixel_%d_%d", tt.x, tt.y), func(t *testing.T) {
			fi.SetPixel(tt.x, tt.y, tt.color)
			assert.Equal(t, *tt.color, *fi.GetPixel(tt.x, tt.y))
		})
	}

	// Test At and Set (image.Image interface)
	fi.Set(0, 1, color.White)
	at01 := fi.At(0, 1)
	r, g, b, a := at01.RGBA()
	assert.Equal(t, uint32(0xffff), r)
	assert.Equal(t, uint32(0xffff), g)
	assert.Equal(t, uint32(0xffff), b)
	assert.Equal(t, uint32(0xffff), a)
}

func TestFloatImage_Fill(t *testing.T) {
	fi := NewFloatImage("test", 10, 10)
	c := color.NewColor(0.5, 0.5, 0.5)
	fi.Fill(c)

	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			assert.Equal(t, *c, *fi.GetPixel(x, y))
		}
	}
}

func TestFloatImage_Copy(t *testing.T) {
	fi := NewFloatImage("test", 2, 2)
	fi.SetPixel(0, 0, color.NewColor(1, 1, 1))
	cp := fi.Copy()

	assert.Equal(t, fi.Name(), cp.Name())
	assert.Equal(t, fi.Width, cp.Width)
	assert.Equal(t, fi.Height, cp.Height)
	assert.Equal(t, fi.GetPixel(0, 0), cp.GetPixel(0, 0))

	cp.SetPixel(0, 0, color.NewColor(0, 0, 0))
	assert.NotEqual(t, fi.GetPixel(0, 0), cp.GetPixel(0, 0))
}

func TestFloatImage_String(t *testing.T) {
	fi := NewFloatImage("test", 10, 20)
	assert.Equal(t, "test (10x20)", fi.String())
}

func TestFloatImage_ColorModel(t *testing.T) {
	fi := NewFloatImage("test", 1, 1)
	assert.NotNil(t, fi.ColorModel())
}

func TestFloatImage_Hash(t *testing.T) {
	fi1 := NewFloatImage("test", 2, 2)
	fi2 := NewFloatImage("test", 2, 2)
	assert.Equal(t, fi1.Hash(), fi2.Hash())

	fi1.SetPixel(0, 0, color.NewColor(1, 1, 1))
	assert.NotEqual(t, fi1.Hash(), fi2.Hash())
}

func TestFloatImage_HashInvalidation(t *testing.T) {
	fi := NewFloatImage("test", 2, 2)
	h1 := fi.Hash()
	assert.NotEmpty(t, h1)

	// Test SetPixel
	fi.SetPixel(0, 0, color.White)
	assert.Empty(t, fi._hash)
	h2 := fi.Hash()
	assert.NotEqual(t, h1, h2)

	// Test Fill
	fi.Fill(color.Black)
	assert.Empty(t, fi._hash)
	h3 := fi.Hash()
	assert.NotEqual(t, h2, h3)

	// Test GammaEncode
	fi.Fill(color.NewColor(0.5, 0.5, 0.5))
	h4 := fi.Hash()
	fi.GammaEncode()
	assert.Empty(t, fi._hash)
	h5 := fi.Hash()
	assert.NotEqual(t, h4, h5)

	// Test GammaDecode
	fi.GammaDecode()
	assert.Empty(t, fi._hash)
	h6 := fi.Hash()
	assert.NotEqual(t, h5, h6)
}

func TestEmptyPlaceholderImage(t *testing.T) {
	// Create a dummy file
	tmpFile, err := os.CreateTemp("", "test_image")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("some data")
	require.NoError(t, err)
	tmpFile.Close()

	fi, err := EmptyPlaceholderImage(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, tmpFile.Name(), fi.Name())
	assert.Equal(t, 0, fi.Width)
	assert.False(t, fi.ContainImageData())

	// Test non-existent file
	_, err = EmptyPlaceholderImage("non_existent_file")
	assert.Error(t, err)

	// Test directory
	dir, err := os.MkdirTemp("", "test_dir")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	_, err = EmptyPlaceholderImage(dir)
	assert.Error(t, err)
}

func TestRead(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, col.NRGBA{R: 255, G: 0, B: 0, A: 255})

	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	require.NoError(t, err)

	fi, err := Read("test.png", &buf)
	assert.NoError(t, err)
	assert.Equal(t, 2, fi.Width)
	assert.Equal(t, 2, fi.Height)

	c := fi.GetPixel(0, 0)
	assert.InDelta(t, 1.0, c.R, 0.001)
	assert.InDelta(t, 0.0, c.G, 0.001)
}

func TestConvertImageToFloatImage(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, col.NRGBA{R: 128, G: 64, B: 32, A: 255})

	fi := ConvertImageToFloatImage("test", img)
	assert.Equal(t, 2, fi.Width)

	c := fi.GetPixel(0, 0)
	assert.InDelta(t, 128.0/255.0, c.R, 0.001)
	assert.InDelta(t, 64.0/255.0, c.G, 0.001)
	assert.InDelta(t, 32.0/255.0, c.B, 0.001)
}

func TestFloatImage_Image(t *testing.T) {
	fi := NewFloatImage("test", 2, 2)
	fi.SetPixel(0, 0, color.NewColor(1, 0, 0))

	img := fi.Image()
	r, g, b, a := img.At(0, 0).RGBA()
	assert.Equal(t, uint32(0xffff), r)
	assert.Equal(t, uint32(0), g)
	assert.Equal(t, uint32(0), b)
	assert.Equal(t, uint32(0xffff), a)
}

func TestGamma(t *testing.T) {
	fi := NewFloatImage("test", 1, 1)
	fi.SetPixel(0, 0, color.NewColor(0.5, 0.5, 0.5))

	fi.GammaEncode()
	c1 := fi.GetPixel(0, 0)
	// 0.5 in linear space is approx 0.735357 in sRGB space
	assert.InDelta(t, 0.735357, c1.R, 0.001)

	// 0.5 ^ (1/2.2) is approx 0.7297
	// assert.InDelta(t, 0.7297, c1.R, 0.001)

	fi.GammaDecode()
	c2 := fi.GetPixel(0, 0)
	assert.InDelta(t, 0.5, c2.R, 0.001)
}

func TestWriteImage(t *testing.T) {
	fi := NewFloatImage("test", 2, 2)
	fi.SetPixel(0, 0, color.NewColor(1, 0, 0))

	tmpFile := "test_write.png"
	defer os.Remove(tmpFile)

	WriteImage(tmpFile, fi)

	_, err := os.Stat(tmpFile)
	assert.NoError(t, err)
}

func TestWriteRawImage(t *testing.T) {
	fi := NewFloatImage("test", 2, 2)
	fi.SetPixel(0, 0, color.NewColor(1, 0, 0))

	tmpFile := "test_write_raw.raw"
	defer os.Remove(tmpFile)

	WriteRawImage(tmpFile, fi)

	_, err := os.Stat(tmpFile)
	assert.NoError(t, err)
}

func TestLoad(t *testing.T) {
	// Reuse existing dice image for loading test if it exists,
	// or skip if it doesn't.
	imagePath := "../../../objects/obj/dice/dice1.png"
	if _, err := os.Stat(imagePath); err == nil {
		img, err := Load(imagePath)
		assert.NoError(t, err)
		assert.NotNil(t, img)
	} else {
		t.Skip("Dice image not found, skipping Load test")
	}
}

func TestLoadOrPanic(t *testing.T) {
	imagePath := "../../../objects/obj/dice/dice1.png"
	if _, err := os.Stat(imagePath); err == nil {
		assert.NotPanics(t, func() {
			LoadOrPanic(imagePath)
		})
	}

	assert.Panics(t, func() {
		LoadOrPanic("non_existent_file")
	})
}
