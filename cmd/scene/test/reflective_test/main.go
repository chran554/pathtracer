package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
	"strconv"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "reflective_test"

var ballRadius float64 = 20

var maxRecursionDepth = 6
var amountSamples = 10 * 1024
var lensRadius float64 = 2

var viewPlaneDistance = 4000.0
var cameraDistanceFactor = 2.0

var lampEmissionFactor = 8.0

var imageWidth = 1600
var imageHeight = 500
var magnification = 1.0

func main() {
	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false)

	scene := scn.NewSceneNode().D(getBoxWalls()...)

	amountSpheres := 5
	sphereSpread := ballRadius * 2.0 * (float64(amountSpheres) + 1)
	sphereCC := sphereSpread / float64(amountSpheres)

	for i := 0; i <= amountSpheres; i++ {
		reflectiveness := float64(i) / float64(amountSpheres)

		sphereOrigin := vec3.T{-sphereSpread/2.0 + float64(i)*sphereCC, ballRadius, 0}
		sphereMaterial := scn.NewMaterial().M(reflectiveness, 0.0)
		sphere := scn.NewSphere(&sphereOrigin, ballRadius, sphereMaterial).N("Sphere with reflectiveness of " + strconv.Itoa(i))

		scene.S(sphere)
	}

	lampMaterial := scn.NewMaterial().E(color.White, lampEmissionFactor, true)
	lampLeft := scn.NewSphere(&vec3.T{-0.3333 * sphereSpread, ballRadius*3 + ballRadius*2*0.75, -ballRadius * 3}, ballRadius*2, lampMaterial).N("Lamp left")
	lampMiddle := scn.NewSphere(&vec3.T{0.0, ballRadius*3 + ballRadius*2*0.75, -ballRadius * 3}, ballRadius*2, lampMaterial).N("Lamp middle")
	lampRight := scn.NewSphere(&vec3.T{0.3333 * sphereSpread, ballRadius*3 + ballRadius*2*0.75, -ballRadius * 3}, ballRadius*2, lampMaterial).N("Lamp right")

	scene.S(lampLeft, lampMiddle, lampRight)

	cameraOrigin := vec3.T{0, ballRadius, -400}
	cameraOrigin.Scale(cameraDistanceFactor)
	focusPoint := vec3.T{0, ballRadius, 0}
	camera := scn.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).V(viewPlaneDistance).A(lensRadius, "").D(maxRecursionDepth)

	frame := scn.NewFrame(animation.AnimationName, -1, camera, scene)

	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, true)
}

func getBoxWalls() []*scn.Disc {
	floorTexture := scn.NewParallelImageProjection("textures/tilesf4.jpeg", &vec3.T{0, 0, 0}, vec3.T{ballRadius * 4, 0, 0}, vec3.T{0, 0, ballRadius * 4})
	backWallTexture := scn.NewParallelImageProjection("textures/bricks_yellow.png", &vec3.T{0, 0, 0}, vec3.T{ballRadius * 9, 0, 0}, vec3.T{0, ballRadius * 9, 0})

	floor := scn.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, 600, scn.NewMaterial().P(&floorTexture)).N("Floor")
	roof := scn.NewDisc(&vec3.T{0, ballRadius * 3, 0}, &vec3.T{0, -1, 0}, 600, scn.NewMaterial().C(color.Color{R: 0.9, G: 1, B: 0.95})).N("Roof")
	backWall := scn.NewDisc(&vec3.T{0, 0, ballRadius * 3}, &vec3.T{0, 0, -1}, 600, scn.NewMaterial().P(&backWallTexture)).N("Back wall")

	return []*scn.Disc{floor, roof, backWall}
}
