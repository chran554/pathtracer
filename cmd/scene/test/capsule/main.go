package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	anm "pathtracer/internal/pkg/renderfile"
	scn "pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "capsule"

var imageWidth = 800
var imageHeight = 600
var magnification = 0.5

var amountSamples = 128 * 4

var apertureSize = 0.2

var skyDomeEmission = 1.5

var maxRayDepth = 4

func main() {
	skyDome := scn.NewSphere(&vec3.T{0, 0, 0}, 200*100, scn.NewMaterial().
		E(color.White, skyDomeEmission, true).
		SP(floatimage.Load("textures/equirectangular/336_PDM_BG7.jpg"), &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")
	skyDome.RotateY(&vec3.Zero, util.DegToRad(-20))

	lightMaterial := scn.NewMaterial().N("light").E(color.KelvinTemperatureColor2(5000), 10, true)
	light := scn.NewSphere(&vec3.T{-200, 150, -200}, 80, lightMaterial)

	capsule := obj.NewCapsule(100)
	capsule.UpdateVertexNormals(false)
	capsule.RotateY(&vec3.T{}, math.Pi+math.Pi/4+math.Pi/8)

	scene := scn.NewSceneNode().S(skyDome).FS(capsule).S(light)

	cameraOrigin := &vec3.T{0, 15, -40}
	cameraOrigin.Scale(5)
	focusPoint := capsule.Bounds.Center().Add(&vec3.T{0, 0, 0})

	viewVector := focusPoint.Subed(cameraOrigin)
	focusDistance := viewVector.Length()

	camera := scn.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).
		A(apertureSize, nil).
		F(focusDistance).
		D(maxRayDepth)

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, true)
	frame := scn.NewFrame(animation.AnimationName, -1, camera, scene)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := anm.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
