package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	anm "pathtracer/internal/pkg/renderfile"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "pokemon_tangela"

var amountFrames = 1

var amountSamples = 1024 * 3 * 3

var imageWidth = 800
var imageHeight = 600
var magnification = 1.0

func main() {
	// Ground
	groundMaterial := scn.NewMaterial().PP(floatimage.Load("textures/floor/Calacatta-Vena-French-Pattern-Architextures.jpg"), &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(150), vec3.UnitZ.Scaled(150))
	ground := &scn.Disc{Name: "ground", Origin: &vec3.T{0, 0, 0}, Normal: &vec3.UnitY, Radius: 5000.0, Material: groundMaterial}

	// Sky
	skyMaterial := scn.NewMaterial().
		E(color.White, 0.5, true).
		SP(floatimage.Load("textures/equirectangular/wirebox 6192x3098.png"), &vec3.T{0, 0, 0}, vec3.UnitX, vec3.UnitY)
	skyDome := scn.NewSphere(&vec3.T{0, 0, 0}, 5000, skyMaterial).N("sky dome")

	// Object
	object := obj.NewPokemonTangela(200.0)
	object.RotateY(&vec3.Zero, math.Pi*7.0/8.0)

	lightMaterial := scn.NewMaterial().E(color.KelvinTemperatureColor2(5500), 40, true)
	light := scn.NewSphere(&vec3.T{-150, 250, -175}, 45.0, lightMaterial).N("light")

	scene := scn.NewSceneNode().
		S(light, skyDome).
		D(ground).
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
