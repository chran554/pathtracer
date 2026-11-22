package scene

import "pathtracer/internal/pkg/floatimage"

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
