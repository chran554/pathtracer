package scene

import (
	"github.com/ungerik/go3d/mat3"
	"github.com/ungerik/go3d/vec3"
)

var (
	Black = Color{R: 0, G: 0, B: 0}
)

type Color struct{ R, G, B float32 }

type Frame struct {
	Filename   string
	FrameIndex int
	Scene      Scene
}

type Animation struct {
	AnimationName string
	Frames        []Frame
	Width         int
	Height        int
}

type Scene struct {
	Camera  Camera
	Spheres []Sphere
	Discs   []Disc
}

type Material struct {
	Color    Color
	Emission Color
}

type Sphere struct {
	Origin vec3.T
	Radius float32

	Material Material
}

type Camera struct {
	Origin            vec3.T
	Heading           vec3.T
	ViewUp            vec3.T
	ViewPlaneDistance float32
	_coordinateSystem mat3.T
	LensRadius        float32
	FocalDistance     float32
	Samples           int
	AntiAlias         bool
}
type Line struct {
	Origin  vec3.T
	Heading vec3.T
}

type Ray Line

type Plane struct {
	Origin vec3.T
	Normal vec3.T

	Material Material
}

type Disc struct {
	Origin vec3.T
	Normal vec3.T
	Radius float32

	Material Material
}

type triangle struct {
	p1      vec3.T
	p2      vec3.T
	p3      vec3.T
	_normal vec3.T

	material Material
}

func (camera *Camera) GetCameraCoordinateSystem() mat3.T {
	if camera._coordinateSystem == (mat3.T{}) {
		heading := camera.Heading.Normalized()

		cameraX := vec3.Cross(&camera.ViewUp, &heading)
		cameraX.Normalize()
		cameraY := vec3.Cross(&heading, &cameraX)
		cameraY.Normalize()

		camera._coordinateSystem = mat3.T{cameraX, cameraY, heading}
	}
	return camera._coordinateSystem
}

func (c *Color) Add(color Color) {
	c.R += color.R
	c.G += color.G
	c.B += color.B
}

func (c *Color) Divide(divider float32) {
	c.R /= divider
	c.G /= divider
	c.B /= divider
}
