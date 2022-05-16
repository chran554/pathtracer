package scene

import (
	_ "image/jpeg"
	_ "image/png"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/image"

	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
)

type ProjectionType string

const (
	Parallel    ProjectionType = "Parallel"
	Cylindrical ProjectionType = "Cylindrical"
	Spherical   ProjectionType = "Spherical"
)

type RenderType string

const (
	Pathtracing RenderType = "Pathtracing"
	Raycasting  RenderType = "Raycasting"
)

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

type SceneNode struct {
	Spheres    []*Sphere
	Discs      []*Disc
	ChildNodes []*SceneNode
	Bounds     *Bounds
}

func (sn *SceneNode) GetSpheres() []*Sphere {
	return sn.Spheres
}

func (sn *SceneNode) GetAmountSpheres() int {
	amountSpheres := len(sn.Spheres)
	for _, node := range sn.GetChildNodes() {
		amountSpheres += node.GetAmountSpheres()
	}
	return amountSpheres
}

func (sn *SceneNode) GetDiscs() []*Disc {
	return sn.Discs
}

func (sn *SceneNode) GetAmountDiscs() int {
	amountDiscs := len(sn.Discs)
	for _, node := range sn.GetChildNodes() {
		amountDiscs += node.GetAmountDiscs()
	}
	return amountDiscs
}

func (sn *SceneNode) Clear() {
	sn.Spheres = nil
	sn.Discs = nil

	for _, node := range sn.GetChildNodes() {
		node.Clear()
	}
}

func (sn *SceneNode) GetChildNodes() []*SceneNode {
	return sn.ChildNodes
}

//func (sn *SceneNode) GetParentNode() *SceneNode {
//	return sn.ParentNode
//}

func (sn *SceneNode) GetBounds() *Bounds {
	return sn.Bounds
}

type Material struct {
	Color           color.Color
	Emission        *color.Color `json:"Emission,omitempty"`
	Glossiness      float64
	Projection      *ImageProjection `json:"Projection,omitempty"`
	RefractionIndex float64
	Transparancy    float64
	RayTerminator   bool
}

type ImageProjection struct {
	ProjectionType                  ProjectionType `json:"ProjectionType"`
	ImageFilename                   string         `json:"ImageFilename"`
	_image                          *image.FloatImage
	_invertedCoordinateSystemMatrix mat3.T
	Origin                          vec3.T  `json:"Origin"`
	U                               vec3.T  `json:"U"`
	V                               vec3.T  `json:"V"`
	RepeatU                         bool    `json:"RepeatU,omitempty"`
	RepeatV                         bool    `json:"RepeatV,omitempty"`
	FlipU                           bool    `json:"FlipU,omitempty"`
	FlipV                           bool    `json:"FlipV,omitempty"`
	Gamma                           float64 `json:"Gamma,omitempty"`
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
	RenderType        RenderType
	RecursionDepth    int
}

type Ray struct {
	Origin          vec3.T
	Heading         vec3.T
	RefractionIndex float64
}

func (r Ray) point(t float64) *vec3.T {
	return &vec3.T{
		r.Origin[0] + r.Heading[0]*t,
		r.Origin[1] + r.Heading[1]*t,
		r.Origin[2] + r.Heading[2]*t,
	}
}

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

func (sphere Sphere) Initialize() {
	projection := sphere.Material.Projection
	if projection != nil {
		projection.Initialize()
	}
}

func (disc *Disc) Initialize() {
	disc.Normal.Normalize()

	projection := disc.Material.Projection
	if projection != nil {
		projection.Initialize()
	}
}
