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

var animationName = "kerosene_lamp"

var imageWidth = 600
var imageHeight = 800
var magnification = 1.0

var amountSamples = 1024 * 32
var maxRayDepth = 10
var apertureSize = 0.2

var keroseneLampEmission = 15.0
var sootness = 1.0
var skyDomeEmission = 0.9

func main() {
	skyDome := scene.NewSphere(&vec3.T{0, 0, 0}, 200*100, scene.NewMaterial().
		E(color.White, skyDomeEmission, true).
		SP(floatimage.LoadOrPanic("textures/equirectangular/331_PDM_BG1.jpg"), &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")
	skyDome.RotateY(&vec3.Zero, util.DegToRad(-20))

	tableBoard := obj.NewBox(obj.BoxCentered)
	tableBoard.Scale(&vec3.Zero, &vec3.T{110 / 2, 3 / 2, 150 / 2})
	tableBoard.Translate(&vec3.T{0, -tableBoard.Bounds.Ymax + 80, 0})
	tableBoard.Material = scene.NewMaterial().
		C(color.NewColorGrey(1.0)).
		PP(floatimage.LoadOrPanic("textures/wallpaper/Blossom2_Image_Tile_Item_9471w.jpg"), &vec3.T{0, 0, 0}, vec3.T{60, 0, 0}, vec3.T{0, 0, 40})

	keroseneLamp := obj.NewKeroseneLamp(40, keroseneLampEmission, sootness)
	keroseneLamp.RotateY(&vec3.Zero, util.DegToRad(-90))
	keroseneLamp.Translate(&vec3.T{20, tableBoard.Bounds.Ymax, -20})

	scn := scene.NewSceneNode().S(skyDome).FS(tableBoard).FS(keroseneLamp)

	cameraOrigin := &vec3.T{-20, 110, -50}
	focusPoint := keroseneLamp.Bounds.Center().Add(&vec3.T{0, 0, 0})

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
