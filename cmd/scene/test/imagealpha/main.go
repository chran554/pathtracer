package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	anm "pathtracer/internal/pkg/renderfile"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "imagealpha"

var amountFrames = 1

var amountSamples = 1024 * 6 * 2

var imageWidth = 800
var imageHeight = 600
var magnification = 1.0

func main() {
	// Object
	object := obj.NewPokemonTangela(200.0)
	object.RotateY(&vec3.Zero, math.Pi*7.0/8.0)

	lightMaterial1 := scn.NewMaterial().E(color.KelvinTemperatureColor2(5500), 40, true)
	lightMaterial2 := scn.NewMaterial().E(color.KelvinTemperatureColor2(5500), 2, true)
	light1 := scn.NewSphere(&vec3.T{-150, 250, -175}, 45.0, lightMaterial1).N("light")
	light2 := scn.NewSphere(&vec3.T{150 * 2, 250, -175 * 2}, 45.0, lightMaterial2).N("light")

	scene := scn.NewSceneNode().
		S(light1, light2).
		FS(object)

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		camera := getCamera(animationProgress)
		frame := scn.NewFrame(animationName, -1, camera, scene)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := anm.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func getCamera(animationProgress float64) *scn.Camera {
	cameraOrigin := &vec3.T{0, 200, -400}
	focusPoint := &vec3.T{0, 100, 0}

	// AnimationInformation
	angle := (math.Pi / 2.0) * animationProgress
	scn.RotateY(cameraOrigin, &vec3.Zero, angle)
	scn.RotateY(focusPoint, &vec3.Zero, angle)

	heading := focusPoint.Subed(cameraOrigin)
	focusDistance := heading.Length() - 150.0

	return scn.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).F(focusDistance)
}
