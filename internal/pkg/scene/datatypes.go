package scene

import (
	_ "image/jpeg"
	_ "image/png"

	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
)

var (
	Black = Color{R: 0, G: 0, B: 0}
	White = Color{R: 1, G: 1, B: 1}
)

type ProjectionType string

const (
	Parallel    ProjectionType = "Parallel"
	Cylindrical ProjectionType = "Cylindrical"
)

type Color struct{ R, G, B float64 }

type Frame struct {
	Filename   string
	FrameIndex int
	Scene      Scene
}

type Animation struct {
	AnimationName     string
	Frames            []Frame
	Width             int
	Height            int
	WriteRawImageFile bool
}

type Scene struct {
	Camera  Camera
	Spheres []Sphere
	Discs   []Disc
}

type Material struct {
	Color      Color
	Emission   *Color           `json:"Emission,omitempty"`
	Projection *ImageProjection `json:"Projection,omitempty"`
}

type ImageProjection struct {
	ProjectionType                  ProjectionType `json:"ProjectionType"`
	ImageFilename                   string         `json:"ImageFilename"`
	_imageData                      []Color
	_imageWidth                     int
	_imageHeight                    int
	_invertedCoordinateSystemMatrix mat3.T
	Origin                          vec3.T `json:"Origin"`
	U                               vec3.T `json:"U"`
	V                               vec3.T `json:"V"`
	RepeatU                         bool   `json:"RepeatU,omitempty"`
	RepeatV                         bool   `json:"RepeatV,omitempty"`
	FlipU                           bool   `json:"FlipU,omitempty"`
	FlipV                           bool   `json:"FlipV,omitempty"`
}

type Sphere struct {
	Name   string
	Origin vec3.T
	Radius float64

	Material Material
}

type Camera struct {
	Origin            vec3.T
	Heading           vec3.T
	ViewUp            vec3.T
	ViewPlaneDistance float64
	_coordinateSystem mat3.T
	LensRadius        float64
	FocalDistance     float64
	Samples           int
	AntiAlias         bool
	Magnification     float64
}
type Line struct {
	Origin  vec3.T
	Heading vec3.T
}

type Ray Line

type Plane struct {
	Name   string
	Origin vec3.T
	Normal vec3.T

	Material Material
}

type Disc struct {
	Name   string
	Origin vec3.T
	Normal vec3.T
	Radius float64

	Material Material
}

/*
type triangle struct {
	p1      vec3.T
	p2      vec3.T
	p3      vec3.T
	_normal vec3.T

	material Material
}
*/
