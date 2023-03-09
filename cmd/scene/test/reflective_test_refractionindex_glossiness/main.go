package main

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
)

var animationName = "reflective_test_refractionindex_glossiness"

var ballRadius float64 = 20

var maxRecursionDepth = 6
var amountSamples = 1024 * 12
var lensRadius float64 = 2

var viewPlaneDistance = 4000.0
var cameraDistanceFactor = 2.0

var lampEmissionFactor = 12.0

var imageWidth = 1800
var imageHeight = 1200
var magnification = 0.75

func main() {
	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false)

	cornellBox := obj.NewCornellBox(&vec3.T{700, 810, 700}, lampEmissionFactor)
	cornellBox.ReplaceMaterial("left", scn.NewMaterial().N("left").C(color.NewColor(0.85, 0.85, 0.85)))
	cornellBox.ReplaceMaterial("right", scn.NewMaterial().N("right").C(color.NewColor(0.85, 0.85, 0.85)))
	cornellBox.ReplaceMaterial("back", scn.NewMaterial().N("back").C(color.NewColor(0.70, 0.70, 0.70)))

	scene := scn.NewSceneNode().FS(cornellBox)

	amountSpheres := 6
	sphereSpread := ballRadius * 2.0 * (float64(amountSpheres) + 1) * 1.3
	sphereCC := sphereSpread / float64(amountSpheres)

	for yIndex := 0; yIndex <= amountSpheres; yIndex++ {
		for xIndex := 0; xIndex <= amountSpheres; xIndex++ {
			yProgress := float64(yIndex) / float64(amountSpheres)
			xProgress := float64(xIndex) / float64(amountSpheres)

			refractiveIndex := interpolate(xProgress, scn.RefractionIndex_Air, scn.RefractionIndex_Diamond)
			glossiness := yProgress
			roughness := 0.0

			sphereMaterial := scn.NewMaterial().
				C(color.NewColor(0.80, 0.95, 0.80)).
				M(glossiness, roughness).
				T(0.0, true, refractiveIndex)

			sphereOrigin := vec3.T{-sphereSpread/2.0 + float64(xIndex)*sphereCC, ballRadius, -sphereSpread/2.0 + float64(yIndex)*sphereCC}
			sphere := scn.NewSphere(&sphereOrigin, ballRadius, sphereMaterial).N(fmt.Sprintf("Sphere (glossy:%02f rough:%02f)", xProgress, yProgress))

			scene.S(sphere)
		}
	}

	cameraOrigin := vec3.T{0, 400, -400}
	cameraOrigin.Scale(cameraDistanceFactor)
	focusPoint := vec3.T{0, ballRadius, -ballRadius * 2}
	camera := scn.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).V(viewPlaneDistance).A(lensRadius, "").D(maxRecursionDepth)

	frame := scn.NewFrame(animation.AnimationName, -1, camera, scene)

	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, true)
}

func interpolate(progress float64, a float64, b float64) float64 {
	return progress*b + (1.0-progress)*a
}
