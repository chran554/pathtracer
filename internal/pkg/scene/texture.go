package scene

import "pathtracer/internal/pkg/floatimage"

type TextureType string

const (
	TextureTypeAlbedo       TextureType = "albedo"
	TextureTypeNormalMap    TextureType = "normalmap"
	TextureTypeRoughness    TextureType = "roughness"
	TextureTypeTransparency TextureType = "transparency"
	TextureTypeMetallicness TextureType = "metallicness"
)

func (t TextureType) String() string { return string(t) }

type Texture struct {
	Type          TextureType
	Strength      float64
	Interpolation floatimage.Interpolation
	Image         *floatimage.FloatImage
}
