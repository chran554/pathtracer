package main

import (
	"fmt"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "reflective_test_glossy_roughness"

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
	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	cornellBox := obj.NewCornellBox(&vec3.T{700, 500, 700}, false, lampEmissionFactor)
	cornellBox.ReplaceMaterial("left", scene.NewMaterial().N("left").C(color.NewColor(0.85, 0.85, 0.85)))
	cornellBox.ReplaceMaterial("right", scene.NewMaterial().N("right").C(color.NewColor(0.85, 0.85, 0.85)))
	cornellBox.ReplaceMaterial("back", scene.NewMaterial().N("back").C(color.NewColor(0.70, 0.70, 0.70)))

	scn := scene.NewSceneNode().FS(cornellBox)

	amountSpheres := 6
	sphereSpread := ballRadius * 2.0 * (float64(amountSpheres) + 1) * 1.3
	sphereCC := sphereSpread / float64(amountSpheres)

	for yIndex := 0; yIndex <= amountSpheres; yIndex++ {
		for xIndex := 0; xIndex <= amountSpheres; xIndex++ {
			yProgress := float64(yIndex) / float64(amountSpheres)
			xProgress := float64(xIndex) / float64(amountSpheres)

			refractiveIndex := scene.RefractionIndex_Air
			glossiness := xProgress
			roughness := yProgress

			sphereMaterial := scene.NewMaterial().
				C(color.NewColor(0.80, 0.95, 0.80)).
				M(glossiness, roughness).
				T(0.0, true, refractiveIndex)

			sphereOrigin := vec3.T{-sphereSpread/2.0 + float64(xIndex)*sphereCC, ballRadius, -sphereSpread/2.0 + float64(yIndex)*sphereCC}
			sphere := scene.NewSphere(&sphereOrigin, ballRadius, sphereMaterial).N(fmt.Sprintf("Sphere (glossy:%02f rough:%02f)", xProgress, yProgress))

			scn.S(sphere)
		}
	}

	cameraOrigin := vec3.T{0, 400, -400}
	cameraOrigin.Scale(cameraDistanceFactor)
	focusPoint := vec3.T{0, ballRadius, -ballRadius * 2}
	camera := scene.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).V(viewPlaneDistance).A(lensRadius, nil).D(maxRecursionDepth)

	frame := scene.NewFrame(animation.AnimationName, -1, camera, scn)

	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
