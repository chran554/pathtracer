package scene

import (
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
)

type TextureType string

const (
	TextureTypeAlbedo            TextureType = "albedo"
	TextureTypeNormalMap         TextureType = "normalmap"
	TextureTypeNormalMap_OpenGL  TextureType = "normalmap-opengl"
	TextureTypeNormalMap_DirectX TextureType = "normalmap-directx"
	TextureTypeRoughness         TextureType = "roughness"
	TextureTypeTransparency      TextureType = "transparency"
	TextureTypeMetallicness      TextureType = "metallicness"
)

func (t TextureType) String() string { return string(t) }

type Texture struct {
	Image         *floatimage.FloatImage
	Type          TextureType
	Strength      float64
	Interpolation floatimage.Interpolation
}

func NewTexture(image *floatimage.FloatImage, textureType TextureType) *Texture {
	return &Texture{
		Image:         image,
		Type:          textureType,
		Strength:      1.0,
		Interpolation: floatimage.InterpolationNearestNeighbor,
	}
}

func NewTextureReferenceOrPanic(imagePath string, textureType TextureType) *Texture {
	image := floatimage.ReferenceOrPanic(imagePath)

	return &Texture{
		Image:         image,
		Type:          textureType,
		Strength:      1.0,
		Interpolation: floatimage.InterpolationNearestNeighbor,
	}
}

// S set the strength of the texture.
// It depends on the texture type whether "strength" is applicable or what aspect the parameter will control.
func (t *Texture) S(s float64) *Texture {
	t.Strength = s
	return t
}

// I set the interpolation mode of the texture.
func (t *Texture) I(i floatimage.Interpolation) *Texture {
	t.Interpolation = i
	return t
}

// GetInterpolatedPixel gets the interpolated pixel color at (x, y).
// x is width relative in range [0, 1).
// y is height relative in range [0, 1).
func (t *Texture) GetInterpolatedPixel(x float64, y float64) *color.Color {
	textureX := x * float64(t.Image.Width)
	textureY := y * float64(t.Image.Height)

	c := t.Image.GetInterpolatedPixel(textureX, textureY, t.Interpolation)

	switch t.Type {
	case TextureTypeAlbedo:
		c.A = c.A * float32(t.Strength)

	case TextureTypeRoughness:
		c.Multiply(float32(t.Strength))

	case TextureTypeMetallicness:
		c.Multiply(float32(t.Strength))
	}

	return c
}
