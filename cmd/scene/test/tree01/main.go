package main

import (
	"fmt"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "tree01"

var imageWidth = 800
var imageHeight = 1200
var magnification = 1.0

var amountSamples = 1024

var apertureSize = 0.2

var skyDomeEmission = 1.5

var maxRayDepth = 10

func main() {
	skyDome := scene.NewSphere(&vec3.T{0, 0, 0}, 200*100, scene.NewMaterial().
		E(color.White, skyDomeEmission, true).
		SP(floatimage.LoadOrPanic("textures/equirectangular/336_PDM_BG7.jpg"), &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")
	skyDome.RotateY(&vec3.Zero, util.DegToRad(-20))

	lightMaterial := scene.NewMaterial().N("light").E(color.KelvinTemperatureColor2(5000), 10, true)
	light := scene.NewSphere(&vec3.T{-200, 150, -200}, 80, lightMaterial)

	tree := obj.NewTree01(300)

	scn := scene.NewSceneNode().S(skyDome).FS(tree).S(light)

	cameraOrigin := &vec3.T{-20, 15, -40}
	cameraOrigin.Scale(5)
	focusPoint := tree.Bounds.Center().Add(&vec3.T{0, 0, 0})

	viewVector := focusPoint.Subed(cameraOrigin)
	focusDistance := viewVector.Length()

	camera := scene.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).
		A(apertureSize, nil).
		F(focusDistance).
		D(maxRayDepth)

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, true)
	frame := scene.NewFrame(animation.AnimationName, -1, camera, scn)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
