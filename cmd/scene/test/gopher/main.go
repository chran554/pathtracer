package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "gopher"

var amountFrames = 1

var amountSamples = 1024 * 4 * 3

var imageWidth = 400
var imageHeight = 500
var magnification = 1.0

func main() {
	// Sky
	groundMaterial := scene.NewMaterial().PP(floatimage.LoadOrPanic("textures/floor/Calacatta-Vena-French-Pattern-Architextures.jpg"), &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(150), vec3.UnitZ.Scaled(150))
	ground := &scene.Disc{Name: "ground", Origin: &vec3.T{0, 0, 0}, Normal: &vec3.UnitY, Radius: 5000.0, Material: groundMaterial}

	// Sky
	skyMaterial := scene.NewMaterial().
		E(color.White, 0.5, true).
		SP(floatimage.LoadOrPanic("textures/equirectangular/wirebox 6192x3098.png"), &vec3.T{0, 0, 0}, vec3.UnitX, vec3.UnitY)
	skyDome := scene.NewSphere(&vec3.T{0, 0, 0}, 5000, skyMaterial).N("sky dome")

	// Gopher
	gopher := obj.NewGopher(200.0)
	gopher.Translate(&vec3.T{0, -gopher.Bounds.Ymin, 0})
	gopher.ScaleUniform(&vec3.Zero, 2.0)
	gopher.RotateY(&vec3.Zero, math.Pi*5.0/6.0)
	gopher.Translate(&vec3.T{0, 0, 0})
	gopher.UpdateBounds()

	gopherLightMaterial := scene.NewMaterial().E(color.NewColor(6.0, 5.3, 4.5), 20, true)
	gopherLight := scene.NewSphere(&vec3.T{-150, 250, -175}, 15.0, gopherLightMaterial).N("Gopher light")

	scn := scene.NewSceneNode().
		S(gopherLight, skyDome).
		D(ground).
		FS(gopher)

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		camera := getCamera(animationProgress)
		frame := scene.NewFrame(animationName, -1, camera, scn)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func getCamera(animationProgress float64) *scene.Camera {
	cameraOrigin := &vec3.T{0, 200, -800}
	focusPoint := &vec3.T{0, 150, 0}

	// Animation
	angle := (math.Pi / 2.0) * animationProgress
	scene.RotateY(cameraOrigin, &vec3.Zero, angle)
	scene.RotateY(focusPoint, &vec3.Zero, angle)

	heading := focusPoint.Subed(cameraOrigin)
	focusDistance := heading.Length() * 1.75

	return scene.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).F(focusDistance)
}
