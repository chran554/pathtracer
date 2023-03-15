package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "fresnel"

var ballRadius float64 = 20

var maxRecursionDepth = 8
var amountSamples = 1024 * 16

var viewPlaneDistance = 1500.0

var imageWidth = 800
var imageHeight = 400
var magnification = 1.0

func main() {
	cornellBoxUnit := ballRadius * 3.0
	cornellBox := obj.NewCornellBox(&vec3.T{2 * cornellBoxUnit, cornellBoxUnit, 3 * cornellBoxUnit}, true, 3)

	rightSphereMaterial := scn.NewMaterial().N("right_sphere").
		C(color.NewColor(0.9, 0.8, 0.9)).
		M(0.04, 0.0).
		T(0.0, true, scn.RefractionIndex_Air)

	leftSphereMaterial := scn.NewMaterial().N("left_sphere").
		C(color.NewColor(0.9, 0.8, 0.9)).
		M(0.0, 0.0).
		T(0.0, true, scn.RefractionIndex_Water)

	middleSphereMaterial := scn.NewMaterial().N("middle_sphere").
		C(color.NewColorKelvin(3000)).
		M(0.0, 0.0).
		E(color.White, 4.0, false)

	lampDepth := cornellBox.GetFirstObjectByName("Lamp").Bounds.Zmax

	sphereX := ballRadius + (ballRadius / 2)
	sphere1 := scn.NewSphere(&vec3.T{sphereX, ballRadius, lampDepth - 2*ballRadius}, ballRadius, rightSphereMaterial).N("Right sphere")
	sphere2 := scn.NewSphere(&vec3.T{-sphereX, ballRadius, lampDepth - 2*ballRadius}, ballRadius, leftSphereMaterial).N("Left sphere")
	sphereM := scn.NewSphere(&vec3.T{0, ballRadius / 2, -ballRadius * 2}, ballRadius/2, middleSphereMaterial).N("Middle sphere")

	scene := scn.NewSceneNode().S(sphere1, sphere2, sphereM).FS(cornellBox)

	boxCenter := cornellBox.Bounds.Center()
	cameraOrigin := boxCenter.Added(&vec3.T{0, 0, -ballRadius * 15.5})
	focusPoint := boxCenter
	camera := scn.NewCamera(&cameraOrigin, focusPoint, amountSamples, magnification).V(viewPlaneDistance).D(maxRecursionDepth)

	frame := scn.NewFrame(animationName, -1, camera, scene)

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true)
	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, false)
}
