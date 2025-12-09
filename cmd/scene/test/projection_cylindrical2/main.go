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

var amountSamples = 1024

func main() {
	cylinderRadius := 3 * 100.0 // 300 cm

	animation := scene.NewAnimation("projection_cylindrical2", 800, 600, 1.0, false, false)

	projectionOrigin := &vec3.T{0, 0, 0}
	projectionU := vec3.UnitZ.Inverted()
	projectionV := vec3.UnitY.Scaled(cylinderRadius * 1.5)

	cylinder := obj.NewCylinder(obj.CylinderYPositive, cylinderRadius, cylinderRadius*1.5)
	cylinder.Material = scene.NewMaterial().
		E(color.White, 1.0, true).
		CP(floatimage.LoadOrPanic("textures/tapeter 2/CaptainsCabin_Image_Flatshot_Item_8887_360.jpg"), projectionOrigin, projectionU, projectionV, false)

	// Ground
	groundMaterial := scene.NewMaterial().N("Ground material").
		PP(floatimage.LoadOrPanic("textures/ground/soil-cracked.png"), &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(200), vec3.UnitZ.Scaled(200))
	ground := scene.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, cylinderRadius*3, groundMaterial).N("Ground")

	scn := scene.NewSceneNode().D(ground).FS(cylinder)

	cameraOrigin := vec3.T{0, cylinderRadius * 2, -cylinderRadius * 3}
	focusPoint := vec3.T{0, cylinderRadius / 2, 0}
	camera := scene.NewCamera(&cameraOrigin, &focusPoint, amountSamples, 1.0)

	frame := scene.NewFrame(animation.AnimationName, -1, camera, scn)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
