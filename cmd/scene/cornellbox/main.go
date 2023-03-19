package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "cornellbox"

var ballRadius float64 = 20

var amountSamples = 1024 * 32

var imageWidth = 800
var imageHeight = 400
var magnification = 1.0

var viewPlaneDistance = 1500.0

func main() {
	cornellBoxUnit := ballRadius * 3
	cornellBox := obj.NewCornellBox(&vec3.T{2 * cornellBoxUnit, cornellBoxUnit, 3 * cornellBoxUnit}, true, 40)

	rightSphereMaterial := scn.NewMaterial().N("Right sphere").C(color.NewColorGrey(0.9))
	leftSphereMaterial := scn.NewMaterial().N("Left sphere").C(color.NewColorGrey(0.9))

	rightSpherePosition := vec3.T{ballRadius + (ballRadius / 2), ballRadius, 0}
	leftSpherePosition := vec3.T{-(ballRadius + (ballRadius / 2)), ballRadius, 0}
	sphere1 := scn.NewSphere(&rightSpherePosition, ballRadius, rightSphereMaterial).N("Right sphere")
	sphere2 := scn.NewSphere(&leftSpherePosition, ballRadius, leftSphereMaterial).N("Left sphere")

	scene := scn.NewSceneNode().S(sphere1, sphere2).FS(cornellBox)

	cameraOrigin := cornellBox.Bounds.Center().Add(&vec3.T{0, 0, -15 * ballRadius})
	focusPoint := cornellBox.Bounds.Center()

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, false)

	camera := scn.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).V(viewPlaneDistance)
	frame := scn.NewFrame(animationName, -1, camera, scene)
	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, false)
}
