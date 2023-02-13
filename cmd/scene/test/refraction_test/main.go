package main

import (
	"fmt"
	"math"
	"os"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
	"strconv"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "refraction_test"

var ballRadius float64 = 30

var renderType = scn.Pathtracing

// var renderType = scn.Raycasting

var maxRecursionDepth = 8
var amountSamples = 512 * 5 * 2 // * 4
var lensRadius = 2.0            // 0.25

var cameraDistanceFactor = 1.0

var imageWidth = 800
var imageHeight = 400
var magnification = 1.5

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

	cornellBox := GetCornellBox(&vec3.T{300, 300, 300}, 10.0) // cm, as units. I.e. a 5x3x5m room

	spheres := GetSpheres(1, &vec3.T{-40, 0, 0})

	glassPokal := obj.NewGlassIkeaPokal(50.0)
	glassPokal.Translate(&vec3.T{10, 0, -20})

	glassSkoja := obj.NewGlassIkeaSkoja(40.0)
	glassSkoja.Translate(&vec3.T{35, 0, 0})

	utahTeapot := obj.NewSolidUtahTeapot(50.0)
	utahTeapot.RotateY(&vec3.T{0, 0, 0}, -math.Pi/3.5-math.Pi/2.0)
	utahTeapot.Translate(&vec3.T{25, 0, 150})

	pixarBallRadius := 20.0
	pixarBallOrigin := &vec3.T{-130, pixarBallRadius, 160}
	pixarBall := obj.NewPixarBall(pixarBallOrigin, pixarBallRadius)
	//pixarBall.RotateY(pixarBallOrigin, 0.0)
	pixarBall.RotateY(pixarBallOrigin, math.Pi/4.0)

	scene := scn.SceneNode{
		Spheres:         spheres,
		FacetStructures: []*scn.FacetStructure{cornellBox /*glassPokal, glassSkoja*/, utahTeapot},
	}
	scene.Spheres = append(scene.Spheres, pixarBall)

	origin := &vec3.T{0, 70, -200}
	focusPoint := spheres[0].Origin.Added(&vec3.T{0, 0, -ballRadius / 3.0})
	camera := scn.NewCamera(origin, &focusPoint).S(amountSamples).D(maxRecursionDepth).A(lensRadius, "").M(magnification)

	animation := &scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             width,
		Height:            height,
		WriteRawImageFile: false,
	}

	frame := scn.Frame{
		Filename:   animation.AnimationName,
		FrameIndex: 0,
		Camera:     camera,
		SceneNode:  &scene,
	}

	animation.Frames = append(animation.Frames, frame)

	anm.WriteAnimationToFile(animation, false)
}

func GetSpheres(amountSpheres int, translation *vec3.T) []*scn.Sphere {
	//sphereSpread := ballRadius * 2.0 * (float64(amountSpheres) + 1)
	//sphereCC := sphereSpread / float64(amountSpheres)

	var spheres []*scn.Sphere
	sphereMaterial := scn.NewMaterial().
		N("glass sphere").
		C(color.Color{R: 0.95, G: 0.95, B: 0.99}, 1.0).
		M(0.5, 0.05).
		T(0.94, true, scn.RefractionIndex_Glass)

	for i := 0; i < amountSpheres; i++ {
		positionOffsetX := 0.0 // (-sphereSpread/2.0 + float64(i)*sphereCC) * 0.5
		positionOffsetZ := 0.0 // (-sphereSpread/2.0 + float64(i)*sphereCC) * 1.0

		groundOffset := 0.02 // ballRadius / 2.0 // Raise the sphere above the ground
		sphere := scn.Sphere{
			Name:     "Glass sphere #" + strconv.Itoa(i),
			Origin:   &vec3.T{positionOffsetX, ballRadius + groundOffset, positionOffsetZ},
			Radius:   ballRadius,
			Material: sphereMaterial,
		}

		sphere.Translate(translation)

		spheres = append(spheres, &sphere)
	}
	return spheres
}

func GetCornellBox(scale *vec3.T, lightIntensityFactor float64) *scn.FacetStructure {
	var cornellBoxFilename = "cornellbox.obj"
	var cornellBoxFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + cornellBoxFilename

	cornellBoxFile, err := os.Open(cornellBoxFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", cornellBoxFilenamePath, err.Error())
	}
	defer cornellBoxFile.Close()

	cornellBox, err := obj.Read(cornellBoxFile)
	cornellBox.Scale(&vec3.Zero, scale)
	cornellBox.ClearMaterials()

	cornellBox.Material = scn.NewMaterial().
		N("Cornell box material").
		C(color.Color{R: 0.95, G: 0.95, B: 0.95}, 1.0)

	backWallMaterial := cornellBox.Material.Copy().
		PP("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png", vec3.Zero, vec3.UnitX.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	sideWallMaterial := cornellBox.Material.Copy().
		PP("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png", vec3.Zero, vec3.UnitZ.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	floorMaterial := cornellBox.Material.Copy().
		PP("textures/floor/7451-diffuse 02 low contrast.png", vec3.Zero, vec3.UnitX.Scaled(scale[0]*0.25), vec3.UnitZ.Scaled(scale[0]*0.25))

	lampMaterial := scn.NewMaterial().N("Lamp").
		C(color.White, 1.0).
		E(color.White, lightIntensityFactor, true)

	cornellBox.GetFirstObjectByName("Lamp_1").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_2").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_3").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_4").Material = lampMaterial

	cornellBox.GetFirstObjectByName("Wall_back").Material = backWallMaterial
	cornellBox.GetFirstObjectByName("Wall_left").Material = sideWallMaterial
	cornellBox.GetFirstObjectByName("Wall_right").Material = sideWallMaterial
	cornellBox.GetFirstObjectByName("Floor_2").Material = floorMaterial

	return cornellBox
}
