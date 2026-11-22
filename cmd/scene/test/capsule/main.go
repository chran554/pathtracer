package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

const (
	animationName = "capsule"

	imageWidth    = 800
	imageHeight   = 600
	magnification = 0.5

	amountSamples = 128 * 4

	apertureSize = 0.2

	skyDomeEmission = 1.5

	maxRayDepth = 4
)

func main() {
	skyDome := scene.NewSphere(&vec3.T{0, 0, 0}, 200*100, scene.NewMaterial().
		E(color.White, skyDomeEmission, true).
		SP(
			scene.NewTextureReferenceOrPanic("textures/equirectangular/336_PDM_BG7.jpg", scene.TextureTypeAlbedo),
			&vec3.T{0, 0, 0},
			vec3.T{1, 0, 0},
			vec3.T{0, 1, 0},
		)).N("sky dome")
	skyDome.RotateY(&vec3.Zero, util.DegToRad(-20))

	lightMaterial := scene.NewMaterial().N("light").E(color.KelvinTemperatureColor2(5000), 10, true)
	light := scene.NewSphere(&vec3.T{-200, 150, -200}, 80, lightMaterial)

	capsule := obj.NewCapsule(100)
	capsule.UpdateVertexNormals(false)
	capsule.RotateY(&vec3.T{}, math.Pi+math.Pi/4+math.Pi/8)

	scn := scene.NewSceneNode().S(skyDome).FS(capsule).S(light)

	cameraOrigin := &vec3.T{0, 15, -40}
	cameraOrigin.Scale(5)
	focusPoint := capsule.Bounds.Center().Add(&vec3.T{0, 0, 0})

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
