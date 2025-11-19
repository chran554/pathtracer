package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "cornellbox"

var ballRadius float64 = 20

// var amountSamples = 1024 * 64
var maxRecursionDepth = 4

var imageWidth = 800
var imageHeight = 400
var magnification = 0.5

var viewPlaneDistance = 1500.0

func main() {
	// TODO Set lamp size of cornell single light to: 		lampPercentageOfCeiling := 0.20
	cornellBoxUnit := ballRadius * 3
	cornellBox := obj.NewCornellBox(&vec3.T{2 * cornellBoxUnit, cornellBoxUnit, 3 * cornellBoxUnit}, true, 40)

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

	//camera := scene.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).V(viewPlaneDistance).D(maxRecursionDepth)
	//frame := scene.NewFrame(animationName, -1, camera, scn)
	//animation.AddFrame(frame)

	for frameIndex := 0; frameIndex <= 16; frameIndex++ {
		amountSamples := int(math.Pow(2, float64(frameIndex)))
		camera := scene.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).V(viewPlaneDistance).D(maxRecursionDepth)
		frame := scene.NewFrame(animationName, amountSamples, camera, scn)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
