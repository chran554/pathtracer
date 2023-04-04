package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "castle_test"

var environmentRadius = 500.0 * 1000.0
var environmentEmissionFactor = 1.0

var imageWidth = 1024
var imageHeight = 768
var magnification = 1.5

var amountSamples = 1024 * 6
var maxRecursion = 4

var apertureSize = 0.75

var cameraDistance = 100.0
var cameraHeight = 0.0

func main() {

	castle := obj.NewCastle(80)
	castleBounds := castle.Bounds

	lampColor := color.Color{R: 1.0, G: 0.87, B: 0.5}
	castleLamp := createLamp("castle_lamp", 6.5, 15.0, castleBounds.Center().Add(&vec3.T{13, -8, -2}), lampColor)
	entranceLamp1 := createLamp("entrance_lamp_1", 3.0, 6.0, castleBounds.Center().Add(&vec3.T{0, -21, -3}), lampColor)
	entranceLamp2 := createLamp("entrance_lamp_2", 3.0, 4.0, castleBounds.Center().Add(&vec3.T{-15, -25, -3}), lampColor)
	entranceLamp3 := createLamp("entrance_lamp_3", 3.0, 4.0, castleBounds.Center().Add(&vec3.T{15, -25, -3}), lampColor)
	towerLamp1 := createLamp("tower_lamp_1", 1.3, 5.0, castleBounds.Center().Add(&vec3.T{-10.3, 1, -13}), lampColor)
	towerLamp2 := createLamp("tower_lamp_2", 1.3, 5.0, castleBounds.Center().Add(&vec3.T{10.3, 1, -13}), lampColor)
	tinnerLamp1 := createLamp("tinner_lamp_1", 3.0, 5.0, castleBounds.Center().Add(&vec3.T{10, 25, 6}), lampColor)
	var lamps = []*scn.Sphere{castleLamp, entranceLamp1, entranceLamp2, entranceLamp3, towerLamp1, towerLamp2, tinnerLamp1}

	// Sky dome
	// var environmentEnvironMap =
	environmentSphere := addEnvironmentMapping("textures/equirectangular/sunset horizon 2800x1400.jpg")
	// environmentSphere := addEnvironmentMapping("textures/equirectangular/nightsky.png")

	// Ground
	groundProjection := scn.NewParallelImageProjection("textures/ground/grass_short.png", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(80/2), vec3.UnitZ.Scaled(50/2))
	groundMaterial := scn.NewMaterial().N("Ground material").P(&groundProjection)
	ground := scn.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, environmentRadius, groundMaterial).N("Ground")

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	cameraOffset := &vec3.T{}
	cameraOffset[0] = 0.0 * cameraDistance
	cameraOffset[1] = cameraHeight
	cameraOffset[2] = -1.0 * cameraDistance

	focusPointOffset := cameraOffset.Normalized()
	focusPointOffset.Scale(30.0) // Focus point is 15 units from object center towards camera point
	// focusPointOffset.Add(&vec3.T{0, -10, 0}) // For castle, set focus point a bit lower than center of object

	cameraOrigin := castleBounds.Center().Add(cameraOffset)
	cameraFocusPoint := castleBounds.Center().Add(&focusPointOffset).Add(&vec3.T{0, -5, 0})

	camera := scn.NewCamera(cameraOrigin, cameraFocusPoint, amountSamples, magnification).D(maxRecursion).A(apertureSize, "")

	scene := scn.NewSceneNode().
		S(environmentSphere).
		S(lamps...).
		D(ground).
		FS(castle)

	frame := scn.NewFrame(animationName, -1, camera, scene)

	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, false)
}

func createLamp(lampName string, size float64, strength float64, lampPosition *vec3.T, lampColor color.Color) *scn.Sphere {
	material := scn.NewMaterial().N(lampName).C(color.White).E(lampColor, strength, true)
	return scn.NewSphere(lampPosition, size, material).N(lampName)
}

func addEnvironmentMapping(filename string) *scn.Sphere {
	origin := vec3.T{0, 0, 0}
	u := vec3.T{-0.2, 0, -1}
	v := vec3.T{0, 1, 0}
	material := scn.NewMaterial().E(color.White, environmentEmissionFactor, true).SP(filename, &origin, u, v)
	sphere := scn.NewSphere(&origin, environmentRadius, material).N("Environment mapping")

	return sphere
}
