package scene

import (
	"math"
	"pathtracer/internal/pkg/color"

	"github.com/ungerik/go3d/float64/vec3"
)

type Material struct {
	Name            string
	Color           *color.Color // Linear RGB base color for dielectric diffuse albedo and metal reflection F0. Alpha scales emission.
	Emission        *color.Color // Emitted RGB is tinted by the surface albedo and scaled by Emission.A * Color.A.
	Metallicness    float64      // [0,1]: dielectric (0) to conductor (1).
	Roughness       float64      // Perceptual GGX roughness in [0,1] for specular reflection and transmission through solids or thin walls. Zero gives a smooth interface.
	Projections     []*ImageProjection
	RefractionIndex float64
	SolidObject     bool         // Solid transmission crosses a volume boundary. Non-solid transmission models two faces of a thin wall within the same scene medium.
	Transparency    float64      // Transmission fraction of the dielectric portion in [0,1].
	Absorption      *color.Color // Homogeneous interior RGB absorption coefficients in inverse scene units (sigma_a >= 0, not limited to 1). Nil means no absorption; alpha is ignored. Used while this material is the active medium, independently of surface Color.
	RayTerminator   bool         // RayTerminator decide if the ray should terminate after hit with the object. Example can be an environment sphere or environment cube where a hit to the wall is the same as "no hit, continue in infinity". Extremely bright lights can also be ray terminators, their appearance will not notably be affected by further tracing.
}

// NewMaterial creates an opaque, rough white dielectric with IOR 1.5.
func NewMaterial() *Material {
	return &Material{
		Name:            "",
		Color:           color.White,
		Emission:        nil,
		Metallicness:    0.0, // Dielectric surface
		Roughness:       1.0, // Full roughness by default (diffuse spread of specularity)
		Projections:     nil,
		RefractionIndex: 1.5,
		SolidObject:     false,
		Transparency:    0.0,
		Absorption:      nil,
		RayTerminator:   false,
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

// C sets the base color. Its RGB is tinted by the albedo texture and its alpha scales emission.
func (m *Material) C(c *color.Color) *Material {
	m.Color = c
	return m
}

// E sets the emission color and scales its RGB by scale. At a hit, emission alpha
// and base color alpha both scale the emitted light. rayTerminator ends the path at the hit.
func (m *Material) E(emission *color.Color, scale float64, rayTerminator bool) *Material {
	m.Emission = emission
	if scale != 1.0 && emission != nil {
		m.Emission = emission.Copy().Multiply(float32(scale))
	}
	m.RayTerminator = rayTerminator
	return m
}

// M sets metallicness and perceptual GGX roughness, both in [0,1].
func (m *Material) M(metallicness float64, roughness float64) *Material {
	m.Roughness = materialUnitInterval(roughness)
	m.Metallicness = materialUnitInterval(metallicness)
	return m
}

// T sets transmission within the dielectric portion, the boundary type, and IOR.
func (m *Material) T(transparency float64, solidObject bool, refractionIndex float64) *Material {
	m.Transparency = materialUnitInterval(transparency)
	m.SolidObject = solidObject
	m.RefractionIndex = refractionIndex
	return m
}

// Absorb sets the interior absorption coefficients in inverse scene units.
//
// Each RGB coefficient must be finite and nonnegative.
// absorption value of nil disables absorption.
// Closed objects need SolidObject and optical transmission to enter the medium.
// Setting absorption alone does not make a surface transparent.
func (m *Material) Absorb(absorption *color.Color) *Material {
	if absorption != nil {
		for _, coefficient := range []float32{absorption.R, absorption.G, absorption.B} {
			if !(coefficient >= 0) || math.IsInf(float64(coefficient), 0) {
				panic("absorption coefficients must be finite and nonnegative")
			}
		}
	}
	m.Absorption = absorption
	return m
}

func ClampAttenuateColor(c *color.Color) *color.Color {
	c.R = max(0.0000001, c.R)
	c.G = max(0.0000001, c.G)
	c.B = max(0.0000001, c.B)
	return c
}

// Attenuate specifies the linear RGB transmission through a reference distance
// of the homogeneous interior, excluding surface reflection and refraction.
//
// transmittance value of nil disables absorption.
// RGB values must be in (0,1] and distance must be finite and positive.
//
// For example, (0.5, 1, 1) at distance 2 halves red every 2 scene units.
//
// Use near-zero transmission for very strong absorption,
// or disable Transparency for an opaque surface.
func (m *Material) Attenuate(transmittance *color.Color, distance float64) *Material {
	if transmittance == nil {
		m.Absorption = nil
		return m
	}
	if !(distance > 0) || math.IsInf(distance, 0) {
		panic("attenuation requires a transmission color and a finite positive distance")
	}
	for _, channel := range []float32{transmittance.R, transmittance.G, transmittance.B} {
		if (channel <= 0) || (channel > 1) {
			panic("attenuation transmission channels must be in (0,1]")
		}
	}
	return m.Absorb(color.NewColor(
		-math.Log(float64(transmittance.R))/distance,
		-math.Log(float64(transmittance.G))/distance,
		-math.Log(float64(transmittance.B))/distance,
	))
}

// P is projection properties
func (m *Material) P(projection *ImageProjection) *Material {
	if m.Projections == nil {
		m.Projections = make([]*ImageProjection, 0)
	}
	m.Projections = append(m.Projections, projection)
	return m
}

// PP is a parallel projection property
func (m *Material) PP(texture *Texture, origin *vec3.T, u vec3.T, v vec3.T) *Material {
	m.Projections = append(m.Projections, NewParallelImageProjection(texture, origin, u, v))
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
func (m *Material) SP(texture *Texture, origin *vec3.T, u vec3.T, v vec3.T) *Material {
	m.Projections = append(m.Projections, NewSphericalImageProjection(texture, origin, u, v))
	return m
}

// CP is a cylindrical projection property
func (m *Material) CP(texture *Texture, origin *vec3.T, u vec3.T, v vec3.T, repeat bool) *Material {
	sphericalImageProjection := NewCylindricalImageProjection(texture, origin, u, v)
	sphericalImageProjection.RepeatV = repeat
	m.Projections = append(m.Projections, sphericalImageProjection)
	return m
}

// ProjectionColor return the image projection color for a texture of matching type.
//
// Return value may be nil as no Projections might be defined for the material,
// no matching texture of requested type is available,
// or no valid projected color is defined for the point
// (i.e. the point is outside the projected area of the texture).
func (m *Material) ProjectionColor(point *vec3.T, textureType TextureType) *color.Color {
	var c *color.Color
	for _, projection := range m.Projections {
		if projection.Texture == nil {
			continue
		}

		if projection.Texture.Type == textureType {
			col := projection.GetColor(point)

			if col != nil {
				if c == nil {
					c = col.Copy()
				} else {
					c.Over(col)
				}
			}
		}
	}

	return c
}

func (m *Material) ProjectionsOfType(textureType TextureType) []*ImageProjection {
	var projections []*ImageProjection
	for _, projection := range m.Projections {
		if projection.Texture != nil && projection.Texture.Type == textureType {
			projections = append(projections, projection)
		}
	}
	return projections
}

func NewMaterialSolidGlass(name string) *Material {
	return NewMaterial().
		N(name).
		C(color.NewColor(0.90, 0.92, 0.95)).
		M(0, 0.030).
		T(1, true, RefractionIndex_Glass)
}

// materialUnitInterval validates authored surface parameters. Public fields are
// clamped again by the renderer so deserialized/directly assigned values are safe.
func materialUnitInterval(value float64) float64 {
	if !(value >= 0 && value <= 1) {
		panic("material parameter must be finite and in [0,1]")
	}
	return value
}
