package main

import (
	"fmt"
	"math"
	"os"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "tessellated_sphere"

var environmentEnvironMap = "textures/equirectangular/sunset horizon 2800x1400.jpg"
var environmentRadius = 500.0 * 1000.0
var environmentEmissionFactor = float32(1.0)

var amountFrames = 180

var imageWidth = 300
var imageHeight = 300
var magnification = 3.0

var renderType = scn.Pathtracing
var amountSamples = 1000 * 24
var maxRecursion = 2

var viewPlaneDistance = 800.0
var apertureSize = 1.5

func main() {
	animation := getAnimation(int(float64(imageWidth)*magnification), int(float64(imageHeight)*magnification))

	// Sphere
	tessellatedSphere := obj.NewTessellatedSphere(5, false)

	/*
		objFile := createFile("tessellated_sphere_4.obj")
		defer objFile.Close()
		mtlFile := createFile("tessellated_sphere_4.mtl")
		defer mtlFile.Close()
		obj.WriteObjFile(objFile, mtlFile, tessellatedSphere, []string{"Tessellated sphere level 4, with vertex normals"})
		os.Exit(0)
	*/

	tessellatedSphere.Translate(&vec3.T{0, -tessellatedSphere.Bounds.Ymin, 0})
	tessellatedSphere.ScaleUniform(&vec3.Zero, 30.0)
	tessellatedSphereBounds := tessellatedSphere.UpdateBounds()
	tessellatedSphere.Material = scn.NewMaterial().N("tessellated sphere").C(color.White, 1.0)

	// Sky dome
	environmentSphere := addEnvironmentMapping(environmentEnvironMap)

	// Ground
	groundProjection := scn.NewParallelImageProjection("textures/ground/soil-cracked.png", vec3.Zero, vec3.UnitX.Scaled(150), vec3.UnitZ.Scaled(150))
	ground := scn.Disc{
		Name:     "Ground",
		Origin:   &vec3.Zero,
		Normal:   &vec3.UnitY,
		Radius:   environmentRadius,
		Material: &scn.Material{Name: "Ground material", Color: &color.White, Emission: &color.Black, Glossiness: 0.0, Roughness: 1.0, Projection: &groundProjection},
	}

	// Camera
	cameraOrigin := tessellatedSphereBounds.Center().Add(&vec3.T{25, 25, -250})
	cameraFocusPoint := tessellatedSphereBounds.Center().Add(&vec3.T{0, 0, -(tessellatedSphereBounds.SizeZ() / 2) * 0.9})
	//cameraFocusPoint := gopherBounds.Center().Add(&vec3.T{0.0, gopherBounds.SizeY() * 2, 0.0})
	camera := scn.NewCamera(cameraOrigin, cameraFocusPoint).S(amountSamples).D(maxRecursion).A(apertureSize, "").M(magnification)

	//for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
	for frameIndex := 88; frameIndex < 89; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		// Lamp
		lamp := &scn.Sphere{
			Name:     "lamp",
			Origin:   &vec3.T{50, 50, -35},
			Radius:   4,
			Material: scn.NewMaterial().N("lamp").C(color.White, 1.0).E(color.White, 125.0, true),
		}
		lamp.RotateY(tessellatedSphere.Bounds.Center(), -math.Pi/4+animationProgress*math.Pi/2)

		scene := scn.SceneNode{
			Spheres:         []*scn.Sphere{&environmentSphere, lamp},
			Discs:           []*scn.Disc{&ground},
			ChildNodes:      []*scn.SceneNode{},
			FacetStructures: []*scn.FacetStructure{tessellatedSphere},
		}

		frame := scn.Frame{
			Filename:   animation.AnimationName + "_" + fmt.Sprintf("%06d", frameIndex),
			FrameIndex: frameIndex,
			Camera:     camera,
			SceneNode:  &scene,
		}

		animation.Frames = append(animation.Frames, frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func addEnvironmentMapping(filename string) scn.Sphere {
	origin := vec3.T{0, 0, 0}

	projection := scn.ImageProjection{
		ProjectionType: scn.ProjectionTypeSpherical,
		ImageFilename:  filename,
		Origin:         &origin,
		U:              &vec3.T{-0.2, 0, -1},
		V:              &vec3.T{0, 1, 0},
		RepeatU:        true,
		RepeatV:        true,
		FlipU:          false,
		FlipV:          false,
	}

	material := scn.Material{
		Color:         &color.Color{R: 1.0, G: 1.0, B: 1.0},
		Emission:      (&color.Color{R: 1.0, G: 1.0, B: 1.0}).Multiply(environmentEmissionFactor),
		RayTerminator: true,
		Projection:    &projection,
	}

	sphere := scn.Sphere{
		Name:     "Environment mapping",
		Origin:   &origin,
		Radius:   environmentRadius,
		Material: &material,
	}

	return sphere
}

func getAnimation(width int, height int) *scn.Animation {
	animation := &scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             width,
		Height:            height,
		WriteRawImageFile: false,
	}
	return animation
}

func createFile(name string) *os.File {
	objFile, err := os.Create(name)
	if err != nil {
		fmt.Printf("could not create file: '%s'\n%s\n", objFile.Name(), err.Error())
		os.Exit(1)
	}
	return objFile
}
