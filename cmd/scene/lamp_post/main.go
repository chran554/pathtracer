package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

const (
	animationName = "lamp_post"

	environmentRadius         = 500.0 * 1000.0
	environmentEmissionFactor = 1.0

	lampPostLightEmission = 3.0
	keroseneLampEmission  = 20.0

	amountFrames = 1

	imageWidth    = 1280
	imageHeight   = 1024
	magnification = 1.0

	amountSamples = 1024 * 8 * 4
	maxRecursion  = 6

	apertureSize = 2.0
)

func main() {
	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, false)

	for frameIndex := range amountFrames {
		//animationProgress := float64(frameIndex) / float64(amountFrames)

		// Sky dome
		environmentSphereOrigin := &vec3.T{0, 0, 0}
		environmentSphereMaterial := scene.NewMaterial().
			E(color.White, environmentEmissionFactor, true).
			SP(
				scene.NewTextureReferenceOrPanic("textures/equirectangular/sunset horizon 2800x1400.jpg", scene.TextureTypeAlbedo),
				environmentSphereOrigin,
				vec3.T{-0.2, 0, -1},
				vec3.T{0, 1, 0},
			)
		environmentSphere := scene.NewSphere(environmentSphereOrigin, environmentRadius, environmentSphereMaterial).N("Environment mapping")

		// Ground
		groundMaterial := scene.NewMaterial().N("Ground material").
			PP(
				scene.NewTextureReferenceOrPanic("textures/ground/soil-cracked.png", scene.TextureTypeAlbedo),
				&vec3.T{0, 0, 0},
				vec3.UnitX.Scaled(150),
				vec3.UnitZ.Scaled(150),
			)
		ground := scene.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, environmentRadius, groundMaterial).N("Ground")

		// Gopher
		gopher := obj.NewGopher(50)
		gopher.RotateY(&vec3.Zero, math.Pi*10.0/10.0)
		gopher.Translate(&vec3.T{75, 0, 100})
		gopherBounds := gopher.Bounds

		// kerosene lamp
		keroseneLamp := obj.NewKeroseneLamp(40, keroseneLampEmission, 1.0)
		keroseneLamp.RotateY(&vec3.Zero, -math.Pi*4.0/10.0)
		keroseneLamp.Translate(&vec3.T{gopherBounds.Center()[0] + gopherBounds.SizeX()/2, 0, gopherBounds.Center()[2] - gopherBounds.SizeY()/2})

		// Lamp post
		lampPost := obj.NewLamppost(200.0, lampPostLightEmission)

		// Camera
		cameraOrigin := gopher.Bounds.Center().Add(&vec3.T{0, 0, -250})
		cameraFocusPoint := gopherBounds.Center().Add(&vec3.T{0, lampPost.Bounds.SizeY() * 0.4, 0})
		camera := scene.NewCamera(cameraOrigin, cameraFocusPoint, amountSamples, magnification).
			A(apertureSize, nil).D(maxRecursion)

		scn := scene.NewSceneNode().
			S(environmentSphere).
			D(ground).
			FS(gopher, lampPost, keroseneLamp)

		fi := frameIndex
		if amountFrames == 1 {
			fi = -1
		}
		frame := scene.NewFrame(animation.AnimationName, fi, camera, scn)

		animation.Frames = append(animation.Frames, frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
