package main

import (
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
)

var animationName = "pokemon_tangela"

var amountFrames = 1

var amountSamples = 1024 * 3 * 3

var imageWidth = 800
var imageHeight = 600
var magnification = 1.0

func main() {
	width := int(float64(imageWidth) * magnification)
	height := int(float64(imageHeight) * magnification)

	// Keep image proportions to an even amount of pixel for mp4 encoding
	if width%2 == 1 {
		width++
	}
	if height%2 == 1 {
		height++
	}

	// Ground
	groundMaterial := scn.NewMaterial().PP("textures/floor/Calacatta-Vena-French-Pattern-Architextures.jpg", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(150), vec3.UnitZ.Scaled(150))
	ground := &scn.Disc{Name: "ground", Origin: &vec3.T{0, 0, 0}, Normal: &vec3.UnitY, Radius: 5000.0, Material: groundMaterial}

	// Sky
	skyMaterial := scn.NewMaterial().
		E(color.White, 0.5, true).
		SP("textures/equirectangular/wirebox 6192x3098.png", &vec3.T{0, 0, 0}, vec3.UnitX, vec3.UnitY)
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

	animation := scn.NewAnimation(animationName, width, height, magnification, false, false)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		camera := getCamera(animationProgress)
		frame := scn.NewFrame(animationName, -1, camera, scene)
		animation.AddFrame(frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func getCamera(animationProgress float64) *scn.Camera {
	cameraOrigin := &vec3.T{0, 200, -400}
	focusPoint := &vec3.T{0, 100, 0}

	// Animation
	angle := (math.Pi / 2.0) * animationProgress
	scn.RotateY(cameraOrigin, &vec3.Zero, angle)
	scn.RotateY(focusPoint, &vec3.Zero, angle)

	heading := focusPoint.Subed(cameraOrigin)
	focusDistance := heading.Length() - 150.0

	return scn.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).F(focusDistance)
}
