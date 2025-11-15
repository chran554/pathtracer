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

var animationName = "lamp_post"

var environmentRadius = 500.0 * 1000.0
var environmentEmissionFactor = 1.0

var amountFrames = 1

var imageWidth = 1280
var imageHeight = 1024
var magnification = 1.0

var amountSamples = 512 * 2 * 4 * 2 // * 2 * 2 / 3
var maxRecursion = 8

var apertureSize = 2.0

func main() {
	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, false)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		//animationProgress := float64(frameIndex) / float64(amountFrames)

		// Sky dome
		environmentSphereOrigin := &vec3.T{0, 0, 0}
		environmentSphereMaterial := scn.NewMaterial().
			E(color.White, environmentEmissionFactor, true).
			SP(floatimage.Load("textures/equirectangular/sunset horizon 2800x1400.jpg"), environmentSphereOrigin, vec3.T{-0.2, 0, -1}, vec3.T{0, 1, 0})
		environmentSphere := scn.NewSphere(environmentSphereOrigin, environmentRadius, environmentSphereMaterial).N("Environment mapping")

		// Ground
		groundMaterial := scn.NewMaterial().N("Ground material").
			PP(floatimage.Load("textures/ground/soil-cracked.png"), &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(150), vec3.UnitZ.Scaled(150))
		ground := scn.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, environmentRadius, groundMaterial).N("Ground")

		// Gopher
		gopher := obj.NewGopher(50)
		gopher.RotateY(&vec3.Zero, math.Pi*10.0/10.0)
		gopher.Translate(&vec3.T{75, 0, 100})
		gopher.UpdateBounds()
		gopherBounds := gopher.Bounds

		// kerosene lamp
		keroseneLamp := obj.NewKeroseneLamp(40, 20)
		keroseneLamp.RotateY(&vec3.Zero, -math.Pi*4.0/10.0)
		keroseneLamp.Translate(&vec3.T{gopherBounds.Center()[0] + gopherBounds.SizeX()/2, 0, gopherBounds.Center()[2] - gopherBounds.SizeY()/2})
		keroseneLamp.UpdateBounds()

		// Lamp post
		lampPost := obj.NewLamppost(200.0, 2.0)

		// Camera
		cameraOrigin := gopher.Bounds.Center().Add(&vec3.T{0, 0, -250})
		cameraFocusPoint := gopherBounds.Center().Add(&vec3.T{0, lampPost.Bounds.SizeY() * 0.4, 0})
		camera := scn.NewCamera(cameraOrigin, cameraFocusPoint, amountSamples, magnification).
			A(apertureSize, nil).D(maxRecursion)

		scene := scn.NewSceneNode().
			S(environmentSphere).
			D(ground).
			FS(gopher, lampPost, keroseneLamp)

		frame := scn.NewFrame(animation.AnimationName, frameIndex, camera, scene)

		animation.Frames = append(animation.Frames, frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := anm.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
