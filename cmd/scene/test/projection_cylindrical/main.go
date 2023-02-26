package main

import (
	"math"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "projection_cylindrical"

var amountAnimationFrames = 1

var imageWidth = 480
var imageHeight = 600
var magnification = 2.0

var amountSamples = 1024 * 12

var apertureSize = 0.5

func main() {
	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true)

	environmentSphere := scn.NewSphere(&vec3.T{0, 0, 0}, 3*100, scn.NewMaterial().
		E(color.White, 1, true).
		//C(color.NewColorGrey(0.2))).
		SP("textures/equirectangular/las-vegas-hotell-lobby.png", &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")

	lamp1 := scn.NewSphere(&vec3.T{-50, 150, -75}, 50, scn.NewMaterial().E(color.NewColorKelvin(5000), 18, true)).N("lamp")

	tableBoard := obj.NewBox(obj.BoxCentered)
	tableBoard.Translate(&vec3.T{0, -tableBoard.Bounds.Ymax, 0})
	tableBoard.Scale(&vec3.Zero, &vec3.T{30, 3, 30})
	tableBoard.Material = scn.NewMaterial().C(color.NewColorGrey(1.0)).M(0.05, 0.8).PP("textures/snooker/cloth02.png", &vec3.T{0, 0, 0}, vec3.T{5, 0, 0}, vec3.T{0, 0, 5})

	sodaCanHeight := 11.6

	sodaCanCocaCola := obj.NewSodaCanCocaCola(sodaCanHeight)
	sodaCanCocaCola.RotateY(&vec3.Zero, math.Pi/2)

	sodaCanPepsi := obj.NewSodaCanPepsi(sodaCanHeight)
	sodaCanPepsi.RotateY(&vec3.Zero, math.Pi*2/3)
	sodaCanPepsi.Translate(&vec3.T{5, 0, 7})

	scene := scn.NewSceneNode().
		S(lamp1, environmentSphere).
		FS(tableBoard, sodaCanCocaCola, sodaCanPepsi)

	//animationStep := 1.0 / float64(amountAnimationFrames)
	for animationFrameIndex := 0; animationFrameIndex < amountAnimationFrames; animationFrameIndex++ {
		// animationProgress := float64(animationFrameIndex) * animationStep

		cameraOrigin := (&vec3.T{0, 15, -15}).Scale(1.3)
		focusPoint := vec3.T{sodaCanHeight / 8, sodaCanHeight * 2 / 3, 0}
		lidCenter := vec3.T{0, sodaCanHeight, 0}
		subed := cameraOrigin.Subed(&lidCenter)
		focusDistance := subed.Length()

		camera := scn.NewCamera(cameraOrigin, &focusPoint, amountSamples, magnification).
			A(apertureSize, "").
			F(focusDistance)

		frame := scn.NewFrame(animation.AnimationName, animationFrameIndex, camera, scene)
		animation.AddFrame(frame)
	}

	anm.WriteAnimationToFile(animation, false)
}
