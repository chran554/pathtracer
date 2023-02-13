package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var environmentEnvironMap = "textures/equirectangular/sunset horizon 2800x1400.jpg"
var environmentRadius = 100.0 * 1000.0
var environmentEmissionFactor = float32(2.0)

var animationName = "aperture_shape_test"
var amountFrames = 1

var imageWidth = 1280
var imageHeight = 1024
var magnification = 1.0

var renderType = scn.Pathtracing
var amountSamples = 512 * 6
var maxRecursion = 4

var gopherLightEmissionFactor float32 = 5

var lampEmissionFactor = 5.0

var amountSpheres = 400
var sphereRadius = 250.0

var sphereMinDistance = 1.0 * 1000.0
var sphereMaxDistance = environmentRadius * 0.75

var viewPlaneDistance = 800.0
var apertureSize = 20.0 //500.0

func main() {
	// Environment sphere

	environmentSphere := addEnvironmentMapping(environmentEnvironMap)

	// Gopher

	gopher := GetGopher(3.0)
	gopher.RotateY(&vec3.Zero, math.Pi*7.0/8.0)
	gopher.Translate(&vec3.T{0, -gopher.Bounds.Ymin, -gopher.Bounds.Zmin})
	//gopher.Translate(&vec3.T{0, 0, focusDistance})
	gopher.UpdateBounds()

	gopherLightPosition := vec3.T{gopher.Bounds.Center()[0] - 400, gopher.Bounds.Center()[1] + 400, gopher.Bounds.Center()[2] - 800}
	gopherLightEmission := (&color.Color{R: 15, G: 14.0, B: 12.0}).Multiply(gopherLightEmissionFactor)
	gopherLightMaterial := scn.Material{Color: &color.White, Emission: gopherLightEmission, Glossiness: 0.0, Roughness: 1.0, RayTerminator: true}
	gopherLight := scn.Sphere{Name: "Gopher light", Origin: &gopherLightPosition, Radius: 100.0, Material: &gopherLightMaterial}

	// Ground

	groundProjection := scn.NewParallelImageProjection("textures/ground/soil-cracked.png", vec3.Zero, vec3.UnitX.Scaled(gopher.Bounds.SizeY()*2), vec3.UnitZ.Scaled(gopher.Bounds.SizeY()*2))
	ground := scn.Disc{
		Name:   "Ground",
		Origin: &vec3.Zero,
		Normal: &vec3.UnitY,
		Radius: environmentRadius,
		Material: &scn.Material{
			Name:       "Ground material",
			Color:      &color.White,
			Emission:   &color.Black,
			Glossiness: 0.0,
			Roughness:  1.0,
			Projection: &groundProjection,
		},
	}

	// Spheres

	var spheres []*scn.Sphere
	for sphereIndex := 0; sphereIndex < amountSpheres; sphereIndex++ {
		sphere := getSphere(sphereRadius, sphereMinDistance, sphereMaxDistance)
		spheres = append(spheres, &sphere)
	}

	scene := scn.SceneNode{
		Spheres:         []*scn.Sphere{&environmentSphere, &gopherLight},
		FacetStructures: []*scn.FacetStructure{gopher},
		Discs:           []*scn.Disc{&ground},
		ChildNodes:      []*scn.SceneNode{{Spheres: spheres}},
	}

	startAngle := math.Pi / 2.0
	xPosMax := 300.0
	yPosMax := gopher.Bounds.SizeY() * 0.75
	yPosMin := gopher.Bounds.SizeY() * 0.15

	var frames []scn.Frame
	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		focusPoint := vec3.T{0, gopher.Bounds.SizeY() * 0.75, 0}

		angle := animationProgress * 2.0 * math.Pi
		xPos := xPosMax * math.Cos(angle+startAngle)
		yPos := yPosMin + (yPosMax-yPosMin)*(math.Sin(angle+startAngle)+1.0)/2.0
		cameraPosition := vec3.T{xPos, yPos, -600}

		camera := getCamera(magnification, cameraPosition, focusPoint)

		frame := scn.Frame{
			Filename:   animationName + "_" + fmt.Sprintf("%06d", frameIndex),
			FrameIndex: frameIndex,
			Camera:     &camera,
			SceneNode:  &scene,
		}

		frames = append(frames, frame)
	}

	animation := &scn.Animation{
		AnimationName:     animationName,
		Frames:            frames,
		Width:             int(float64(imageWidth) * magnification),
		Height:            int(float64(imageHeight) * magnification),
		WriteRawImageFile: false,
	}

	anm.WriteAnimationToFile(animation, false)
}

func getSphere(radius float64, minDistance, maxDistance float64) scn.Sphere {
	r := 0.50 + rand.Float64()*0.50
	g := 0.35 + rand.Float64()*0.45
	b := 0.35 + rand.Float64()*0.45
	sphereMaterial := scn.Material{
		Color:         &color.Color{R: float32(r), G: float32(g), B: float32(b)},
		Emission:      &color.Color{R: float32(r * lampEmissionFactor), G: float32(g * lampEmissionFactor), B: float32(b * lampEmissionFactor)},
		Glossiness:    0,
		Roughness:     1.0,
		RayTerminator: true,
	}

	x := (rand.Float64() - 0.5) * 2 * (maxDistance / 2.0)
	y := math.Pow(rand.Float64(), 2.0) * (maxDistance / 2.0)
	z := minDistance + rand.Float64()*(maxDistance-minDistance)

	origin := vec3.T{x, y, z}

	return scn.Sphere{
		Origin:   &origin,
		Radius:   radius,
		Material: &sphereMaterial,
	}
}

func GetGopher(scale float64) *scn.FacetStructure {
	var objFilename = "go_gopher_color.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/" + objFilename

	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objFilenamePath, err.Error())
	}
	defer objFile.Close()

	obj, err := obj.Read(objFile)

	// obj.ClearMaterials()

	obj.ScaleUniform(&vec3.Zero, scale)
	obj.UpdateBounds()

	fmt.Printf("Gopher bounds: %+v\n", obj.Bounds)

	return obj
}

func addEnvironmentMapping(filename string) scn.Sphere {
	origin := vec3.T{0, 0, 0}

	projection := scn.ImageProjection{
		ProjectionType: scn.ProjectionTypeSpherical,
		ImageFilename:  filename,
		Origin:         &origin,
		U:              &vec3.T{0, 0, 1},
		V:              &vec3.T{0, 1, 0},
		RepeatU:        true,
		RepeatV:        true,
		FlipU:          false,
		FlipV:          false,
	}

	material := scn.Material{
		Color:         &color.Color{R: 1.0, G: 1.0, B: 1.0},
		Emission:      &color.Color{R: 1.0 * environmentEmissionFactor, G: 1.0 * environmentEmissionFactor, B: 1.0 * environmentEmissionFactor},
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

func getCamera(magnification float64, cameraOrigin vec3.T, cameraFocus vec3.T) scn.Camera {
	cameraHeading := cameraFocus.Subed(&cameraOrigin)

	return scn.Camera{
		Origin:            &cameraOrigin,
		Heading:           &cameraHeading,
		ViewUp:            &vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		ApertureSize:      apertureSize,
		ApertureShape:     "textures/aperture/heart.png",
		FocusDistance:     cameraHeading.Length(),
		Samples:           amountSamples,
		AntiAlias:         true,
		Magnification:     magnification,
		RenderType:        renderType,
		RecursionDepth:    maxRecursion,
	}
}
