package main

import (
	"fmt"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "cornellbox"

var ballRadius float64 = 20

var lightIntensityFactor = 1.5

var amountSamples = 1024 * 32
var maxRayDepth = 6

var imageWidth = 800
var imageHeight = 400
var magnification = 1.0

var viewPlaneDistance = 1500.0

func main() {
	cornellBoxUnit := ballRadius * 3
	cornellBox := obj.NewCornellBox(&vec3.T{2 * cornellBoxUnit, cornellBoxUnit, 3 * cornellBoxUnit}, true, lightIntensityFactor)

	rightSphereMaterial := scene.NewMaterial().N("Right sphere").C(color.NewColorGrey(0.9))
	leftSphereMaterial := scene.NewMaterial().N("Left sphere").C(color.NewColorGrey(0.9))

	rightSpherePosition := vec3.T{ballRadius + (ballRadius / 2), ballRadius, 0}
	leftSpherePosition := vec3.T{-(ballRadius + (ballRadius / 2)), ballRadius, 0}
	sphere1 := scene.NewSphere(&rightSpherePosition, ballRadius, rightSphereMaterial).N("Right sphere")
	sphere2 := scene.NewSphere(&leftSpherePosition, ballRadius, leftSphereMaterial).N("Left sphere")

	scn := scene.NewSceneNode().S(sphere1, sphere2).FS(cornellBox)

	cameraOrigin := cornellBox.Bounds.Center().Add(&vec3.T{0, 0, -15 * ballRadius})
	focusPoint := cornellBox.Bounds.Center()

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, false)

	camera := scene.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).V(viewPlaneDistance).D(maxRayDepth)
	frame := scene.NewFrame(animationName, -1, camera, scn)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
