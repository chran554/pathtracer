package scene

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"

	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
)

// RenderType is the type used to define different render types
type RenderType string

const (
	// Pathtracing render type is used in camera settings to denote the path tracing algorithm to be used rendering the frame
	Pathtracing RenderType = "Pathtracing"
	// Raycasting render type is used in camera settings to denote the cheap and simple ray casting algorithm to be used rendering the frame
	Raycasting RenderType = "Raycasting"
)

type ScreenResolution struct {
	name   string
	width  int
	height int
}

var (
	ScreenResolution_HD_854x480_WVGA     ScreenResolution = ScreenResolution{name: "WVGA (HD 16:9)", width: 854, height: 480}
	ScreenResolution_HD_1024x576_PAL     ScreenResolution = ScreenResolution{name: "PAL (HD 16:9)", width: 1024, height: 576}
	ScreenResolution_HD_1366x768         ScreenResolution = ScreenResolution{name: "1366x768 (HD 16:9)", width: 1366, height: 768}
	ScreenResolution_HD_1600x900         ScreenResolution = ScreenResolution{name: "1600x900 (HD 16:9)", width: 1600, height: 900}
	ScreenResolution_HD_1920x1080_FullHD ScreenResolution = ScreenResolution{name: "Full HD (HD 16:9)", width: 1920, height: 1080}
	ScreenResolution_HD_2560x1440_WQHD   ScreenResolution = ScreenResolution{name: "WQHD (HD 16:9)", width: 2560, height: 1440}
	ScreenResolution_HD_3840x2160_UHD1   ScreenResolution = ScreenResolution{name: "UHD-1 (HD 16:9)", width: 3840, height: 2160}
	ScreenResolution_HD_5120x2880_5K     ScreenResolution = ScreenResolution{name: "5K (HD 16:9)", width: 5120, height: 2880}
	ScreenResolution_HD_7680x4320_UHD2   ScreenResolution = ScreenResolution{name: "UHD-2 (HD 16:9)", width: 7680, height: 4320}
	ScreenResolution_VGA_320x240_QVGA    ScreenResolution = ScreenResolution{name: "QVGA (VGA 4:3)", width: 320, height: 240}
	ScreenResolution_VGA_640x480_VGA     ScreenResolution = ScreenResolution{name: "VGA (VGA 4:3)", width: 640, height: 480}
	ScreenResolution_VGA_800x600_SGA     ScreenResolution = ScreenResolution{name: "SVGA (VGA 4:3)", width: 800, height: 600}
	ScreenResolution_VGA_1024x768_XGA    ScreenResolution = ScreenResolution{name: "XGA (VGA 4:3)", width: 1024, height: 768}
	ScreenResolution_VGA_1400x1050_SXGA  ScreenResolution = ScreenResolution{name: "SXGA (VGA 4:3)", width: 1400, height: 1050}
	ScreenResolution_VGA_1600x1200_UXGA  ScreenResolution = ScreenResolution{name: "UXGA (VGA 4:3)", width: 1600, height: 1200}
	ScreenResolution_VGA_2048x1536_QXGA  ScreenResolution = ScreenResolution{name: "QXGA (VGA 4:3)", width: 2048, height: 1536}
)

type Ray struct {
	Origin  *vec3.T
	Heading *vec3.T
}

type Animation struct {
	AnimationName      string
	Frames             []*Frame
	Width              int
	Height             int
	WriteRawImageFile  bool
	WriteImageInfoFile bool
}

func NewAnimation(name string, pixelWidth int, pixelHeight int, magnification float64, rawFile bool, infoFile bool) *Animation {
	width := int(float64(pixelWidth) * magnification)
	height := int(float64(pixelHeight) * magnification)

	// Keep image proportions to an even amount of pixel for mp4 encoding
	if width%2 == 1 {
		width++
	}
	if height%2 == 1 {
		height++
	}

	return &Animation{
		AnimationName:      name,
		Width:              width,
		Height:             height,
		WriteRawImageFile:  rawFile,
		WriteImageInfoFile: infoFile,
	}
}

func (a *Animation) AddFrame(frame *Frame) *Animation {
	a.Frames = append(a.Frames, frame)
	return a
}

type Frame struct {
	Filename  string
	Index     int
	Camera    *Camera
	SceneNode *SceneNode
}

func NewFrame(fileName string, frameIndex int, camera *Camera, scene *SceneNode) *Frame {
	frameFileName := fileName
	if frameIndex != -1 {
		frameFileName += "_" + fmt.Sprintf("%06d", frameIndex)
	}

	return &Frame{
		Filename:  frameFileName,
		Index:     frameIndex,
		Camera:    camera,
		SceneNode: scene,
	}
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
