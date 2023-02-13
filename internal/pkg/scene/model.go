package scene

import (
	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
	_ "image/jpeg"
	_ "image/png"
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
	Origin  *vec3.T
	Heading *vec3.T
}

type Animation struct {
	AnimationName     string
	Frames            []Frame
	Width             int
	Height            int
	WriteRawImageFile bool
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
