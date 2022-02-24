package main

import (
	"github.com/ungerik/go3d/mat3"
	"github.com/ungerik/go3d/vec3"
)

var (
	black = Color{R: 0, G: 0, B: 0}
)

type Color struct{ R, G, B float32 }

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

type Scene struct {
	Camera  Camera
	Spheres []Sphere
	Discs   []Disc
}

type line struct {
	origin  vec3.T
	heading vec3.T
}

type ray line

type plane struct {
	origin vec3.T
	normal vec3.T

	material Material
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
