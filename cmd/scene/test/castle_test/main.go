package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "castle_test"

var amountAnimationFrames = 180
var animationStartIndex = 0

var environmentRadius = 500.0 * 1000.0
var environmentEmissionFactor = 1.0

var imageWidth = 1024
var imageHeight = 768
var magnification = 0.5 // 0.4

var amountSamples = 1400 // 1024 * 6
var maxRecursion = 5

var apertureSize = 0.25

var cameraDistance = 100.0
var cameraHeight = 25.0

func main() {
	castle := obj.NewCastle(80, color.NewColorKelvin(2500), 10)
	castle.UpdateVertexNormalsWithThreshold(false, 0)
	castleBounds := castle.Bounds

	// Sky dome
	// var environmentEnvironMap =
	environmentSphere := addEnvironmentMapping("textures/equirectangular/sunset horizon 2800x1400.jpg")
	// environmentSphere := addEnvironmentMapping("textures/equirectangular/nightsky.png")

	// Ground
	groundProjection := scene.NewParallelImageProjection(floatimage.LoadOrPanic("textures/ground/grass_short.png"), &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(80/2), vec3.UnitZ.Scaled(50/2))
	groundMaterial := scene.NewMaterial().N("Ground material").P(&groundProjection)
	ground := scene.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, environmentRadius, groundMaterial).N("Ground")

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	scn := scene.NewSceneNode().
		S(environmentSphere).
		D(ground).
		FS(castle)

	for frameIndex := animationStartIndex; frameIndex < amountAnimationFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountAnimationFrames)
		animationProgress = 0.5 + math.Sin(-math.Pi/2+math.Pi*animationProgress)*0.5 // Ease in, and ease out

		angle := util.DegToRad(animationProgress*180) + util.DegToRad(-90)

		cameraOffsetFromCastleCenter := &vec3.T{}
		cameraOffsetFromCastleCenter[0] = math.Cos(angle) * 0.5 * cameraDistance
		cameraOffsetFromCastleCenter[1] = cameraHeight
		cameraOffsetFromCastleCenter[2] = math.Sin(angle) * 1.0 * cameraDistance

		focusPointOffset := cameraOffsetFromCastleCenter.Normalized()
		focusPointOffset.Scale(castleBounds.Zmax * 0.75) // Focus point is 15 units from object center towards camera point
		// focusPointOffset.Add(&vec3.T{0, -10, 0}) // For castle, set focus point a bit lower than center of object

		cameraOrigin := castleBounds.Center().Add(cameraOffsetFromCastleCenter)
		cameraFocusPoint := castleBounds.Center().Add(&focusPointOffset).Add(&vec3.T{0, -5, 0})

		camera := scene.NewCamera(cameraOrigin, cameraFocusPoint, amountSamples, magnification).D(maxRecursion).A(apertureSize, nil)

		fi := -1
		if amountAnimationFrames > 1 {
			fi = frameIndex
		}

		frame := scene.NewFrame(animationName, fi, camera, scn)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func addEnvironmentMapping(filename string) *scene.Sphere {
	origin := vec3.T{0, 0, 0}
	u := vec3.T{-0.2, 0, -1}
	v := vec3.T{0, 1, 0}
	material := scene.NewMaterial().E(color.White, environmentEmissionFactor, true).SP(floatimage.LoadOrPanic(filename), &origin, u, v)
	sphere := scene.NewSphere(&origin, environmentRadius, material).N("Environment mapping")

	return sphere
}
