package main

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"os"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
)

var animationName = "gopher"

var renderType = scn.Pathtracing

var amountFrames = 1

var maxRecursionDepth = 3
var amountSamples = 1024 * 4 * 3
var apertureRadius = 0.0

var viewPlaneDistance = 800.0

var imageWidth = 400
var imageHeight = 500
var magnification = 1.0

func main() {
	width := int(float64(imageWidth) * magnification)
	height := int(float64(imageHeight) * magnification)

	// Keep image proportions to an even amount of pixel for mp4 encoding
	if width%2 == 1 {
		width++
	}
	if height%2 == 1 {
		height++
	}

	// Sky
	groundProjection := scn.NewParallelImageProjection("textures/floor/Calacatta-Vena-French-Pattern-Architextures.jpg", vec3.Zero, vec3.UnitX.Scaled(150), vec3.UnitZ.Scaled(150))
	groundMaterial := scn.Material{Color: &color.White, Glossiness: 0.0, Roughness: 1.0, Projection: &groundProjection}
	ground := &scn.Disc{Name: "ground", Origin: &vec3.Zero, Normal: &vec3.UnitY, Radius: 5000.0, Material: &groundMaterial}

	// Sky
	skyProjection := scn.NewSphericalImageProjection("textures/equirectangular/wirebox 6192x3098.png", vec3.Zero, vec3.UnitX, vec3.UnitY)
	skyMaterial := scn.Material{Color: &color.White, Emission: (&color.White).Multiply(0.5), Glossiness: 0.0, Roughness: 1.0, Projection: &skyProjection}
	skyDome := &scn.Sphere{Name: "sky dome", Origin: &vec3.Zero, Radius: 5000.0, Material: &skyMaterial}

	// Gopher
	gopher := GetGopher(&vec3.T{1, 1, 1})
	gopher.Translate(&vec3.T{0, -gopher.Bounds.Ymin, 0})
	gopher.ScaleUniform(&vec3.Zero, 2.0)
	gopher.RotateY(&vec3.Zero, math.Pi*5.0/6.0)
	gopher.Translate(&vec3.T{0, 0, 0})
	gopher.UpdateBounds()

	gopherLightMaterial := scn.Material{Color: &color.White, Emission: (&color.Color{R: 6.0, G: 5.3, B: 4.5}).Multiply(20), Glossiness: 0.0, Roughness: 1.0, RayTerminator: true}
	gopherLight := &scn.Sphere{Name: "Gopher light", Origin: &vec3.T{-150, 250, -175}, Radius: 15.0, Material: &gopherLightMaterial}

	scene := &scn.SceneNode{
		FacetStructures: []*scn.FacetStructure{gopher},
		Spheres:         []*scn.Sphere{gopherLight, skyDome},
		Discs:           []*scn.Disc{ground},
	}

	animation := &scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             width,
		Height:            height,
		WriteRawImageFile: false,
	}

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		camera := getCamera(animationProgress)

		frame := scn.Frame{
			Filename:   animationName + "_" + fmt.Sprintf("%06d", frameIndex),
			FrameIndex: 0,
			Camera:     camera,
			SceneNode:  scene,
		}

		animation.Frames = append(animation.Frames, frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func getCamera(animationProgress float64) *scn.Camera {
	cameraOrigin := vec3.T{0, 200, -800}
	focusPoint := vec3.T{0, 150, 0}

	// Animation
	angle := (math.Pi / 2.0) * animationProgress
	scn.RotateY(&cameraOrigin, &vec3.Zero, angle)
	scn.RotateY(&focusPoint, &vec3.Zero, angle)

	heading := focusPoint.Subed(&cameraOrigin)
	focalDistance := heading.Length() * 1.75

	return &scn.Camera{
		Origin:            &cameraOrigin,
		Heading:           &heading,
		ViewUp:            &vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		ApertureSize:      apertureRadius,
		FocusDistance:     focalDistance,
		Samples:           amountSamples,
		AntiAlias:         true,
		Magnification:     magnification,
		RenderType:        renderType,
		RecursionDepth:    maxRecursionDepth,
	}
}

func GetGopher(scale *vec3.T) *scn.FacetStructure {
	var objFilename = "go_gopher_color.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objFilename

	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objFilenamePath, err.Error())
	}
	defer objFile.Close()

	object, err := obj.Read(objFile)
	object.Scale(&vec3.Zero, scale)
	// obj.ClearMaterials()
	object.UpdateBounds()
	fmt.Printf("Gopher bounds: %+v\n", object.Bounds)

	return object
}
