package main

import (
	"fmt"
	"math"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

// var environmentEnvironMap = "textures/planets/environmentmap/space_fake_02_flip.png"
// var environmentEnvironMap = "textures/equirectangular/open_grassfield_sunny_day.jpg"

// var environmentEnvironMap = "textures/equirectangular/canyon 3200x1600.jpeg"
var environmentRadius = 100.0 * 100.0 // 100m (if 1 unit is 1 cm)
var environmentEmissionFactor = 1.7

//var environmentEnvironMap = "textures/planets/environmentmap/Stellarium3.jpeg"
//var environmentRadius = 100.0 * 80.0 // 80m (if 1 unit is 1 cm)

var animationName = "gordian_knot"

var amountFrames = 36 * 10

var imageWidth = 1280
var imageHeight = 1024
var magnification = 0.5

var renderType = scn.Pathtracing
var amountSamples = 256
var maxRecursion = 3

var cameraDistanceFactor = 2.8

var viewPlaneDistance = 800.0
var lensRadius = 0.0

var pipeLength = 0.83333
var pipeRadius = 40.0
var amountBalls = 75
var scale = 100.0

func main() {
	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, false)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		//ballSpeed := amountBalls / 5.0 // All balls make 1/5:th of a full loop during the whole animation
		ballSpeed := float64(amountBalls) / pipeLength // All balls make a full loop during the whole animation
		spheres := getKnotBalls(pipeRadius, amountBalls, scale, animationProgress, ballSpeed)

		environmentOrigin := &vec3.T{0, 0, 0}
		environmentMaterial := scn.NewMaterial().
			E(color.White, environmentEmissionFactor, true).
			SP("textures/equirectangular/forest_sunny_day.jpg", environmentOrigin, vec3.UnitZ, vec3.UnitY)
		environmentSphere := scn.NewSphere(environmentOrigin, environmentRadius, environmentMaterial).N("Environment mapping")

		scene := scn.NewSceneNode().S(spheres...).S(environmentSphere)
		scene.Bounds = nil

		camera := getCamera(magnification, animationProgress)

		frame := scn.NewFrame(animation.AnimationName, frameIndex, camera, scene)

		animation.AddFrame(frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func getKnotBalls(ballRadius float64, amountBalls int, scale float64, animationProgress float64, ballSpeed float64) []*scn.Sphere {
	spheres := make([]*scn.Sphere, 0)
	radianDistanceBetweenBalls := (math.Pi * 2.0 * pipeLength) / float64(amountBalls)

	sphereMaterial := scn.NewMaterial().C(color.NewColor(0.95, 0.95, 0.95)).M(0.9, 0.1)

	for ballIndex := 0; ballIndex < amountBalls; ballIndex++ {
		t := (radianDistanceBetweenBalls * float64(ballIndex)) + (animationProgress * ballSpeed * radianDistanceBetweenBalls)

		x := scale * (math.Cos(t) + 2.0*math.Cos(2.0*t))
		y := scale * (math.Sin(t) - 2.0*math.Sin(2.0*t))
		z := scale * -1.0 * math.Sin(3.0*t)

		sphere := scn.NewSphere(&vec3.T{x, y, z}, ballRadius, sphereMaterial).N(fmt.Sprintf("sphere #%d", ballIndex))

		spheres = append(spheres, sphere)
	}

	return spheres
}

func getCamera(magnification float64, progress float64) *scn.Camera {
	degrees45 := math.Pi / 4.0
	strideAngle := degrees45 * math.Sin(2.0*math.Pi*progress)
	cameraDistance := 200.0 * cameraDistanceFactor
	cameraHeight := 75.0 * cameraDistanceFactor

	cameraOrigin := vec3.T{
		cameraDistance * math.Cos(-math.Pi/2.0+strideAngle+degrees45),
		cameraHeight + (cameraHeight/2.0)*math.Sin(2.0*math.Pi*2.0*progress),
		cameraDistance * math.Sin(-math.Pi/2.0+strideAngle*0.5+degrees45),
	}

	focusPoint := vec3.T{0, 0, 0}

	return scn.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification)
}
