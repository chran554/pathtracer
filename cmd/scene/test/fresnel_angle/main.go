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

var amountSamples = 1024 * 12

func main() {
	cylinderRadius := 3 * 100.0 // 300 cm

	animation := scene.NewAnimation("fresnel_angle", 800, 600, 1.0, false, false)

	projectionOrigin := &vec3.T{0, 0, 0}
	projectionU := vec3.UnitZ.Inverted()
	projectionV := vec3.UnitY.Scaled(cylinderRadius * 1.5 * 2)

	cylinder := obj.NewCylinder(obj.CylinderYPositive, cylinderRadius, cylinderRadius/2.2)
	cylinder.Material = scene.NewMaterial().
		E(color.White, 4.0, true).
		CP(floatimage.Load("textures/floor/checkered.jpg"), projectionOrigin, projectionU, projectionV, false)

	// Ground
	// groundMaterial := scene.NewMaterial().N("Ground material").
	// 	PP("textures/ground/soil-cracked.png", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(200), vec3.UnitZ.Scaled(200))
	//ground := scene.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, cylinderRadius, scene.NewMaterial()).N("Ground")

	sphereRadius := 0.25 * cylinderRadius
	sphereCenter := vec3.T{0, sphereRadius * 1.5, 0}
	sphereMaterial := scene.NewMaterial().T(0.0, true, scene.RefractionIndex_Porcelain).M(0.0, 0.0)
	sphere := scene.NewSphere(&sphereCenter, sphereRadius, sphereMaterial)

	skyDome := scene.NewSphere(&vec3.T{0, 0, 0}, 10*100, scene.NewMaterial().
		E(color.White, 1, true).
		//C(color.NewColorGrey(0.2))).
		SP(floatimage.Load("textures/equirectangular/leaf_trees_by_lake.jpg"), &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")

	scn := scene.NewSceneNode().
		// D(ground).
		FS(cylinder).
		S(sphere, skyDome)

	cameraOrigin := vec3.T{0, sphereCenter[1] + cylinderRadius/2, -cylinderRadius + 10}
	focusPoint := sphereCenter
	camera := scene.NewCamera(&cameraOrigin, &focusPoint, amountSamples, 1.0)

	frame := scene.NewFrame(animation.AnimationName, -1, camera, scn)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
