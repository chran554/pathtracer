package main

import (
	"math"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

// 3 soda cans: Test, Coca Cola, and Pepsi
var animationName = "projection_cylindrical"

var amountAnimationFrames = 1

var imageWidth = 1024
var imageHeight = 576
var magnification = 1.0

var amountSamples = 1024 * 16 // 1024 * 12

var apertureSize = 0.2

func main() {
	dx := 1.4
	dy := -10.0

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, false)

	skyDome := scn.NewSphere(&vec3.T{0, 0, 0}, 4*100, scn.NewMaterial().
		E(color.White, 1, true).
		//C(color.NewColorGrey(0.2))).
		SP("textures/equirectangular/las-vegas-hotell-lobby.png", &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")
	skyDome.RotateY(&vec3.Zero, math.Pi+(math.Pi*6/8))

	lamp1 := scn.NewSphere(&vec3.T{-50, 150 + dy, -75}, 60, scn.NewMaterial().E(color.NewColorKelvin(4000), 12, true)).N("lamp")

	tableBoard := obj.NewBox(obj.BoxCentered)
	tableBoard.Translate(&vec3.T{0, -tableBoard.Bounds.Ymax, 0})
	tableBoard.Scale(&vec3.Zero, &vec3.T{30, 3, 20})
	tableBoard.Translate(&vec3.T{10, dy, 0})
	tableBoard.Material = scn.NewMaterial().
		C(color.NewColorGrey(1.0)).
		M(0.15, 0.3).
		PP("textures/wood/darkwood.png", &vec3.T{0, 0, 0}, vec3.T{30, 0, 0}, vec3.T{0, 0, 20})

	sodaCanHeight := 11.6

	sodaCanCocaCola := obj.NewSodaCanCocaColaModern(sodaCanHeight)
	sodaCanCocaCola.RotateY(&vec3.Zero, math.Pi*2/3)
	sodaCanCocaCola.Translate(&vec3.T{-4 * dx, 0 + dy, 0})

	sodaCanPepsi := obj.NewSodaCanPepsi(sodaCanHeight)
	sodaCanPepsi.RotateY(&vec3.Zero, math.Pi*2/3)
	sodaCanPepsi.Translate(&vec3.T{-0.5 * dx, 0 + dy, 9})

	sodaCanMtnDew := obj.NewSodaCanMtnDew(sodaCanHeight)
	sodaCanMtnDew.RotateY(&vec3.Zero, math.Pi*5/6)
	sodaCanMtnDew.Translate(&vec3.T{5 * dx, 0 + dy, 5})

	sodaCanTest := obj.NewSodaCanTest(sodaCanHeight)
	sodaCanTest.RotateY(&vec3.Zero, -math.Pi*2/3)
	sodaCanTest.Translate(&vec3.T{6 * dx, 0 + dy, -1.5})

	scene := scn.NewSceneNode().
		S(lamp1, skyDome).
		FS(tableBoard, sodaCanCocaCola, sodaCanPepsi, sodaCanMtnDew, sodaCanTest)

	//animationStep := 1.0 / float64(amountAnimationFrames)
	for animationFrameIndex := 0; animationFrameIndex < amountAnimationFrames; animationFrameIndex++ {
		// animationProgress := float64(animationFrameIndex) * animationStep

		cameraOrigin := (&vec3.T{0, 9, -15}).Scale(1.7).Add(&vec3.T{0, dy, 0})
		focusPoint := vec3.T{0, sodaCanHeight*0.5 + dy, sodaCanHeight / 8}
		lidCenter := vec3.T{0, sodaCanHeight + dy, 0}
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
