package scene

import (
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"

	"github.com/ungerik/go3d/float64/vec3"
)

type Material struct {
	Name                 string
	Color                *color.Color
	Diffuse              float64
	Emission             *color.Color
	Glossiness           float64 // Glossiness is the percent amount that will make out specular reflection. Values [0.0, 1.0] with default 0.0. The lower value, the more diffuse color will appear, and the higher value, the more mirror reflection will appear.
	Roughness            float64 // Roughness is the diffuse spread of the glossiness (specular reflection). Values [0.0, 1.0] with default 1.0. Higher value is like "brushed metal" or "foggy/hazy reflection" and a lower value gives a more mirror like reflection. A value of 0.0 is perfect mirror reflection and a value of 1.0 is a perfect diffuse material (no mirror at all).
	Projection           *ImageProjection
	RefractionIndex      float64
	SolidObject          bool    // SolidObject is if the material denotes a solid object with volume, not a hollow or open object or object nor an object with plane-thin walls. Solid transparent objects can refract light, hollow objects don't.
	Transparency         float64 // Transparency is the amount [0,1.0) of transparency vs diffuse contribution.
	RayTerminator        bool    // RayTerminator decide if the ray should terminate after hit with the object. Example can be an environment sphere or environment cube where a hit to the wall is the same as "no hit, continue in infinity". Extremely bright lights can also be ray terminators, their appearance will not notably be affected by further tracing.
	ColorizeReflection   bool    // ColorizeReflection is if the reflection color should be colorized based on the material color. Shiny metal materials will colorize reflection. Dielectric materials will not colorize reflection.
	FresnelMaxGlossiness float64
}

// NewMaterial creates a new material with sensible defaults.
func NewMaterial() *Material {
	return &Material{
		Name:                 "",
		Color:                color.White,
		Diffuse:              1.0,
		Emission:             nil,
		Glossiness:           0.0, // No mirror reflection (specularity)
		Roughness:            1.0, // Full roughness by default (diffuse spread of specularity)
		Projection:           nil,
		RefractionIndex:      RefractionIndex_Air,
		SolidObject:          false,
		Transparency:         0.0,
		RayTerminator:        false,
		ColorizeReflection:   false,
		FresnelMaxGlossiness: 1.0,
	}
}

// Copy copies the material.
// Any reference to other property objects (colors or projection)
// in the old material will be the same reference in the new material.
func (m *Material) Copy() *Material {
	newMaterial := *m
	return &newMaterial
}

// N sets the material name
func (m *Material) N(name string) *Material {
	m.Name = name
	return m
}

// C is color properties
func (m *Material) C(c *color.Color) *Material {
	m.Color = c
	return m
}

// E is emission properties
func (m *Material) E(emission *color.Color, scale float64, rayTerminator bool) *Material {
	m.Emission = emission
	if scale != 1.0 && emission != nil {
		m.Emission = emission.Copy().Multiply(float32(scale))
	}
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

// PP is a parallel projection property
func (m *Material) PP(texture *floatimage.FloatImage, origin *vec3.T, u vec3.T, v vec3.T) *Material {
	parallelImageProjection := NewParallelImageProjection(texture, origin, u, v)
	m.Projection = &parallelImageProjection
	return m
}

// SP is a spherical projection (of equirectangular images) property
//
// u and v vectors should be orthogonal to one another.
//
// u vector points to the middle of the sphere where the texture left edge starts.
// The intersection point on the sphere (of the u vector) corresponds to the pixel
// in the middle of the left most pixel column in the texture.
//
// v vector is the "up" vector of the projection.
// The intersection point on the sphere (of the v vector) corresponds to the pixel row at the top of the texture.
func (m *Material) SP(texture *floatimage.FloatImage, origin *vec3.T, u vec3.T, v vec3.T) *Material {
	sphericalImageProjection := NewSphericalImageProjection(texture, origin, u, v)
	m.Projection = &sphericalImageProjection
	return m
}

// CP is a cylindrical projection property
func (m *Material) CP(texture *floatimage.FloatImage, origin *vec3.T, u vec3.T, v vec3.T, repeat bool) *Material {
	sphericalImageProjection := NewCylindricalImageProjection(texture, origin, u, v)
	sphericalImageProjection.RepeatV = repeat
	m.Projection = &sphericalImageProjection
	return m
}

func NewMaterialGlass(name string) *Material {
	return NewMaterial().
		N(name).
		C(color.NewColor(0.90, 0.92, 0.95)).
		M(0.270, 0.030).
		T(0.700, true, RefractionIndex_Glass)
}
