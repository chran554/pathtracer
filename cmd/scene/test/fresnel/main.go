package main

import (
	"fmt"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

// Two Spheres in a cornell box (+ a small sphere light in the center).
// The left sphere has reflections through Fresnel (using refraction index).
// The right sphere has matching reflections through common mirror (glossiness and roughness).

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

	rightSphereMaterial := scene.NewMaterial().N("right_sphere").
		C(color.NewColor(0.9, 0.8, 0.9)).
		M(0.04, 0.0).
		T(0.0, true, scene.RefractionIndex_Air)

	leftSphereMaterial := scene.NewMaterial().N("left_sphere").
		C(color.NewColor(0.9, 0.8, 0.9)).
		M(0.0, 0.0).
		T(0.0, true, scene.RefractionIndex_Water)

	middleSphereMaterial := scene.NewMaterial().N("middle_sphere").
		C(color.NewColorKelvin(3000)).
		M(0.0, 0.0).
		E(color.White, 4.0, false)

	lampDepth := cornellBox.GetFirstObjectByName("Lamp").Bounds.Zmax

	sphereX := ballRadius + (ballRadius / 2)
	sphere1 := scene.NewSphere(&vec3.T{sphereX, ballRadius, lampDepth - 2*ballRadius}, ballRadius, rightSphereMaterial).N("Right sphere")
	sphere2 := scene.NewSphere(&vec3.T{-sphereX, ballRadius, lampDepth - 2*ballRadius}, ballRadius, leftSphereMaterial).N("Left sphere")
	sphereM := scene.NewSphere(&vec3.T{0, ballRadius / 2, -ballRadius * 2}, ballRadius/2, middleSphereMaterial).N("Middle sphere")

	scn := scene.NewSceneNode().S(sphere1, sphere2, sphereM).FS(cornellBox)

	boxCenter := cornellBox.Bounds.Center()
	cameraOrigin := boxCenter.Added(&vec3.T{0, 0, -ballRadius * 15.5})
	focusPoint := boxCenter
	camera := scene.NewCamera(&cameraOrigin, focusPoint, amountSamples, magnification).V(viewPlaneDistance).D(maxRecursionDepth)

	frame := scene.NewFrame(animationName, -1, camera, scn)

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, false)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
