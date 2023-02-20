package main

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

var animationName = "reflective_test"

var ballRadius float64 = 20

var maxRecursionDepth = 6
var amountSamples = 3 * 1024
var lensRadius float64 = 2

var viewPlaneDistance = 4000.0
var cameraDistanceFactor = 2.0

var lampEmissionFactor = 15.0

var imageWidth = 1800
var imageHeight = 1200
var magnification = 0.75

func main() {
	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false)

	scene := scn.NewSceneNode().D(getBoxWalls()...)

	amountSpheres := 6
	sphereSpread := ballRadius * 2.0 * (float64(amountSpheres) + 1) * 1.3
	sphereCC := sphereSpread / float64(amountSpheres)

	for roughIndex := 0; roughIndex <= amountSpheres; roughIndex++ {
		roughness := float64(roughIndex) / float64(amountSpheres)

		for glossyIndex := 0; glossyIndex <= amountSpheres; glossyIndex++ {
			glossiness := float64(glossyIndex) / float64(amountSpheres)

			sphereOrigin := vec3.T{-sphereSpread/2.0 + float64(glossyIndex)*sphereCC, ballRadius, -sphereSpread/2.0 + float64(roughIndex)*sphereCC}
			sphereMaterial := scn.NewMaterial().C(color.NewColor(0.8, 1.0, 0.6)).M(glossiness, roughness)
			sphere := scn.NewSphere(&sphereOrigin, ballRadius, sphereMaterial).N(fmt.Sprintf("Sphere (glossy:%02f rough:%02f)", glossiness, roughness))

			scene.S(sphere)
		}
	}

	lampMaterial := scn.NewMaterial().E(color.White, lampEmissionFactor, true)
	lampHeight := 400.0 // ballRadius*3 + ballRadius*2*0.75
	lampRadius := ballRadius * 4
	lampLeft := scn.NewSphere(&vec3.T{-0.75 * sphereSpread, lampHeight, -ballRadius * 3}, lampRadius, lampMaterial).N("Lamp left")
	lampMiddle := scn.NewSphere(&vec3.T{0.0, lampHeight, -ballRadius * 3}, lampRadius, lampMaterial).N("Lamp middle")
	lampRight := scn.NewSphere(&vec3.T{0.75 * sphereSpread, lampHeight, -ballRadius * 3}, lampRadius, lampMaterial).N("Lamp right")

	scene.S(lampLeft, lampMiddle, lampRight)

	cameraOrigin := vec3.T{0, 400, -400}
	cameraOrigin.Scale(cameraDistanceFactor)
	focusPoint := vec3.T{0, ballRadius, -ballRadius * 2}
	camera := scn.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).V(viewPlaneDistance).A(lensRadius, "").D(maxRecursionDepth)

	frame := scn.NewFrame(animation.AnimationName, -1, camera, scene)

	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, true)
}

func getBoxWalls() []*scn.Disc {
	floorTexture := scn.NewParallelImageProjection("textures/floor/tilesf4.jpeg", &vec3.T{0, 0, 0}, vec3.T{ballRadius * 4, 0, 0}, vec3.T{0, 0, ballRadius * 4})

	floor := scn.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, 6000, scn.NewMaterial().P(&floorTexture)).N("Floor")
	roof := scn.NewDisc(&vec3.T{0, 410 * cameraDistanceFactor, 0}, &vec3.T{0, -1, 0}, 6000, scn.NewMaterial()).N("Roof")
	backWall := scn.NewDisc(&vec3.T{0, 0, 400}, &vec3.T{0, 0, -1}, 3000, scn.NewMaterial()).N("Back wall")
	leftWall := scn.NewDisc(&vec3.T{-350, 0, 0}, &vec3.T{1, 0, 0}, 3000, scn.NewMaterial()).N("Left wall")
	rightWall := scn.NewDisc(&vec3.T{350, 0, 0}, &vec3.T{-10, 0, 0}, 3000, scn.NewMaterial()).N("Right wall")

	return []*scn.Disc{floor, roof, backWall, leftWall, rightWall}
}
