package main

import (
	"fmt"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	anm "pathtracer/internal/pkg/renderfile"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var amountSamples = 1024

func main() {
	cylinderRadius := 3 * 100.0 // 300 cm

	animation := scn.NewAnimation("projection_cylindrical2", 800, 600, 1.0, false, false)

	projectionOrigin := &vec3.T{0, 0, 0}
	projectionU := vec3.UnitZ.Inverted()
	projectionV := vec3.UnitY.Scaled(cylinderRadius * 1.5)

	cylinder := obj.NewCylinder(obj.CylinderYPositive, cylinderRadius, cylinderRadius*1.5)
	cylinder.Material = scn.NewMaterial().
		E(color.White, 1.0, true).
		CP(floatimage.Load("textures/tapeter 2/CaptainsCabin_Image_Flatshot_Item_8887_360.jpg"), projectionOrigin, projectionU, projectionV, false)

	// Ground
	groundMaterial := scn.NewMaterial().N("Ground material").
		PP(floatimage.Load("textures/ground/soil-cracked.png"), &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(200), vec3.UnitZ.Scaled(200))
	ground := scn.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, cylinderRadius*3, groundMaterial).N("Ground")

	scene := scn.NewSceneNode().D(ground).FS(cylinder)

	cameraOrigin := vec3.T{0, cylinderRadius * 2, -cylinderRadius * 3}
	focusPoint := vec3.T{0, cylinderRadius / 2, 0}
	camera := scn.NewCamera(&cameraOrigin, &focusPoint, amountSamples, 1.0)

	frame := scn.NewFrame(animation.AnimationName, -1, camera, scene)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := anm.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
