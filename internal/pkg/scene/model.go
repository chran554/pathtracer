package scene

import (
	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
	_ "image/jpeg"
	_ "image/png"
	"pathtracer/internal/pkg/color"
)

// RenderType is the type used to define different render types
type RenderType string

const (
	// Pathtracing render type is used in camera settings to denote the path tracing algorithm to be used rendering the frame
	Pathtracing RenderType = "Pathtracing"
	// Raycasting render type is used in camera settings to denote the cheap and simple ray casting algorithm to be used rendering the frame
	Raycasting RenderType = "Raycasting"
)

type Ray struct {
	Origin          *vec3.T
	Heading         *vec3.T
	RefractionIndex float64
}

type Animation struct {
	AnimationName     string
	Frames            []Frame
	Width             int
	Height            int
	WriteRawImageFile bool
}

type Material struct {
	Name            string           `json:"Name,omitempty"`
	Color           *color.Color     `json:"Color,omitempty"`
	Emission        *color.Color     `json:"Emission,omitempty"`
	Glossiness      float32          `json:"Glossiness,omitempty"` // Glossiness is the percent amount that will make out specular reflection. Values [0.0 .. 1.0] with default 0.0. Lower value the more diffuse color will appear and higher value the more mirror reflection will appear.
	Roughness       float32          `json:"Roughness,omitempty"`  // Roughness is the diffuse spread of the specular reflection. Values [0.0 .. 1.0] with default 0.0. Lower is like "brushed metal" or "foggy/hazy reflection" and higher value give a more mirror like reflection. A value of 0.0 is perfect mirror reflection and a value of 0.0 is a perfect diffuse material (no mirror at al).
	Projection      *ImageProjection `json:"Projection,omitempty"`
	RefractionIndex float64          `json:"RefractionIndex,omitempty"`
	Transparency    float64          `json:"Transparency,omitempty"`
	RayTerminator   bool             `json:"RayTerminator,omitempty"`
}

type Camera struct {
	Origin            *vec3.T
	Heading           *vec3.T
	ViewUp            *vec3.T
	ViewPlaneDistance float64
	_coordinateSystem *mat3.T
	ApertureSize      float64 // ApertureSize is the size of the aperture opening. The wider the aperture the less focus depth. Value 0.0 is infinite focus depth.
	ApertureShape     string  // ApertureShape file path to a black and white image where white define the aperture shape. Aperture size determine the size of the longest side (width or height) of the image. If nil then a default round aperture shape is used.
	FocusDistance     float64
	Samples           int
	AntiAlias         bool
	Magnification     float64
	RenderType        RenderType
	RecursionDepth    int
}

type Frame struct {
	Filename   string
	FrameIndex int
	Camera     *Camera
	SceneNode  *SceneNode
}

type Plane struct {
	Name     string
	Origin   *vec3.T
	Normal   *vec3.T
	Material *Material `json:"Material,omitempty"`
}

func RotateY(point *vec3.T, rotationOrigin *vec3.T, angle float64) {
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(angle)

	origin := *point
	origin.Sub(rotationOrigin)
	origin[2] *= -1 // Change to right hand coordinate system from left hand coordinate system
	rotatedOrigin := rotationMatrix.MulVec3(&origin)
	rotatedOrigin[2] *= -1 // Change back from right hand coordinate system to left hand coordinate system
	rotatedOrigin.Add(rotationOrigin)

	point[0] = rotatedOrigin[0]
	point[1] = rotatedOrigin[1]
	point[2] = rotatedOrigin[2]
}

func (r Ray) point(t float64) *vec3.T {
	return &vec3.T{
		r.Origin[0] + r.Heading[0]*t,
		r.Origin[1] + r.Heading[1]*t,
		r.Origin[2] + r.Heading[2]*t,
	}
}

func NewPlane(v1, v2, v3 *vec3.T, name string, material *Material) *Plane {
	a := v2.Subed(v1)
	b := v3.Subed(v1)
	n := vec3.Cross(&a, &b)
	n.Normalize()

	return &Plane{
		Name:     name,
		Origin:   v1,
		Normal:   &n,
		Material: material,
	}
}
