package main

import (
	"fmt"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "cornellbox_lucy"

var cornellBoxUnit float64 = 60

var amountSamples = 1024 * 4 // 1024 * 8 // 1024 * 32
var maxRayDepth = 3          // 4      // 6 // max ray recursion depth
var apertureSize = 0.8

var imageWidth = 480
var imageHeight = 480
var magnification = 3.0

var viewPlaneDistance = 1500.0

var lampIntensity = 5.0 * 2.1

func main() {
	cornellBox := obj.NewCornellBox(&vec3.T{cornellBoxUnit, cornellBoxUnit, 3 * cornellBoxUnit}, true, lampIntensity)

	lamp := cornellBox.GetFirstObjectByName("Lamp")
	lamp.Scale(&vec3.Zero, &vec3.T{0.35, 1.0, 1.0})
	lamp.Material.E(color.NewColor(1.0, 0.95, 0.9), lampIntensity, true)

	skyTexture := floatimage.LoadOrPanic("textures/sky/pink clouds.jpg")

	floor := cornellBox.GetFirstObjectByMaterialName("Floor")
	floor.Material = scene.NewMaterial().N("Floor").E(color.White, 1.0, true)
	floor.Material.PP(skyTexture, &vec3.Zero, vec3.T{cornellBoxUnit * 2, 0, 0}, vec3.T{0, 0, cornellBoxUnit})

	roof := cornellBox.GetFirstObjectByMaterialName("Ceiling")
	roof.Material = scene.NewMaterial().N("Ceiling").E(color.White, 1.0, true)
	roof.Material.PP(skyTexture, &vec3.Zero, vec3.T{cornellBoxUnit * 2, 0, 0}, vec3.T{0, 0, cornellBoxUnit})

	backWall := cornellBox.GetFirstObjectByMaterialName("Back")
	backWall.Material = scene.NewMaterial().N("Back").E(color.White, 1.0, true)
	backWall.Material.PP(skyTexture, &vec3.T{-cornellBoxUnit / 2, 0, 0}, vec3.T{cornellBoxUnit * 2, 0, 0}, vec3.T{0, cornellBoxUnit, 0})

	leftWall := cornellBox.GetFirstObjectByMaterialName("Left")
	leftWall.Material = scene.NewMaterial().N("Left").E(color.White, 1.0, true)
	leftWall.Material.PP(skyTexture, &vec3.T{0, 0, -cornellBoxUnit * 3 / 2}, vec3.T{0, 0, cornellBoxUnit * 2}, vec3.T{0, cornellBoxUnit, 0})

	rightWall := cornellBox.GetFirstObjectByMaterialName("Right")
	rightWall.Material = scene.NewMaterial().N("Right").E(color.White, 1.0, true)
	rightWall.Material.PP(skyTexture, &vec3.T{0, 0, -cornellBoxUnit * 3 / 2}, vec3.T{0, 0, cornellBoxUnit * 2}, vec3.T{0, cornellBoxUnit, 0})

	lucy := obj.NewLucy(cornellBoxUnit * 0.8)

	//lucy := obj.NewTessellatedSphere(3, true)
	//lucy.Translate(&vec3.T{0.0, 1.0, 0.0})
	//lucy.Scale(&vec3.Zero, &vec3.T{0.35 * cornellBoxUnit / 2, 0.8 * cornellBoxUnit / 2, 0.35 * cornellBoxUnit / 2})

	lucy.Translate(&vec3.T{0, -cornellBoxUnit * 0.005, 0})
	v := vec3.T{0, cornellBoxUnit / 4, 0}
	u := vec3.T{1, 0, 0}
	lucy.Material = scene.NewMaterial().N("lucy").
		C(color.NewColorGrey(0.90)).
		CP(floatimage.LoadOrPanic("textures/marble/white_marble_double_width.png"), &vec3.Zero, u, v, true)

	scn := scene.NewSceneNode().FS(lucy).FS(cornellBox)

	cameraOrigin := &vec3.T{cornellBox.Bounds.Xmax - 1.0, cornellBox.Bounds.Ymax - 1.0, lucy.Bounds.Center()[2] - 40}
	focusPoint := lucy.Bounds.Center().Added(&vec3.T{3, 0.6 * lucy.Bounds.SizeY() / 2, -0.6 * lucy.Bounds.SizeZ() / 2})
	//cameraOrigin := lucy.Bounds.Center().Added(&vec3.T{0, 0, -40})
	//focusPoint := lucy.Bounds.Center().Added(&vec3.T{3, 0.6 * lucy.Bounds.SizeY() / 2, -0.6 * lucy.Bounds.SizeZ() / 2})

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	camera := scene.NewCamera(cameraOrigin, &focusPoint, amountSamples, magnification).V(viewPlaneDistance).D(maxRayDepth).A(apertureSize, nil)
	frame := scene.NewFrame(animationName, -1, camera, scn)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
