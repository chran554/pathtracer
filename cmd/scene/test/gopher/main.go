package main

import (
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
)

var animationName = "gopher"

var amountFrames = 1

var amountSamples = 1024 * 4 * 3

var imageWidth = 400
var imageHeight = 500
var magnification = 1.0

func main() {
	// Sky
	groundMaterial := scn.NewMaterial().PP("textures/floor/Calacatta-Vena-French-Pattern-Architextures.jpg", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(150), vec3.UnitZ.Scaled(150))
	ground := &scn.Disc{Name: "ground", Origin: &vec3.T{0, 0, 0}, Normal: &vec3.UnitY, Radius: 5000.0, Material: groundMaterial}

	// Sky
	skyMaterial := scn.NewMaterial().
		E(color.White, 0.5, true).
		SP("textures/equirectangular/wirebox 6192x3098.png", &vec3.T{0, 0, 0}, vec3.UnitX, vec3.UnitY)
	skyDome := scn.NewSphere(&vec3.T{0, 0, 0}, 5000, skyMaterial).N("sky dome")

	// Gopher
	gopher := obj.NewGopher(200.0)
	gopher.Translate(&vec3.T{0, -gopher.Bounds.Ymin, 0})
	gopher.ScaleUniform(&vec3.Zero, 2.0)
	gopher.RotateY(&vec3.Zero, math.Pi*5.0/6.0)
	gopher.Translate(&vec3.T{0, 0, 0})
	gopher.UpdateBounds()

	gopherLightMaterial := scn.NewMaterial().E(color.NewColor(6.0, 5.3, 4.5), 20, true)
	gopherLight := scn.NewSphere(&vec3.T{-150, 250, -175}, 15.0, gopherLightMaterial).N("Gopher light")

	scene := scn.NewSceneNode().
		S(gopherLight, skyDome).
		D(ground).
		FS(gopher)

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		camera := getCamera(animationProgress)
		frame := scn.NewFrame(animationName, -1, camera, scene)
		animation.AddFrame(frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func getCamera(animationProgress float64) *scn.Camera {
	cameraOrigin := &vec3.T{0, 200, -800}
	focusPoint := &vec3.T{0, 150, 0}

	// Animation
	angle := (math.Pi / 2.0) * animationProgress
	scn.RotateY(cameraOrigin, &vec3.Zero, angle)
	scn.RotateY(focusPoint, &vec3.Zero, angle)

	heading := focusPoint.Subed(cameraOrigin)
	focusDistance := heading.Length() * 1.75

	return scn.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).F(focusDistance)
}
