package scene

import "pathtracer/internal/pkg/color"

const (
	RefractionIndex_Vacuum         = 1.0
	RefractionIndex_Air            = 1.000273
	RefractionIndex_Water          = 1.333
	RefractionIndex_Ice            = 1.31
	RefractionIndex_PyrexGlass     = 1.470
	RefractionIndex_AcrylicPlastic = 1.495
	RefractionIndex_Glass          = 1.50
	RefractionIndex_WindowGlass    = 1.52
	RefractionIndex_PetPlastic     = 1.5750
	RefractionIndex_Sapphire       = 1.762
	RefractionIndex_Diamond        = 2.417
)

type Material struct {
	Name            string           `json:"Name,omitempty"`
	Color           *color.Color     `json:"Color,omitempty"`
	Emission        *color.Color     `json:"Emission,omitempty"`
	Glossiness      float64          `json:"Glossiness,omitempty"` // Glossiness is the percent amount that will make out specular reflection. Values [0.0 .. 1.0] with default 0.0. Lower value the more diffuse color will appear and higher value the more mirror reflection will appear.
	Roughness       float64          `json:"Roughness,omitempty"`  // Roughness is the diffuse spread of the specular reflection. Values [0.0 .. 1.0] with default 0.0. Lower is like "brushed metal" or "foggy/hazy reflection" and higher value give a more mirror like reflection. A value of 0.0 is perfect mirror reflection and a value of 0.0 is a perfect diffuse material (no mirror at al).
	Projection      *ImageProjection `json:"Projection,omitempty"`
	RefractionIndex float64          `json:"RefractionIndex,omitempty"`
	SolidObject     bool             `json:"SolidObject,omitempty"`   // SolidObject is if the material denotes a solid object with volume, not a hollow or open object or object nor an object with plane-thin walls. Solid transparent objects can refract light, hollow objects don't.
	Transparency    float64          `json:"Transparency,omitempty"`  // Transparency is the amount [0,1.0) of transparency vs diffuse contribution.
	RayTerminator   bool             `json:"RayTerminator,omitempty"` // RayTerminator decide if the ray should terminate after hit with object. Example can be an environment sphere or environment cube where a hit to the wall is the same as "no hit, continue in infinity". Extremely bright lights can also be ray terminators, their appearance will not notably be affected by further tracing.
}

func NewMaterial() *Material {
	return &Material{
		Name:            "",
		Color:           &color.White,
		Emission:        &color.Black,
		Glossiness:      0.0,
		Roughness:       1.0,
		Projection:      nil,
		RefractionIndex: 0.0,
		SolidObject:     false,
		Transparency:    0.0,
		RayTerminator:   false,
	}
}

func (m *Material) C(color color.Color, scale float64) *Material {
	m.Color = (&color).Multiply(float32(scale))
	return m
}

func (m *Material) E(emission color.Color, scale float64, rayTerminator bool) *Material {
	m.Emission = (&emission).Multiply(float32(scale))
	m.RayTerminator = rayTerminator
	return m
}

func (m *Material) M(glossiness float64, roughness float64) *Material {
	m.Roughness = roughness
	m.Glossiness = glossiness
	return m
}

func (m *Material) T(transparency float64, solidObject bool, refractionIndex float64) *Material {
	m.Transparency = transparency
	m.SolidObject = solidObject
	m.RefractionIndex = refractionIndex
	return m
}

func (m *Material) P(projection *ImageProjection) *Material {
	m.Projection = projection
	return m
}
