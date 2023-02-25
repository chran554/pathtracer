package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
	"strconv"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "dof_test"

var ballRadius float64 = 30

var amountSamples = 128 * 8 * 8 * 3
var lensRadius = 0.0 // 12.0

var viewPlaneDistance = 2000.0
var cameraDistanceFactor = 1.0

var imageWidth = 800
var imageHeight = 400
var magnification = 1.0

func main() {
	cornellBox := obj.NewCornellBox(&vec3.T{500, 300, 500}, 7.0) // cm, as units. I.e. a 5x3x5m room

	amountSpheres := 5
	sphereSpread := ballRadius * 2.0 * (float64(amountSpheres) + 1)
	sphereCC := sphereSpread / float64(amountSpheres)

	sphereMaterial := scn.NewMaterial().C(color.Color{R: 0.85, G: 0.95, B: 0.80}).M(0.4, 0.05)

	var spheres []*scn.Sphere
	for i := 0; i <= amountSpheres; i++ {
		positionOffsetX := (-sphereSpread/2.0 + float64(i)*sphereCC) * 0.5
		positionOffsetZ := (-sphereSpread/2.0 + float64(i)*sphereCC) * 1.0

		sphere := scn.NewSphere(&vec3.T{positionOffsetX, ballRadius, positionOffsetZ}, ballRadius, sphereMaterial).
			N("Glass sphere with transparency " + strconv.Itoa(i))

		spheres = append(spheres, sphere)
	}

	scene := scn.NewSceneNode().S(spheres...).FS(cornellBox)

	cameraOrigin := vec3.T{0, ballRadius * 3, -800}
	cameraOrigin.Scale(cameraDistanceFactor)
	focusPoint := vec3.T{0, ballRadius, 0}
	camera := scn.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).V(viewPlaneDistance).A(lensRadius, "")

	frame := scn.NewFrame(animationName, -1, camera, scene)

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false)
	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, false)
}
