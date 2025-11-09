package scene

import (
	"math"
	"pathtracer/internal/pkg/color"

	"github.com/ungerik/go3d/float64/vec3"
)

type Material struct {
	Name            string           `json:"Name,omitempty"`
	Color           *color.Color     `json:"Color,omitempty"`
	Diffuse         float64          `json:"Diffuse,omitempty"`
	Emission        *color.Color     `json:"Emission,omitempty"`
	Glossiness      float64          `json:"Glossiness,omitempty"` // Glossiness is the percent amount that will make out specular reflection. Values [0.0 .. 1.0] with default 0.0. Lower value the more diffuse color will appear and higher value the more mirror reflection will appear.
	Roughness       float64          `json:"Roughness,omitempty"`  // Roughness is the diffuse spread of the specular reflection. Values [0.0 .. 1.0] with default 0.0. Lower is like "brushed metal" or "foggy/hazy reflection" and higher value give a more mirror like reflection. A value of 1.0 is perfect mirror reflection and a value of 0.0 is a perfect diffuse material (no mirror at al).
	Projection      *ImageProjection `json:"Projection,omitempty"`
	RefractionIndex float64          `json:"RefractionIndex,omitempty"`
	SolidObject     bool             `json:"SolidObject,omitempty"`   // SolidObject is if the material denotes a solid object with volume, not a hollow or open object or object nor an object with plane-thin walls. Solid transparent objects can refract light, hollow objects don't.
	Transparency    float64          `json:"Transparency,omitempty"`  // Transparency is the amount [0,1.0) of transparency vs diffuse contribution.
	RayTerminator   bool             `json:"RayTerminator,omitempty"` // RayTerminator decide if the ray should terminate after hit with object. Example can be an environment sphere or environment cube where a hit to the wall is the same as "no hit, continue in infinity". Extremely bright lights can also be ray terminators, their appearance will not notably be affected by further tracing.
}

// NewMaterial creates a new material with sensible defaults.
func NewMaterial() *Material {
	return &Material{
		Name:            "",
		Color:           &color.White,
		Diffuse:         1.0,
		Emission:        &color.Black,
		Glossiness:      0.0,
		Roughness:       1.0,
		Projection:      nil,
		RefractionIndex: RefractionIndex_Air,
		SolidObject:     false,
		Transparency:    0.0,
		RayTerminator:   false,
	}
}

// Copy copies the material. Any reference to other objects (color or projection)
// in the old material will be the same reference in the new material.
func (m *Material) Copy() *Material {
	newMaterial := *m
	return &newMaterial
}

// N is name properties
func (m *Material) N(name string) *Material {
	m.Name = name
	return m
}

// C is color properties
func (m *Material) C(color color.Color) *Material {
	m.Color = &color
	return m
}

// E is emission properties
func (m *Material) E(emission color.Color, scale float64, rayTerminator bool) *Material {
	m.Emission = (&emission).Multiply(float32(scale))
	m.RayTerminator = rayTerminator
	return m
}

// M is metallic properties
func (m *Material) M(glossiness float64, roughness float64) *Material {
	m.Roughness = roughness
	m.Glossiness = glossiness
	m.Diffuse = 1.0 - math.Min(m.Glossiness+m.Transparency, 1.0)
	return m
}

// T is transparency properties
func (m *Material) T(transparency float64, solidObject bool, refractionIndex float64) *Material {
	m.Transparency = transparency
	m.SolidObject = solidObject
	m.RefractionIndex = refractionIndex
	m.Diffuse = 1.0 - math.Min(m.Glossiness+m.Transparency, 1.0)
	return m
}

// P is projection properties
func (m *Material) P(projection *ImageProjection) *Material {
	m.Projection = projection
	return m
}

// PP is parallel projection properties
func (m *Material) PP(textureFilename string, origin *vec3.T, u vec3.T, v vec3.T) *Material {
	parallelImageProjection := NewParallelImageProjection(textureFilename, origin, u, v)
	m.Projection = &parallelImageProjection
	return m
}

// SP is spherical projection (of equirectangular images) properties
func (m *Material) SP(textureFilename string, origin *vec3.T, u vec3.T, v vec3.T) *Material {
	sphericalImageProjection := NewSphericalImageProjection(textureFilename, origin, u, v)
	m.Projection = &sphericalImageProjection
	return m
}

// TP is texture projection
func (m *Material) TP(textureFilename string) *Material {
	textureMappingImageProjection := NewTextureMappingImageProjection(textureFilename)
	m.Projection = &textureMappingImageProjection
	return m
}

// CP is cylindrical projection properties
func (m *Material) CP(textureFilename string, origin *vec3.T, u vec3.T, v vec3.T, repeat bool) *Material {
	sphericalImageProjection := NewCylindricalImageProjection(textureFilename, origin, u, v)
	sphericalImageProjection.RepeatV = repeat
	m.Projection = &sphericalImageProjection
	return m
}
